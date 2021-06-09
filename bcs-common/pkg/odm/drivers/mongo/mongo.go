/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// Options options for mongo db
type Options struct {
	AuthMechanism         string
	Username              string
	Password              string
	AuthDatabase          string
	Database              string
	ConnectTimeoutSeconds int
	MaxPoolSize           uint64
	MinPoolSize           uint64
	Hosts                 []string
}

// DB mongodb
type DB struct {
	dbName string
	mCli   *mongo.Client
}

// NewDB create db
func NewDB(opt *Options) (*DB, error) {
	credential := mopt.Credential{
		AuthMechanism: opt.AuthMechanism,
		AuthSource:    opt.AuthDatabase,
		Username:      opt.Username,
		Password:      opt.Password,
		PasswordSet:   true,
	}
	if len(credential.AuthMechanism) == 0 {
		credential.AuthMechanism = mongoAuthMichanismSha256
	}
	// construct mongo client options
	mCliOpt := &mopt.ClientOptions{
		Auth:  &credential,
		Hosts: opt.Hosts,
	}
	if opt.MaxPoolSize != 0 {
		mCliOpt.MaxPoolSize = &opt.MaxPoolSize
	}
	if opt.MinPoolSize != 0 {
		mCliOpt.MinPoolSize = &opt.MinPoolSize
	}
	var timeoutDuration time.Duration
	if opt.ConnectTimeoutSeconds != 0 {
		timeoutDuration = time.Duration(opt.ConnectTimeoutSeconds) * time.Second
	}
	mCliOpt.ConnectTimeout = &timeoutDuration

	// create mongo client
	mCli, err := mongo.NewClient(mCliOpt)
	if err != nil {
		return nil, err
	}
	// connect to mongo
	if err := mCli.Connect(context.TODO()); err != nil {
		return nil, err
	}

	return &DB{
		dbName: opt.Database,
		mCli:   mCli,
	}, nil
}

// DataBase get database
func (db *DB) DataBase() string {
	return db.dbName
}

// Close close db connection
func (db *DB) Close() error {
	return db.mCli.Disconnect(context.TODO())
}

// Ping ping database
func (db *DB) Ping() error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("ping", err, startTime)
	}()
	err = db.mCli.Ping(context.TODO(), nil)
	return err
}

// HasTable if table exists
func (db *DB) HasTable(ctx context.Context, tableName string) (bool, error) {
	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("hasTable", err, startTime)
	}()
	cursor, err = db.mCli.Database(db.dbName).ListCollections(ctx, bson.M{
		"name": tableName,
		"type": "collection",
	})
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		return true, nil
	}
	return false, nil
}

// ListTableNames list collection names
func (db *DB) ListTableNames(ctx context.Context) ([]string, error) {
	var retList []string
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("listTableNames", err, startTime)
	}()
	retList, err = db.mCli.Database(db.dbName).ListCollectionNames(ctx, bson.M{})
	return retList, err
}

// CreateTable create collection
func (db *DB) CreateTable(ctx context.Context, tableName string) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("createTable", err, startTime)
	}()
	err = db.mCli.Database(db.dbName).RunCommand(ctx, map[string]interface{}{
		"create": tableName,
	}).Err()
	return err
}

// DropTable drop table
func (db *DB) DropTable(ctx context.Context, tableName string) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("dropTable", err, startTime)
	}()
	err = db.mCli.Database(db.dbName).Collection(tableName).Drop(ctx)
	return err
}

// Table get collection object
func (db *DB) Table(tableName string) drivers.Table {
	return &Collection{
		collectionName: tableName,
		DB:             db,
	}
}

// Collection collection for mongodb
type Collection struct {
	collectionName string
	*DB
}

// CreateIndex create index for collection
func (c *Collection) CreateIndex(ctx context.Context, idx drivers.Index) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("createIndex", err, startTime)
	}()
	if len(idx.Name) == 0 {
		err = fmt.Errorf("index name cannot be empty")
		return err
	}
	indexOpt := mopt.Index()
	indexOpt.SetUnique(idx.Unique)
	indexOpt.SetName(idx.Name)
	indexOpt.SetBackground(idx.Background)
	indexModel := mongo.IndexModel{
		Keys:    idx.Key,
		Options: indexOpt,
	}

	_, err = c.mCli.Database(c.dbName).Collection(c.collectionName).Indexes().CreateOne(ctx, indexModel)
	return err
}

// DropIndex drop index for collection
func (c *Collection) DropIndex(ctx context.Context, indexName string) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("dropIndex", err, startTime)
	}()
	_, err = c.mCli.Database(c.dbName).Collection(c.collectionName).Indexes().DropOne(ctx, indexName)
	return err
}

// HasIndex if has index with certain name
func (c *Collection) HasIndex(ctx context.Context, indexName string) (bool, error) {
	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("hasIndex", err, startTime)
	}()
	cursor, err = c.mCli.Database(c.dbName).Collection(c.collectionName).Indexes().List(ctx)
	if err != nil {
		return false, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		tmpIndex := &drivers.Index{}
		cursor.Decode(tmpIndex)
		if tmpIndex.Name == indexName {
			return true, nil
		}
	}
	return false, nil
}

