package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	documentpb "github.com/yunomu/kansousen/proto/document"
)

const (
	kifuAttr      = "kifu"
	stepAttr      = "step"
	versionAttr   = "version"
	seqAttr       = "seq"
	userIdAttr    = "userId"
	kifuIdAttr    = "kifuId"
	createdTsAttr = "createdTs"
	sfenAttr      = "sfen"
	posAttr       = "pos"

	BatchUnit = 25
)

type DynamoDBKifuRecord struct {
	UserId    string `dynamodbav:"userId,omitempty"`
	KifuId    string `dynamodbav:"kifuId"`
	CreatedTs int64  `dynamodbav:"createdTs,omitempty"`
	StartTs   int64  `dynamodbav:"startTs,omitempty"`
	Sfen      string `dynamodbav:"sfen,omitempty"`
	Seq       int32  `dynamodbav:"seq"`
	Pos       string `dynamodbav:"pos,omitempty"`
	Kifu      []byte `dynamodbav:"kifu,omitempty"`
	Step      []byte `dynamodbav:"step,omitempty"`
	Version   int64  `dynamodbav:"version,omitempty"`
}

type DynamoDB struct {
	client    *dynamodb.DynamoDB
	tableName string

	parallelism int
}

var _ DB = (*DynamoDB)(nil)

type DynamoDBOption func(*DynamoDB)

func SetParallelism(i int) DynamoDBOption {
	return func(db *DynamoDB) {
		db.parallelism = i
	}
}

func NewDynamoDB(client *dynamodb.DynamoDB, tableName string, ops ...DynamoDBOption) *DynamoDB {
	db := &DynamoDB{
		client:    client,
		tableName: tableName,

		parallelism: 2,
	}
	for _, f := range ops {
		f(db)
	}

	return db
}

func (db *DynamoDB) PutKifu(ctx context.Context,
	kifu *documentpb.Kifu,
	steps []*documentpb.Step,
	version int64,
) error {
	bs, err := proto.Marshal(kifu)
	if err != nil {
		return err
	}
	newVersion := time.Now().UnixNano()
	kifuAv, err := dynamodbattribute.MarshalMap(DynamoDBKifuRecord{
		UserId:    kifu.GetUserId(),
		KifuId:    kifu.GetKifuId(),
		CreatedTs: kifu.GetCreatedTs(),
		StartTs:   kifu.GetStartTs(),
		Sfen:      kifu.GetSfen(),
		Seq:       0,
		Kifu:      bs,
		Version:   newVersion,
	})
	if err != nil {
		return err
	}

	out, err := db.client.TransactWriteItemsWithContext(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []*dynamodb.TransactWriteItem{
			{
				Put: &dynamodb.Put{
					ConditionExpression: aws.String("attribute_not_exists(#version) OR #version = :version"),
					ExpressionAttributeNames: map[string]*string{
						"#version": aws.String(versionAttr),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":version": &dynamodb.AttributeValue{
							N: aws.String(fmt.Sprintf("%d", version)),
						},
					},
					Item: kifuAv,
				},
			},
		},
	})
	if err != nil {
		return err
	}
	var _ = out

	g, ctx := errgroup.WithContext(ctx)

	reqsCh := make(chan []*dynamodb.WriteRequest, db.parallelism)
	g.Go(func() error {
		defer close(reqsCh)

		var reqs []*dynamodb.WriteRequest
		for _, step := range steps {
			bs, err := proto.Marshal(kifu)
			if err != nil {
				return err
			}
			av, err := dynamodbattribute.MarshalMap(DynamoDBKifuRecord{
				UserId: step.GetUserId(),
				KifuId: step.GetKifuId(),
				Seq:    step.GetSeq(),
				Step:   bs,
			})

			reqs = append(reqs, &dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{Item: av},
			})

			if len(reqs) == BatchUnit {
				select {
				case reqsCh <- reqs:
					reqs = nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		select {
		case reqsCh <- reqs:
		case <-ctx.Done():
			return ctx.Err()
		}

		return nil
	})

	for i := 0; i < db.parallelism; i++ {
		g.Go(func() error {
			for reqs := range reqsCh {
				out, err := db.client.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]*dynamodb.WriteRequest{
						db.tableName: reqs,
					},
				})
				if err == nil {
					return err
				}

				var _ = out
			}

			return nil
		})
	}

	return g.Wait()
}

