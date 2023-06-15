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

// Package mongo xxx
package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

const (
	objectIndexName = "bcsrawobject"
)

// Store client for mongo
type Store struct {
	mDriver  drivers.DB
	colCache mapset.Set
}

// NewMongoStore create mongo store
func NewMongoStore(mDriver drivers.DB) *Store {
	return &Store{
		mDriver:  mDriver,
		colCache: mapset.NewSet(),
	}
}

// ensureCollection ensure collection
func (s *Store) ensureCollection(ctx context.Context, obj *types.RawObject) error {
	// Check if the collection is already in the cache, if so, return nil
	colName := obj.GetObjectType()
	if s.colCache.Contains(colName) {
		return nil
	}

	// Check if the table exists in the database, if not, create it
	hasTable, err := s.mDriver.HasTable(ctx, string(colName))
	if err != nil {
		return err
	}
	if !hasTable {
		tErr := s.mDriver.CreateTable(ctx, string(colName))
		if tErr != nil {
			return tErr
		}
	}

	// Check if the index exists in the table, if not, create it
	hasIndex, err := s.mDriver.Table(string(colName)).HasIndex(ctx, objectIndexName)
	if err != nil {
		return err
	}
	if !hasIndex {
		iErr := s.mDriver.Table(string(colName)).CreateIndex(ctx, drivers.Index{
			Name:   objectIndexName,
			Unique: true,
			Key: bson.D{
				bson.E{Key: types.TagResourceType, Value: 1},
				bson.E{Key: types.TagResourceName, Value: 1},
				bson.E{Key: types.TagNamespace, Value: 1},
				bson.E{Key: types.TagClusterID, Value: 1},
			},
		})
		if iErr != nil {
			return iErr
		}
	}

	// Add the collection to the cache
	s.colCache.Add(colName)
	return nil
}

// Get get object
func (s *Store) Get(ctx context.Context, resourceType types.ObjectType, key types.ObjectKey, opt *store.GetOptions) (
	*types.RawObject, error) {

	// Check if the resourceType is empty, if so, return an error
	if len(resourceType) == 0 {
		return nil, fmt.Errorf("object type cannot be empty")
	}

	// Create a new RawObject
	rawObj := &types.RawObject{}

	// Create a keyM map with the resource name, namespace, and cluster ID
	keyM := operator.M{
		types.TagResourceName: key.Name,
		types.TagNamespace:    key.Namespace,
		types.TagClusterID:    key.ClusterID,
	}

	// Find the RawObject in the database using the resourceType and keyM
	err := s.mDriver.Table(string(resourceType)).
		Find(operator.NewLeafCondition(operator.Eq, keyM)).One(ctx, rawObj)
	if err != nil {
		blog.Errorf("find one by object key %+v failed, err %s", key, err.Error())
		return nil, fmt.Errorf("find one by object key %+v failed, err %s", key, err.Error())
	}

	// Return the RawObject and nil error
	return rawObj, nil
}

