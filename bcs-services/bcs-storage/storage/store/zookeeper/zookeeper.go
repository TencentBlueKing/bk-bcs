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

package zookeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

// Options options for zookeeper store
type Options struct {
	Addrs                 []string
	ConnectTimeoutSeconds int
	Database              string
	Username              string
	Password              string
}

// Store client for zookeeper
type Store struct {
	basePath string
	zk       *zkclient.ZkClient
}

func getClusterType(clusterID string) string {
	if strings.Contains(clusterID, "mesos") {
		return "mesos"
	} else if strings.Contains(clusterID, "k8s") {
		return "k8s"
	}
	return ""
}

func (s *Store) getObjectPath(objectType, clusterID, namespace, name string) string {
	path := s.basePath
	sysType := getClusterType(clusterID)
	if len(sysType) != 0 {
		path = filepath.Join(s.basePath, sysType)
	}
	return filepath.Join(path, clusterID, string(objectType), namespace+"."+name)
}

// Get implement Store
func (s *Store) Get(ctx context.Context, resourceType types.ObjectType, key types.ObjectKey) (*types.RawObject, error) {
	if len(key.ClusterID) == 0 || len(key.Name) == 0 || len(key.Namespace) == 0 {
		return nil, fmt.Errorf("field in object key cannot be empty")
	}
	path := s.getObjectPath(string(resourceType), key.ClusterID, key.Namespace, key.Name)
	content, err := s.zk.Get(path)
	if err != nil {
		blog.Errorf("zk get path %s failed, err %s", path, err.Error())
		return nil, err
	}

	data := make(operator.M)
	err = json.Unmarshal([]byte(content), &data)
	if err != nil {
		blog.Errorf("decode content of zk path %s failed, err %s", path, err.Error())
		return nil, err
	}

	return &types.RawObject{
		Meta: types.Meta{
			Type:      resourceType,
			ClusterID: key.ClusterID,
			Namespace: key.Namespace,
			Name:      key.Name,
		},
		Data: data,
	}, nil
}

// Create implement Store
func (s *Store) Create(ctx context.Context, obj *types.RawObject, opt *store.CreateOptions) error {
	if len(obj.GetObjectType()) == 0 ||
		len(obj.GetName()) == 0 ||
		len(obj.GetNamespace()) == 0 ||
		len(obj.GetClusterID()) == 0 {

		return fmt.Errorf("field type, name, namespace, clusterid cannot be empty")
	}
	path := s.getObjectPath(
		string(obj.GetObjectType()),
		obj.GetClusterID(),
		obj.GetNamespace(),
		obj.GetName())

	data, err := json.Marshal(obj.GetData())
	if err != nil {
		blog.Errorf("get data of type %s, cluster %s, ns %s, name %s failed, err %s",
			obj.GetObjectType(), obj.GetClusterID(), obj.GetNamespace(),
			obj.GetName(), err.Error())
		return fmt.Errorf("get data of type %s, cluster %s, ns %s, name %s failed, err %s",
			obj.GetObjectType(), obj.GetClusterID(), obj.GetNamespace(),
			obj.GetName(), err.Error())
	}

	isExisted, err := s.zk.Exist(path)
	if err != nil {
		return err
	}
	if isExisted {
		if opt.UpdateExists {
			err = s.zk.Update(path, string(data))
			if err != nil {
				blog.Errorf("update zk path %s failed, err %s", path, err.Error())
				return fmt.Errorf("update zk path %s failed, err %s", path, err.Error())
			}
			return nil
		}
		return fmt.Errorf("path %s to create already exists", path)
	}
	s.zk.CreateDeepNode(path, data)
	if err != nil {
		blog.Errorf("create path %s failed, err %s", path, err.Error())
		return fmt.Errorf("create path %s failed, err %s", path, err.Error())
	}
	return nil
}

// Update implement Store
func (s *Store) Update(ctx context.Context, obj *types.RawObject, opt *store.UpdateOptions) error {
	if len(obj.GetObjectType()) == 0 ||
		len(obj.GetName()) == 0 ||
		len(obj.GetNamespace()) == 0 ||
		len(obj.GetClusterID()) == 0 {

		return fmt.Errorf("field type, name, namespace, clusterid cannot be empty")
	}
	path := s.getObjectPath(
		string(obj.GetObjectType()),
		obj.GetClusterID(),
		obj.GetNamespace(),
		obj.GetName())

	data, err := json.Marshal(obj.GetData())
	if err != nil {
		blog.Errorf("get data of type %s, cluster %s, ns %s, name %s failed, err %s",
			obj.GetObjectType(), obj.GetClusterID(), obj.GetNamespace(),
			obj.GetName(), err.Error())
		return fmt.Errorf("get data of type %s, cluster %s, ns %s, name %s failed, err %s",
			obj.GetObjectType(), obj.GetClusterID(), obj.GetNamespace(),
			obj.GetName(), err.Error())
	}

	isExisted, err := s.zk.Exist(path)
	if err != nil {
		return err
	}
	if !isExisted {
		if opt.CreateNotExists {
			err = s.zk.CreateDeepNode(path, data)
			if err != nil {
				blog.Errorf("create path %s failed, err %s", path, err.Error())
				return fmt.Errorf("create path %s failed, err %s", path, err.Error())
			}
			return nil
		}
		return fmt.Errorf("path %s to update is not existed", path)
	}
	err = s.zk.Update(path, string(data))
	if err != nil {
		blog.Errorf("update zk path %s failed, err %s", path, err.Error())
		return fmt.Errorf("update zk path %s failed, err %s", path, err.Error())
	}
	return nil
}