func (db *DynamoDB) GetKifu(
	ctx context.Context,
	kifuId string,
) (*documentpb.Kifu, int64, error) {
	key, err := dynamodbattribute.MarshalMap(DynamoDBKifuRecord{
		KifuId: kifuId,
		Seq:    0,
	})
	if err != nil {
		return nil, 0, err
	}
	out, err := db.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName:            aws.String(db.tableName),
		Key:                  key,
		ProjectionExpression: aws.String(strings.Join([]string{kifuAttr, versionAttr}, ",")),
	})
	if err != nil {
		return nil, 0, err
	}

	record := DynamoDBKifuRecord{}
	if err := dynamodbattribute.UnmarshalMap(out.Item, &record); err != nil {
		return nil, 0, err
	}

	var kifu documentpb.Kifu
	if err := proto.Unmarshal(record.Kifu, &kifu); err != nil {
		return nil, 0, &ErrInvalidValue{
			Details: err.Error(),
		}
	}

	return &kifu, record.Version, nil
}

type StepSlice []*documentpb.Step

func (s StepSlice) Len() int               { return len(s) }
func (s StepSlice) Less(i int, j int) bool { return s[i].GetSeq() < s[j].GetSeq() }
func (s StepSlice) Swap(i int, j int)      { s[i], s[j] = s[j], s[i] }

func (db *DynamoDB) GetKifuAndSteps(
	ctx context.Context,
	kifuId string,
) (*documentpb.Kifu, []*documentpb.Step, int64, error) {
	g, ctx := errgroup.WithContext(ctx)

	itemsCh := make(chan []map[string]*dynamodb.AttributeValue, db.parallelism)
	g.Go(func() error {
		defer close(itemsCh)

		var rerr error
		if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(db.tableName),
			KeyConditionExpression: aws.String("#kifuId = :kifuId AND #seq = :seq"),
			ExpressionAttributeNames: map[string]*string{
				"#kifuId": aws.String(kifuIdAttr),
				"#seq":    aws.String(seqAttr),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":kifuId": &dynamodb.AttributeValue{S: aws.String(kifuId)},
				":seq":    &dynamodb.AttributeValue{N: aws.String("0")},
			},
			ProjectionExpression: aws.String(strings.Join([]string{kifuAttr, versionAttr, stepAttr}, ",")),
		}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			case itemsCh <- out.Items:
				return true
			}
		}); err != nil {
			return err
		}

		return rerr
	})

	var kifu *documentpb.Kifu
	var version int64
	stepsCh := make(chan []*documentpb.Step, db.parallelism)
	for i := 0; i < db.parallelism; i++ {
		g.Go(func() error {
			for items := range itemsCh {
				recs := []DynamoDBKifuRecord{}
				if err := dynamodbattribute.UnmarshalListOfMaps(items, &recs); err != nil {
					return err
				}

				var steps []*documentpb.Step
				for _, rec := range recs {
					switch rec.Seq {
					case 0: // Kifu
						var k documentpb.Kifu
						if err := proto.Unmarshal(rec.Kifu, &k); err != nil {
							return &ErrInvalidValue{
								Details: err.Error(),
							}
						}

						kifu = &k
						version = rec.Version
					default: // Step
						var s documentpb.Step
						if err := proto.Unmarshal(rec.Step, &s); err != nil {
							return &ErrInvalidValue{
								Details: err.Error(),
							}
						}

						steps = append(steps, &s)
					}
				}

				select {
				case stepsCh <- steps:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	go func() {
		g.Wait()
		close(stepsCh)
	}()

	var steps []*documentpb.Step
	for ss := range stepsCh {
		steps = append(steps, ss...)
	}

	if err := g.Wait(); err != nil {
		return nil, nil, 0, err
	}

	sort.Sort(StepSlice(steps))

	return kifu, steps, version, nil
}

type versionedKifu struct {
	kifu    *documentpb.Kifu
	version int64
}

func (db *DynamoDB) ListKifu(ctx context.Context, userId string, f func(kifu *documentpb.Kifu, version int64)) error {
	g, ctx := errgroup.WithContext(ctx)

	itemsCh := make(chan []map[string]*dynamodb.AttributeValue, db.parallelism)
	g.Go(func() error {
		defer close(itemsCh)

		var rerr error
		if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(db.tableName),
			IndexName:              aws.String("User"),
			KeyConditionExpression: aws.String("#userId = :userId AND #seq = :seq"),
			ExpressionAttributeNames: map[string]*string{
				"#userId": aws.String(userIdAttr),
				"#seq":    aws.String(seqAttr),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":userId": &dynamodb.AttributeValue{S: aws.String(userId)},
				":seq":    &dynamodb.AttributeValue{N: aws.String("0")},
			},
			ProjectionExpression: aws.String(strings.Join([]string{kifuAttr, versionAttr}, ",")),
		}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			case itemsCh <- out.Items:
				return true
			}
		}); err != nil {
			return err
		}

		return rerr
	})

	ch := make(chan *versionedKifu, db.parallelism)
	for i := 0; i < db.parallelism; i++ {
		g.Go(func() error {
			for items := range itemsCh {
				recs := []DynamoDBKifuRecord{}
				if err := dynamodbattribute.UnmarshalListOfMaps(items, &recs); err != nil {
					return err
				}

				for _, rec := range recs {
					var kifu documentpb.Kifu
					if err := proto.Unmarshal(rec.Kifu, &kifu); err != nil {
						return &ErrInvalidValue{
							Details: err.Error(),
						}
					}

					select {
					case ch <- &versionedKifu{kifu: &kifu, version: rec.Version}:
					case <-ctx.Done():
					}
				}
			}

			return nil
		})
	}

	go func() {
		g.Wait()
		close(ch)
	}()

	for vk := range ch {
		f(vk.kifu, vk.version)
	}

	return g.Wait()
}