// Create create object
func (s *Store) Create(ctx context.Context, obj *types.RawObject, opt *store.CreateOptions) error {
	// Check if the obj or object type is empty, if so, return an error
	if obj == nil || len(obj.GetObjectType()) == 0 {
		return fmt.Errorf("object or object type cannot be empty")
	}

	// Check if the create options are empty, if so, return an error
	if opt == nil {
		return fmt.Errorf("create options cannot be empty")
	}

	// Ensure that the collection exists in the database
	if err := s.ensureCollection(ctx, obj); err != nil {
		return err
	}

	// Create a new RawObject and set found to true
	rawObj := types.RawObject{}
	found := true

	// Find the RawObject in the database using the object name, namespace, and cluster ID
	if err := s.mDriver.Table(string(obj.GetObjectType())).
		Find(operator.NewLeafCondition(operator.Eq, operator.M{
			types.TagResourceName: obj.GetName(),
			types.TagNamespace:    obj.GetNamespace(),
			types.TagClusterID:    obj.GetClusterID(),
		})).One(ctx, &rawObj); err != nil {
		if !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("search failed %s/%s/%s when create, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
			return fmt.Errorf("search failed %s/%s/%s when create, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
		}
		found = false
	}

	// If the object already exists and update exists is false, return an error
	if !opt.UpdateExists && found {
		blog.Errorf("object %s/%s/%s to create already exists",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
		return fmt.Errorf("object %s/%s/%s to create already exists",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
	}

	var err error
	if found {
		obj.SetCreateTime(rawObj.GetCreateTime())
		obj.SetUpdateTime(time.Now())
		err = s.mDriver.Table(string(obj.GetObjectType())).
			Update(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
				types.TagResourceName: obj.GetName(),
				types.TagNamespace:    obj.GetNamespace(),
				types.TagClusterID:    obj.GetClusterID(),
			}), operator.M{"$set": obj})
	} else {
		obj.SetCreateTime(time.Now())
		obj.SetUpdateTime(time.Now())
		_, err = s.mDriver.Table(string(obj.GetObjectType())).Insert(ctx, []interface{}{obj})
	}
	if err != nil {
		blog.Errorf("create object %s failed, err %s", obj.ToString(), err.Error())
		return fmt.Errorf("create object %s failed, err %s", obj.ToString(), err.Error())
	}
	return nil
}

// Update update object
func (s *Store) Update(ctx context.Context, obj *types.RawObject, opt *store.UpdateOptions) error {
	// Check if the obj or object type is empty, if so, return an error
	if obj == nil || len(obj.GetObjectType()) == 0 {
		return fmt.Errorf("object or object type cannot be empty")
	}

	// Check if the update options are empty, if so, return an error
	if opt == nil {
		return fmt.Errorf("update options cannot be empty")
	}

	// Ensure that the collection exists in the database
	if err := s.ensureCollection(ctx, obj); err != nil {
		return err
	}

	rawObj := types.RawObject{}
	found := true
	// Find the RawObject in the database using the object name, namespace, and cluster ID
	if err := s.mDriver.Table(string(obj.GetObjectType())).
		Find(operator.NewLeafCondition(operator.Eq, operator.M{
			types.TagResourceName: obj.GetName(),
			types.TagNamespace:    obj.GetNamespace(),
			types.TagClusterID:    obj.GetClusterID(),
		})).One(ctx, &rawObj); err != nil {
		if !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("search failed %s/%s/%s when update, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
			return fmt.Errorf("search failed %s/%s/%s when update, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
		}
		found = false
	}
	if !opt.CreateNotExists && !found {
		blog.Errorf("object %s/%s/%s to update does not exists",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
		return fmt.Errorf("object %s/%s/%s to update does not exists",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
	}

	var err error
	if found {
		obj.SetCreateTime(rawObj.GetCreateTime())
		obj.SetUpdateTime(time.Now())
		err = s.mDriver.Table(string(obj.GetObjectType())).
			Update(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
				types.TagResourceName: obj.GetName(),
				types.TagNamespace:    obj.GetNamespace(),
				types.TagClusterID:    obj.GetClusterID(),
			}), operator.M{"$set": obj})
	} else {
		obj.SetCreateTime(time.Now())
		obj.SetUpdateTime(time.Now())
		_, err = s.mDriver.Table(string(obj.GetObjectType())).Insert(ctx, []interface{}{obj})
	}
	if err != nil {
		blog.Errorf("update object %s failed, err %s", obj.ToString(), err.Error())
		return fmt.Errorf("update object %s failed, err %s", obj.ToString(), err.Error())
	}
	return nil
}

// Delete delete object
func (s *Store) Delete(ctx context.Context, obj *types.RawObject, opt *store.DeleteOptions) error {
	// Check if the obj or object type is empty, if so, return an error
	if obj == nil || len(obj.GetObjectType()) == 0 {
		return fmt.Errorf("object or object type cannot be empty")
	}

	// Check if the delete options are empty, if so, return an error
	if opt == nil {
		return fmt.Errorf("update options cannot be empty")
	}

	// Ensure that the collection exists in the database
	if err := s.ensureCollection(ctx, obj); err != nil {
		return err
	}

	rawObj := types.RawObject{}
	found := true
	// Find the RawObject in the database using the object name, namespace, and cluster ID
	if err := s.mDriver.Table(string(obj.GetObjectType())).
		Find(operator.NewLeafCondition(operator.Eq, operator.M{
			types.TagResourceName: obj.GetName(),
			types.TagNamespace:    obj.GetNamespace(),
			types.TagClusterID:    obj.GetClusterID(),
		})).One(ctx, &rawObj); err != nil {
		if !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("search failed %s/%s/%s when delete, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
			return fmt.Errorf("search failed %s/%s/%s when delete, err %s",
				obj.GetName(), obj.GetNamespace(), obj.GetClusterID(), err.Error())
		}
		found = false
	}
	if !found && !opt.IgnoreNotFound {
		blog.Errorf("object %s/%s/%s to be deleted not found",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
		return fmt.Errorf("object %s/%s/%s to be deleted not found",
			obj.GetName(), obj.GetNamespace(), obj.GetClusterID())
	}

	return nil
}

// List list objects
func (s *Store) List(ctx context.Context, objectType types.ObjectType, opts *store.ListOptions) (
	[]*types.RawObject, error) {

	// Create a conditionValue map and add any relevant conditions based on the ListOptions
	conditionValue := operator.M{}
	if len(opts.Cluster) != 0 {
		conditionValue[types.TagClusterID] = opts.Cluster
	}
	if len(opts.Namespace) != 0 {
		conditionValue[types.TagNamespace] = opts.Namespace
	}
	if opts.Selector != nil {
		pairs := opts.Selector.GetPairs()
		for path, value := range pairs {
			conditionValue[path] = value
		}
	}

	// Create a condition using the conditionValue map
	condition := operator.NewLeafCondition(operator.Eq, conditionValue)
	// Create a finder using the objectType and condition
	finder := s.mDriver.Table(string(objectType)).Find(condition)
	// Add any relevant options to the finder
	if opts.Offset != 0 {
		finder = finder.WithStart(opts.Offset)
	}
	if opts.Limit != 0 {
		finder = finder.WithLimit(opts.Limit)
	}

	// Create a slice of RawObjects and execute the finder
	rawList := make([]*types.RawObject, 0)
	err := finder.All(ctx, &rawList)
	if err != nil {
		return nil, err
	}
	return rawList, nil
}

// Watch object with certain type
func (s *Store) Watch(ctx context.Context, resourceType types.ObjectType, opt *store.WatchOptions) (chan *store.Event,
	error) {
	// Check if the watch options are empty, if so, return an error
	if opt == nil {
		return nil, fmt.Errorf("update options cannot be empty")
	}
	// Create a conditionValue map and add any relevant conditions based on the WatchOptions
	conditionValue := operator.M{}
	if opt.Selector != nil {
		pairs := opt.Selector.GetPairs()
		for path, value := range pairs {
			conditionValue[path] = value
		}
	}
	conditionList := make([]*operator.Condition, 0)
	if len(conditionValue) != 0 {
		condition := operator.NewLeafCondition(operator.Eq, conditionValue)
		condition = operator.NewBranchCondition(operator.Mat, condition)
		conditionList = append(conditionList, condition)
	}

	// Create a watcher using the resourceType and conditionList
	watcher := s.mDriver.Table(string(resourceType)).Watch(conditionList)
	if opt.BatchSize != 0 {
		watcher.WithBatchSize(opt.BatchSize)
	}
	// always return full document
	watcher.WithFullContent(true)
	// Add any relevant options to the watcher
	if opt.MaxAwaitTime != 0 {
		watcher.WithMaxAwaitTime(opt.MaxAwaitTime)
	}
	if opt.StartTime != nil {
		watcher.WithStartTimestamp(opt.StartTime.T, opt.StartTime.I)
	}
	// Execute the watcher and create a storeChannel to return
	mongoChannel, err := watcher.DoWatch(ctx)
	if err != nil {
		return nil, err
	}
	storeChannel := make(chan *store.Event, 100)
	// Start a goroutine to decode the mongo events and send them to the storeChannel
	go func() {
		for {
			select {
			case mEvent := <-mongoChannel:
				sEvent, inErr := decodeMongoEvent(mEvent)
				if inErr != nil {
					blog.Infof("delete mongo event %+v failed, err %s", mEvent, inErr)
					storeChannel <- &store.Event{
						Type: store.EventError,
					}
					return
				}
				storeChannel <- sEvent
				if sEvent.Type == store.EventError || sEvent.Type == store.EventClose {
					return
				}
			case <-ctx.Done():
				storeChannel <- &store.Event{
					Type: store.EventClose,
				}
				return
			}
		}
	}()
	return storeChannel, nil
}
