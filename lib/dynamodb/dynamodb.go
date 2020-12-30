package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	pkField              = "PK"
	skField              = "SK"
	valueField           = "Bytes"
	versionField         = "Version"
	inversedVersionField = "InversedVersion"
	inversedVersionIndex = "InversedVersionIndex"
)

var (
	ErrOptimisticLocking = errors.New("optimistic locking error")
)

type DynamoDB struct {
	client *dynamodb.DynamoDB
	table  string

	logger Logger
}

type DynamoDBTableOption func(*DynamoDB)

func SetLogger(l Logger) DynamoDBTableOption {
	if l == nil {
		l = &nopLogger{}
	}

	return func(d *DynamoDB) {
		d.logger = l
	}
}

func NewDynamoDBTable(client *dynamodb.DynamoDB, table string, opts ...DynamoDBTableOption) *DynamoDB {
	db := &DynamoDB{
		client: client,
		table:  table,

		logger: &nopLogger{},
	}

	for _, f := range opts {
		f(db)
	}

	return db
}

func (d *DynamoDB) Init(ctx context.Context) error {
	return d.client.WaitUntilTableExists(&dynamodb.DescribeTableInput{
		TableName: aws.String(d.table),
	})
}

func toDynamoValue(pk, sk string, value []byte, version int64) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		pkField: {
			S: aws.String(pk),
		},
		skField: {
			S: aws.String(sk),
		},
		valueField: {
			B: value,
		},
		versionField: {
			N: aws.String(strconv.FormatInt(version, 10)),
		},
		inversedVersionField: {
			N: aws.String(strconv.FormatInt(-version, 10)),
		},
	}
}

func (d *DynamoDB) Put(ctx context.Context, pk, sk string, value []byte, version int64) (int64, error) {
	newVersion := time.Now().UnixNano()

	out, err := d.client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.table),

		ConditionExpression: aws.String(
			fmt.Sprintf("attribute_not_exists(%s) OR (%s = :version)", skField, versionField),
		),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":version": &dynamodb.AttributeValue{
				N: aws.String(strconv.FormatInt(version, 10)),
			},
		},

		Item: toDynamoValue(pk, sk, value, newVersion),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return 0, ErrOptimisticLocking
			default:
				// do nothing
			}
		}
		return 0, err
	}

	d.logger.Info("PutItem", out.GoString())

	return newVersion, nil
}

type WriteItem struct {
	PK, SK  string
	Bytes   []byte
	Version int64
}

const BatchWriteUnit = 25

func (d *DynamoDB) BatchPut(ctx context.Context, items []*WriteItem) (int64, error) {
	newVersion := time.Now().UnixNano()

	var reqs []*dynamodb.WriteRequest
	for _, item := range items {
		reqs = append(reqs, &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: toDynamoValue(item.PK, item.SK, item.Bytes, newVersion),
			},
		})
	}

	_, err := d.client.BatchWriteItemWithContext(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			d.table: reqs,
		},
	})
	if err != nil {
		return 0, err
	}

	return newVersion, nil
}

type Item struct {
	Bytes   []byte
	Version int64
}

func dynamodbItemValue(item map[string]*dynamodb.AttributeValue) (*Item, error) {
	val, ok := item[valueField]
	if !ok {
		return nil, fmt.Errorf("invalid object: %s is not found", valueField)
	}

	ts, ok := item[versionField]
	if !ok {
		return nil, fmt.Errorf("invalid object: %s is not found", versionField)
	}
	version, err := strconv.ParseInt(aws.StringValue(ts.N), 10, 64)
	if err != nil {
		return nil, err
	}

	return &Item{Bytes: val.B, Version: version}, nil
}

func dynamodbItemKey(item map[string]*dynamodb.AttributeValue) (*Key, error) {
	pk, ok := item[pkField]
	if !ok {
		return nil, fmt.Errorf("invalid object: %s is not found", pkField)
	}

	sk, ok := item[skField]
	if !ok {
		return nil, fmt.Errorf("invalid object: %s is not found", skField)
	}

	return &Key{PK: aws.StringValue(pk.S), SK: aws.StringValue(sk.S)}, nil
}

func requestItem(key *Key) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		pkField: {
			S: aws.String(key.PK),
		},
		skField: {
			S: aws.String(key.SK),
		},
	}
}

func (d *DynamoDB) Get(ctx context.Context, pk, sk string) (*Item, error) {
	out, err := d.client.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key:       requestItem(&Key{PK: pk, SK: sk}),
	})
	if err != nil {
		return nil, err
	}

	d.logger.Info("GetItem", out.GoString())

	item, err := dynamodbItemValue(out.Item)

	return item, err
}