func (db *DynamoDB) GetKifuIdsBySfen(ctx context.Context, sfen string) ([]*UserKifu, error) {
	var ret []*UserKifu
	var rerr error
	if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("Sfen"),
		KeyConditionExpression: aws.String("#sfen = :sfen"),
		ExpressionAttributeNames: map[string]*string{
			"#sfen": aws.String(sfenAttr),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sfen": &dynamodb.AttributeValue{S: aws.String(sfen)},
		},
		ProjectionExpression: aws.String(strings.Join([]string{kifuIdAttr, userIdAttr}, ",")),
	}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
		select {
		case <-ctx.Done():
			rerr = ctx.Err()
			return false
		default:
		}

		var records []DynamoDBKifuRecord
		if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &records); err != nil {
			rerr = err
			return false
		}

		for _, r := range records {
			ret = append(ret, &UserKifu{
				UserId: r.UserId,
				KifuId: r.KifuId,
			})
		}

		return true
	}); err != nil {
		return nil, err
	} else if rerr != nil {
		return nil, rerr
	}

	return ret, nil
}

type stepKey struct {
	kifuId string
	userId string
	seq    int32
}

func (db *DynamoDB) GetSamePositions(ctx context.Context, userIds []string, pos string, options ...GetSamePositionsOption) ([]*Position, error) {
	opts := &getSamePositionsOptions{
		numStep: 5,
	}
	for _, f := range options {
		f(opts)
	}

	g, ctx := errgroup.WithContext(ctx)

	stepKeyCh := make(chan *stepKey, db.parallelism)
	g.Go(func() error {
		defer close(stepKeyCh)

		var rerr error
		if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(db.tableName),
			IndexName:              aws.String("Position"),
			KeyConditionExpression: aws.String("#pos = :pos"),
			ExpressionAttributeNames: map[string]*string{
				"#pos": aws.String(posAttr),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":pos": &dynamodb.AttributeValue{S: aws.String(pos)},
			},
			ProjectionExpression: aws.String(strings.Join([]string{kifuIdAttr, seqAttr, userIdAttr}, ",")),
		}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			default:
			}

			var records []DynamoDBKifuRecord
			if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &records); err != nil {
				rerr = err
				return false
			}

			for _, r := range records {
				select {
				case stepKeyCh <- &stepKey{
					kifuId: r.KifuId,
					userId: r.UserId,
					seq:    r.Seq,
				}:
				case <-ctx.Done():
					rerr = ctx.Err()
					return false
				}
			}

			return true
		}); err != nil {
			return err
		}

		return rerr
	})

	posCh := make(chan *Position, db.parallelism)
	for i := 0; i < db.parallelism; i++ {
		g.Go(func() error {
			for stepKey := range stepKeyCh {
				var steps []*documentpb.Step
				var rerr error
				if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
					TableName:              aws.String(db.tableName),
					KeyConditionExpression: aws.String("#kifuId = :kifuId AND #seq >= :seqStart AND #seq < :seqEnd"),
					ExpressionAttributeNames: map[string]*string{
						"#kifuId": aws.String(kifuIdAttr),
						"#seq":    aws.String(seqAttr),
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":kifuId":   &dynamodb.AttributeValue{S: aws.String(stepKey.kifuId)},
						":seqStart": &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", stepKey.seq))},
						":seqEnd":   &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", int32(stepKey.seq)+opts.numStep))},
					},
					ProjectionExpression: aws.String(stepAttr),
				}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
					select {
					case <-ctx.Done():
						rerr = ctx.Err()
						return false
					default:
					}

					var records []DynamoDBKifuRecord
					if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &records); err != nil {
						rerr = err
						return false
					}
					for _, r := range records {
						var step documentpb.Step
						if err := proto.Unmarshal(r.Step, &step); err != nil {
							rerr = err
							return false
						}

						steps = append(steps, &step)
					}

					return true
				}); err != nil {
					return err
				} else if rerr != nil {
					return rerr
				}

				select {
				case posCh <- &Position{
					KifuId: stepKey.kifuId,
					UserId: stepKey.userId,
					Steps:  steps,
				}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			return nil
		})
	}

	go func() {
		g.Wait()
		close(posCh)
	}()

	var ret []*Position
	for pos := range posCh {
		ret = append(ret, pos)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (db *DynamoDB) GetRecentKifu(ctx context.Context, userId string, limit int) ([]*documentpb.Kifu, error) {
	var ret []*documentpb.Kifu
	var rerr error
	if err := db.client.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("Created"),
		KeyConditionExpression: aws.String("#userId = :userId"),
		ExpressionAttributeNames: map[string]*string{
			"#userId": aws.String(userIdAttr),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userId": &dynamodb.AttributeValue{S: aws.String(userId)},
		},
		ProjectionExpression: aws.String(kifuAttr),
		ScanIndexForward:     aws.Bool(false),
		Limit:                aws.Int64(int64(limit)),
	}, func(out *dynamodb.QueryOutput, lastPage bool) bool {
		select {
		case <-ctx.Done():
			rerr = ctx.Err()
			return false
		default:
		}

		var records []DynamoDBKifuRecord
		if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &records); err != nil {
			rerr = err
			return false
		}
		for _, rec := range records {
			var kifu documentpb.Kifu
			if err := proto.Unmarshal(rec.Kifu, &kifu); err != nil {
				rerr = err
				return false
			}

			ret = append(ret, &kifu)
		}

		return true
	}); err != nil {
		return nil, err
	} else if rerr != nil {
		return nil, rerr
	}

	return ret, nil
}

func (db *DynamoDB) DeleteKifu(ctx context.Context, kifuId string) error {
	return fmt.Errorf("not implemented")
}