// Indexes list indexes of collection
func (c *Collection) Indexes(ctx context.Context) ([]drivers.Index, error) {
	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("indexes", err, startTime)
	}()
	cursor, err = c.mCli.Database(c.dbName).Collection(c.collectionName).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var idxArr []drivers.Index
	for cursor.Next(ctx) {
		tmpIndex := drivers.Index{}
		err := cursor.Decode(&tmpIndex)
		if err != nil {
			return nil, err
		}
		idxArr = append(idxArr, tmpIndex)
	}
	return idxArr, nil
}

// Find return finder
func (c *Collection) Find(condition *operator.Condition) drivers.Find {
	return &Finder{
		Collection: c,
		condition:  condition,
	}
}

// Aggregation do aggregation
func (c *Collection) Aggregation(ctx context.Context, pipeline interface{}, result interface{}) error {
	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("aggregation", err, startTime)
	}()
	cursor, err = c.mCli.Database(c.dbName).
		Collection(c.collectionName).
		Aggregate(ctx, pipeline, &mopt.AggregateOptions{})
	if err != nil {
		return err
	}
	defer func() {
		cursor.Close(ctx)
	}()
	return cursor.All(ctx, result)
}

// Insert insert many data
func (c *Collection) Insert(ctx context.Context, docs []interface{}) (int, error) {
	var ret *mongo.InsertManyResult
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("insert", err, startTime)
	}()
	ret, err = c.mCli.Database(c.dbName).Collection(c.collectionName).InsertMany(ctx, docs)
	if err != nil {
		if strings.Contains(err.Error(), "E11000 duplicate key") {
			return len(ret.InsertedIDs), drivers.ErrTableRecordDuplicateKey
		}
		return len(ret.InsertedIDs), err
	}
	return len(ret.InsertedIDs), nil
}

// Update update data by condition
func (c *Collection) Update(ctx context.Context, condition *operator.Condition, data interface{}) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("insert", err, startTime)
	}()
	// convert condition to filter
	filter := condition.Combine(leafNodeProcessor, branchNodeProcessor)
	_, err = c.mCli.Database(c.dbName).Collection(c.collectionName).UpdateOne(ctx, filter, data)
	return err
}

// UpdateMany update many data by condition
func (c *Collection) UpdateMany(ctx context.Context, condition *operator.Condition, data interface{}) (int64, error) {
	var err error
	var ret *mongo.UpdateResult
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("updateMany", err, startTime)
	}()
	// convert condition to filter
	filter := condition.Combine(leafNodeProcessor, branchNodeProcessor)
	ret, err = c.mCli.Database(c.dbName).Collection(c.collectionName).UpdateMany(ctx, filter, data)
	if err != nil {
		return 0, err
	}
	return ret.ModifiedCount, nil
}

// Upsert update or insert data by condition
func (c *Collection) Upsert(ctx context.Context, condition *operator.Condition, data interface{}) error {
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("upsert", err, startTime)
	}()
	// convert condition to filter
	filter := condition.Combine(leafNodeProcessor, branchNodeProcessor)
	upsertFlag := true
	updateOpt := &mopt.UpdateOptions{
		Upsert: &upsertFlag,
	}
	_, err = c.mCli.Database(c.dbName).Collection(c.collectionName).UpdateOne(ctx, filter, data, updateOpt)
	return err
}

// Delete delete data
func (c *Collection) Delete(ctx context.Context, condition *operator.Condition) (int64, error) {
	var ret *mongo.DeleteResult
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("delete", err, startTime)
	}()
	// convert condition to filter
	filter := condition.Combine(leafNodeProcessor, branchNodeProcessor)
	ret, err = c.mCli.Database(c.dbName).Collection(c.collectionName).DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}
	return ret.DeletedCount, nil
}

// Watch watch data
func (c *Collection) Watch(conditions []*operator.Condition) drivers.Watch {
	return &Watcher{
		Collection: c,
		conditions: conditions,
	}
}

// Finder do mongodb search
type Finder struct {
	sort       map[string]interface{}
	projection map[string]int
	start      int64
	limit      int64
	condition  *operator.Condition
	*Collection
}

// WithProjection set returned fields
func (f *Finder) WithProjection(projection map[string]int) drivers.Find {
	f.projection = projection
	return f
}

// WithSort set sort order
func (f *Finder) WithSort(sort map[string]interface{}) drivers.Find {
	f.sort = sort
	return f
}

// WithStart set start offset
func (f *Finder) WithStart(start int64) drivers.Find {
	f.start = start
	return f
}

// WithLimit set limit of result
func (f *Finder) WithLimit(limit int64) drivers.Find {
	f.limit = limit
	return f
}

