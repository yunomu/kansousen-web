package db

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"google.golang.org/protobuf/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"

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
	ckKifuIdSeq := fmt.Sprintf("%s#0", kifu.GetKifuId())
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
					Item: map[string]*dynamodb.AttributeValue{
						userIdAttr:    &dynamodb.AttributeValue{S: aws.String(kifu.GetUserId())},
						kifuIdAttr:    &dynamodb.AttributeValue{S: aws.String(kifu.GetKifuId())},
						"createdTs":   &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", kifu.GetCreatedTs()))},
						"startTs":     &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", kifu.GetStartTs()))},
						"sfen":        &dynamodb.AttributeValue{S: aws.String(kifu.GetSfen())},
						seqAttr:       &dynamodb.AttributeValue{N: aws.String("0")},
						"ckKifuIdSeq": &dynamodb.AttributeValue{S: aws.String(ckKifuIdSeq)},
						kifuAttr:      &dynamodb.AttributeValue{B: bs},
						versionAttr:   &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", newVersion))},
					},
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
			ckKifuIdSeq := fmt.Sprintf("%s#%d", step.GetKifuId(), step.GetSeq())

			reqs = append(reqs, &dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: map[string]*dynamodb.AttributeValue{
						userIdAttr:    &dynamodb.AttributeValue{S: aws.String(step.GetUserId())},
						kifuIdAttr:    &dynamodb.AttributeValue{S: aws.String(step.GetKifuId())},
						seqAttr:       &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", step.GetSeq()))},
						"ckKifuIdSeq": &dynamodb.AttributeValue{S: aws.String(ckKifuIdSeq)},
						stepAttr:      &dynamodb.AttributeValue{B: bs},
					},
				},
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

func kifuFromItem(item map[string]*dynamodb.AttributeValue) (*documentpb.Kifu, int64, error) {
	attr, ok := item[kifuAttr]
	if !ok {
		return nil, 0, &ErrInvalidValue{
			Details: "kifu field not found",
		}
	} else if attr.B == nil {
		return nil, 0, &ErrInvalidValue{
			Details: "kifu field is not bytes",
		}
	}

	attrVer, ok := item[versionAttr]
	if !ok {
		return nil, 0, &ErrInvalidValue{
			Details: "version field not found",
		}
	} else if attrVer.N == nil {
		return nil, 0, &ErrInvalidValue{
			Details: "version field is not string",
		}
	}

	var kifu documentpb.Kifu
	if err := proto.Unmarshal(attr.B, &kifu); err != nil {
		return nil, 0, &ErrInvalidValue{
			Details: err.Error(),
		}
	}

	ver, err := strconv.ParseInt(aws.StringValue(attrVer.N), 10, 64)
	if err != nil {
		return nil, 0, &ErrInvalidValue{
			Details: err.Error(),
		}
	}

	return &kifu, ver, nil
}

func stepFromItem(item map[string]*dynamodb.AttributeValue) (*documentpb.Step, error) {
	attr, ok := item[stepAttr]
	if !ok {
		return nil, &ErrInvalidValue{
			Details: "step field not found",
		}
	} else if attr.B == nil {
		return nil, &ErrInvalidValue{
			Details: "step field is not number",
		}
	}

	var step documentpb.Step
	if err := proto.Unmarshal(attr.B, &step); err != nil {
		return nil, &ErrInvalidValue{
			Details: err.Error(),
		}
	}

	return &step, nil
}

func stringFromItem(name string, item map[string]*dynamodb.AttributeValue) (string, error) {
	attr, ok := item[name]
	if !ok {
		return "", &ErrInvalidValue{
			Details: fmt.Sprintf("%s field not found", name),
		}
	} else if attr.S == nil {
		return "", &ErrInvalidValue{
			Details: fmt.Sprintf("%s field is not string", name),
		}
	}

	return aws.StringValue(attr.S), nil
}

func intFromItem(name string, item map[string]*dynamodb.AttributeValue) (int64, error) {
	attr, ok := item[name]
	if !ok {
		return 0, &ErrInvalidValue{
			Details: fmt.Sprintf("%s field not found", name),
		}
	} else if attr.N == nil {
		return 0, &ErrInvalidValue{
			Details: fmt.Sprintf("%s field is not number", name),
		}
	}

	i, err := strconv.ParseInt(aws.StringValue(attr.N), 10, 64)
	if err != nil {
		return 0, &ErrInvalidValue{
			Details: err.Error(),
		}
	}

	return i, nil
}

func (db *DynamoDB) GetKifu(
	ctx context.Context,
	kifuId string,
) (*documentpb.Kifu, int64, error) {
	out, err := db.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			kifuIdAttr: &dynamodb.AttributeValue{S: aws.String(kifuId)},
			seqAttr:    &dynamodb.AttributeValue{N: aws.String("0")},
		},
		ProjectionExpression: aws.String(strings.Join([]string{kifuAttr, versionAttr}, ",")),
	})
	if err != nil {
		return nil, 0, err
	}

	return kifuFromItem(out.Item)
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
				var steps []*documentpb.Step
				for _, item := range items {
					seq, err := intFromItem(seqAttr, item)
					if err != nil {
						return err
					}

					switch seq {
					case 0: // Kifu
						k, v, err := kifuFromItem(item)
						if err != nil {
							return err
						}
						kifu = k
						version = v
					default: // Step
						s, err := stepFromItem(item)
						if err != nil {
							return err
						}
						steps = append(steps, s)
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
				for _, item := range items {
					kifu, version, err := kifuFromItem(item)
					if err != nil {
						return err
					}

					select {
					case ch <- &versionedKifu{kifu: kifu, version: version}:
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

		for _, item := range out.Items {
			kifuId, err := stringFromItem(kifuIdAttr, item)
			if err != nil {
				rerr = err
				return false
			}

			userId, err := stringFromItem(userIdAttr, item)
			if err != nil {
				rerr = err
				return false
			}

			ret = append(ret, &UserKifu{
				UserId: userId,
				KifuId: kifuId,
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
	seq    int64
}

func strInclude(ss []string, t string) bool {
	for _, s := range ss {
		if s == t {
			return true
		}
	}
	return false
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

			for _, item := range out.Items {
				kifuId, err := stringFromItem(kifuIdAttr, item)
				if err != nil {
					rerr = err
					return false
				}

				if strInclude(opts.excludeKifuIds, kifuId) {
					return true
				}

				userId, err := stringFromItem(userIdAttr, item)
				if err != nil {
					rerr = err
					return false
				}

				seq, err := intFromItem(seqAttr, item)
				if err != nil {
					rerr = err
					return false
				}

				select {
				case stepKeyCh <- &stepKey{
					kifuId: kifuId,
					userId: userId,
					seq:    seq,
				}:
				case <-ctx.Done():
					rerr = err
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

					for _, item := range out.Items {
						step, err := stepFromItem(item)
						if err != nil {
							rerr = err
							return false
						}

						steps = append(steps, step)
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

		for _, item := range out.Items {
			kifu, _, err := kifuFromItem(item)
			if err != nil {
				rerr = err
				return false
			}

			ret = append(ret, kifu)
		}

		return true
	}); err != nil {
		return nil, err
	} else if rerr != nil {
		return nil, rerr
	}

	return ret, nil
}

func (db *DynamoDB) DeleteKifu(ctx context.Context, userId, kifuId string) error {
	// TODO
	return fmt.Errorf("not implemented")
}