// Delete implement Store
func (s *Store) Delete(ctx context.Context, obj *types.RawObject, opt *store.DeleteOptions) error {
	if len(obj.GetObjectType()) == 0 ||
		len(obj.GetName()) == 0 ||
		len(obj.GetNamespace()) == 0 ||
		len(obj.GetClusterID()) == 0 {

		return fmt.Errorf("field type, name, namespace, clusterid cannot be empty")
	}
	path := s.getObjectPath(
		string(obj.GetObjectType()),
		obj.GetClusterID(),
		obj.GetNamespace(),
		obj.GetName())
	isExisted, err := s.zk.Exist(path)
	if err != nil {
		return err
	}
	if !isExisted {
		if opt.IgnoreNotFound {
			return nil
		}
		blog.Errorf("path %s to be deleted not found", path)
		return fmt.Errorf("path %s to be deleted not found", path)
	}
	err = s.zk.Del(path, 0)
	if err != nil {
		blog.Errorf("del path %s failed, err %s", path, err.Error())
		return fmt.Errorf("del path %s failed, err %s", path, err.Error())
	}
	return nil
}

func matchSelector(data operator.M, selectorMap map[string]interface{}) bool {
	for k, v := range selectorMap {
		value, ok := data[k]
		if !ok {
			return false
		}
		if !reflect.DeepEqual(value, v) {
			return false
		}
	}
	return true
}

// List implement Store
func (s *Store) List(ctx context.Context, objectType types.ObjectType, opts *store.ListOptions) (
	[]*types.RawObject, error) {

	if len(opts.Cluster) != 0 {
		return nil, fmt.Errorf("cluster cannot be empty")
	}

	path := s.basePath
	sysType := getClusterType(opts.Cluster)
	if len(sysType) != 0 {
		path = filepath.Join(s.basePath, sysType)
	}
	parentPath := filepath.Join(path, opts.Cluster)

	children, err := s.zk.GetChildren(parentPath)
	if err != nil {
		blog.Errorf("get children of path %s failed, err %s", parentPath, err.Error())
		return nil, fmt.Errorf("get children of path %s failed, err %s", parentPath, err.Error())
	}
	if len(opts.Namespace) != 0 {
		var tmpList []string
		for _, child := range children {
			if strings.HasPrefix(child, opts.Namespace+".") {
				tmpList = append(tmpList, child)
			}
		}
		children = tmpList
	}

	var retList []*types.RawObject
	for _, child := range children {
		childData, err := s.zk.Get(filepath.Join(parentPath, child))
		if err != nil {
			blog.Errorf("get data of path %s failed, err %s", filepath.Join(parentPath, child), err.Error())
			return nil, fmt.Errorf("get data of path %s failed, err %s", filepath.Join(parentPath, child), err.Error())
		}
		childM := make(operator.M)
		err = json.Unmarshal([]byte(childData), &childM)
		if err != nil {
			blog.Errorf("decode data failed, err %s", err.Error())
			return nil, fmt.Errorf("decode data failed, err %s", err.Error())
		}
		if opts.Selector != nil {
			if matchSelector(childM, opts.Selector.GetPairs()) {
				strs := strings.Split(child, ".")
				if len(strs) != 0 {
					blog.Errorf("child %s is invalid, should include .", child)
					return nil, fmt.Errorf("child %s is invalid, should include \".\"", child)
				}
				retList = append(retList, &types.RawObject{
					Meta: types.Meta{
						Type:      objectType,
						ClusterID: opts.Cluster,
						Namespace: strs[0],
						Name:      strs[1],
					},
					Data: childM,
				})
			}
		}
	}
	return retList, nil
}

// Watch implement Store
func (s *Store) Watch(ctx context.Context, resourceType types.ObjectType, opts *store.WatchOptions) (
	chan *store.Event, error) {
	return nil, nil
}