// One find one data by find option
func (f *Finder) One(ctx context.Context, result interface{}) error {
	findOpts := &mopt.FindOptions{}
	if len(f.projection) != 0 {
		findOpts.Projection = f.projection
	}
	if f.start != 0 {
		findOpts.SetSkip(f.start)
	}
	findOpts.SetLimit(1)
	if len(f.sort) != 0 {
		findOpts.SetSort(f.sort)
	}

	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("findOne", err, startTime)
	}()

	// convert condition to filter
	filter := f.condition.Combine(leafNodeProcessor, branchNodeProcessor)
	cursor, err = f.mCli.Database(f.dbName).Collection(f.collectionName).Find(ctx, filter, findOpts)
	if err != nil {
		return err
	}
	defer func() {
		cursor.Close(ctx)
	}()
	for cursor.Next(ctx) {
		return cursor.Decode(result)
	}
	return drivers.ErrTableRecordNotFound
}

// All find all data by find option
func (f *Finder) All(ctx context.Context, result interface{}) error {
	findOpts := &mopt.FindOptions{}
	if len(f.projection) != 0 {
		findOpts.Projection = f.projection
	}
	if f.start != 0 {
		findOpts.SetSkip(f.start)
	}
	if f.limit != 0 {
		findOpts.SetLimit(f.limit)
	}
	if len(f.sort) != 0 {
		findOpts.SetSort(f.sort)
	}

	var err error
	var cursor *mongo.Cursor
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("findAll", err, startTime)
	}()
	// convert condition to filter
	filter := f.condition.Combine(leafNodeProcessor, branchNodeProcessor)
	cursor, err = f.mCli.Database(f.dbName).Collection(f.collectionName).Find(ctx, filter, findOpts)
	if err != nil {
		return err
	}
	return cursor.All(ctx, result)
}

// Count count data, only condition takes effective
func (f *Finder) Count(ctx context.Context) (int64, error) {
	// convert condition to filter
	var counter int64
	var err error
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("count", err, startTime)
	}()
	filter := f.condition.Combine(leafNodeProcessor, branchNodeProcessor)
	counter, err = f.mCli.Database(f.dbName).Collection(f.collectionName).CountDocuments(ctx, filter)
	return counter, err
}

// Watcher wrap mongodb change stream
type Watcher struct {
	projection       map[string]int
	batchSize        int32
	isFull           bool
	maxAwaitDuration time.Duration
	startTimestamp   *primitive.Timestamp
	conditions       []*operator.Condition
	*Collection
}

// WithBatchSize set the maximum number of documents to be included in each batch returned by the server
func (w *Watcher) WithBatchSize(batch int32) drivers.Watch {
	w.batchSize = batch
	return w
}

// WithFullContent set if watch action returned the full document
func (w *Watcher) WithFullContent(isFull bool) drivers.Watch {
	w.isFull = isFull
	return w
}

// WithMaxAwaitTime set the maximum amount of time
func (w *Watcher) WithMaxAwaitTime(duration time.Duration) drivers.Watch {
	w.maxAwaitDuration = duration
	return w
}

// WithStartTimestamp set operation time that watch start
func (w *Watcher) WithStartTimestamp(timeSec uint32, index uint32) drivers.Watch {
	w.startTimestamp = &primitive.Timestamp{
		T: timeSec,
		I: index,
	}
	return w
}

// DoWatch do watch action
func (w *Watcher) DoWatch(ctx context.Context) (chan *drivers.WatchEvent, error) {
	changeStreamOpt := &mopt.ChangeStreamOptions{}
	if w.batchSize > 0 {
		changeStreamOpt.BatchSize = &w.batchSize
	}
	if w.startTimestamp != nil {
		changeStreamOpt.StartAtOperationTime = w.startTimestamp
	}
	if w.maxAwaitDuration > 0 {
		changeStreamOpt.MaxAwaitTime = &w.maxAwaitDuration
	}
	var fullDoc mopt.FullDocument
	if w.isFull {
		fullDoc = mopt.UpdateLookup
	} else {
		fullDoc = mopt.Default
	}
	changeStreamOpt.FullDocument = &fullDoc

	filters := make([]interface{}, 0)
	for _, condition := range w.conditions {
		filter := condition.Combine(leafNodeProcessor, branchNodeProcessor)
		filters = append(filters, filter)
	}

	var err error
	var changeStream *mongo.ChangeStream
	startTime := time.Now()
	defer func() {
		reportMongdbMetrics("watch", err, startTime)
	}()
	changeStream, err = w.mCli.Database(w.dbName).Collection(w.collectionName).Watch(ctx, filters, changeStreamOpt)
	if err != nil {
		return nil, err
	}

	eventChannel := make(chan *drivers.WatchEvent, 100)
	go func() {
		defer changeStream.Close(ctx)
		errEvent := &drivers.WatchEvent{
			Type:           drivers.EventError,
			DBName:         w.dbName,
			CollectionName: w.collectionName,
		}
		for changeStream.Next(ctx) {
			var data Event
			if err := changeStream.Decode(&data); err != nil {
				blog.Errorf("decode data failed, err %s", err.Error())
				eventChannel <- errEvent
				return
			}

			newEvent, err := convertMongoEvent(data)
			if err != nil {
				blog.Errorf("convert mongo event failed, err %s", err.Error())
				eventChannel <- errEvent
				return
			}
			eventChannel <- newEvent
		}
	}()

	return eventChannel, nil
}