type Key struct {
	PK, SK string
}

func (d *DynamoDB) BatchGet(ctx context.Context, keys []*Key) ([]*Item, error) {
	var requestItems []map[string]*dynamodb.AttributeValue
	for _, key := range keys {
		requestItems = append(requestItems, requestItem(key))
	}

	var ret []*Item
	var rerr error
	if err := d.client.BatchGetItemPagesWithContext(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			d.table: &dynamodb.KeysAndAttributes{
				Keys: requestItems,
			},
		},
	}, func(out *dynamodb.BatchGetItemOutput, lastPage bool) bool {
		res, ok := out.Responses[d.table]
		if !ok {
			return false
		}

		for _, item := range res {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			default:
			}

			i, err := dynamodbItemValue(item)
			if err != nil {
				rerr = err
				return false
			}

			ret = append(ret, i)
		}

		return true
	}); err != nil {
		return nil, err
	}
	if rerr != nil {
		return nil, rerr
	}

	return ret, nil
}

type queryOption struct {
	expression string
	values     map[string]*dynamodb.AttributeValue
	limit      int64
}

type QueryOption func(*queryOption)

func SetQueryRange(start, end string) QueryOption {
	return func(op *queryOption) {
		op.expression = fmt.Sprintf("%s BETWEEN :start AND :end", skField)
		op.values = map[string]*dynamodb.AttributeValue{
			":start": &dynamodb.AttributeValue{
				S: aws.String(start),
			},
			":end": &dynamodb.AttributeValue{
				S: aws.String(end),
			},
		}
	}
}

func SetQueryRangeStart(start string) QueryOption {
	return func(op *queryOption) {
		op.expression = fmt.Sprintf("%s >= :sk", skField)
		op.values = map[string]*dynamodb.AttributeValue{
			":sk": &dynamodb.AttributeValue{
				S: aws.String(start),
			},
		}
	}
}

func SetQueryLimit(limit int64) QueryOption {
	return func(op *queryOption) {
		op.limit = limit
	}
}

func (d *DynamoDB) Query(ctx context.Context, pk string, f func(*Item), ops ...QueryOption) error {
	o := &queryOption{}
	for _, f := range ops {
		f(o)
	}

	expression := fmt.Sprintf("%s = :pk", pkField)
	if o.expression != "" {
		expression += " AND " + o.expression
	}

	values := map[string]*dynamodb.AttributeValue{
		":pk": &dynamodb.AttributeValue{S: aws.String(pk)},
	}
	for k, v := range o.values {
		values[k] = v
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(d.table),
		KeyConditionExpression:    aws.String(expression),
		ExpressionAttributeValues: values,
	}

	//if o.limit > 0 {
	//	input.Limit = aws.Int64(o.limit)
	//}

	var rerr error
	if err := d.client.QueryPagesWithContext(ctx, input, func(out *dynamodb.QueryOutput, lastPage bool) bool {
		for _, item := range out.Items {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			default:
			}

			i, err := dynamodbItemValue(item)
			if err != nil {
				rerr = err
				return false
			}

			f(i)
		}

		return true
	}); err != nil {
		return err
	}

	return rerr
}

func (d *DynamoDB) RecentlyUpdated(ctx context.Context, pk string, limit int) ([]*Key, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(d.table),
		IndexName:              aws.String(inversedVersionIndex),
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :pk", pkField)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: aws.String(pk)},
		},
		Limit: aws.Int64(int64(limit)),
	}
	d.logger.Info("RecentlyUpdated request", input.GoString())
	var ret []*Key
	var rerr error
	if err := d.client.QueryPagesWithContext(ctx, input, func(out *dynamodb.QueryOutput, lastPage bool) bool {
		for _, item := range out.Items {
			select {
			case <-ctx.Done():
				rerr = ctx.Err()
				return false
			default:
			}

			key, err := dynamodbItemKey(item)
			if err != nil {
				rerr = err
				return false
			}

			ret = append(ret, key)
		}

		return true
	}); err != nil {
		if rerr != nil {
			return nil, rerr
		}
		return nil, err
	}

	return ret, nil
}

func (d *DynamoDB) Delete(ctx context.Context, keys []*Key) error {
	if len(keys) == 0 {
		return nil
	}

	var items []*dynamodb.WriteRequest
	for _, key := range keys {
		items = append(items, &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: requestItem(key),
			},
		})
	}

	_, err := d.client.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			d.table: items,
		},
	})

	return err
}
