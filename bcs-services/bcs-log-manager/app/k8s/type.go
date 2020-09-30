/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package k8s

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/api/proto/logmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	internalclientset "github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/pkg/generated/clientset/versioned"
)

// RequestMessage is structure of information extrange between goroutines
// Data is request data
// RespCh is response data channel
type RequestMessage struct {
	Data   interface{}
	RespCh chan interface{}
}

// LogManagerInterface defines the interface of log manager service
type LogManagerInterface interface {
	Start()
	HandleListLogCollectionTask(context.Context, *config.CollectionFilterConfig) map[string][]config.CollectionConfig
	HandleAddLogCollectionTask(context.Context, *config.CollectionConfig) *proto.CollectionTaskCommonResp
	HandleDeleteLogCollectionTask(context.Context, *config.CollectionFilterConfig) *proto.CollectionTaskCommonResp
}

// LogManager contains the log-manager module's main funcations
type LogManager struct {
	GetLogCollectionTask    chan *RequestMessage
	AddLogCollectionTask    chan *RequestMessage
	DeleteLogCollectionTask chan *RequestMessage

	userManagerCli bcsapi.UserManager
	config         *config.ManagerConfig
	// controllers              map[string]*ClusterLogController
	clientRWMutex            sync.RWMutex
	logClients               map[string]*LogClient
	dataidChMap              map[string]chan string
	currCollectionConfigInd  int
	bkDataAPIConfigClientset *internalclientset.Clientset
	bkDataAPIConfigInformer  cache.SharedIndexInformer
	stopCh                   chan struct{}
	ctx                      context.Context
}

// LogClient is client for BcsLogConfigs operation of single cluster
type LogClient struct {
	ClusterInfo *bcsapi.ClusterCredential
	Client      rest.Interface
}

// GetRateLimiter is a passthrough to rest.RESTClient
func (lc *LogClient) GetRateLimiter() flowcontrol.RateLimiter { return lc.Client.GetRateLimiter() }

// Verb is a passthrough to rest.RESTClient
func (lc *LogClient) Verb(verb string) *rest.Request { return lc.Client.Verb(verb) }

// Post is a passthrough to rest.RESTClient
func (lc *LogClient) Post() *rest.Request { return lc.Client.Post() }

// Put is a passthrough to rest.RESTClient
func (lc *LogClient) Put() *rest.Request { return lc.Client.Put() }

// Patch is a passthrough to rest.RESTClient
func (lc *LogClient) Patch(pt types.PatchType) *rest.Request { return lc.Client.Patch(pt) }

// Get is a passthrough to rest.RESTClient
func (lc *LogClient) Get() *rest.Request { return lc.Client.Get() }

// Delete is a passthrough to rest.RESTClient
func (lc *LogClient) Delete() *rest.Request { return lc.Client.Delete() }

// APIVersion is a passthrough to rest.RESTClient
func (lc *LogClient) APIVersion() schema.GroupVersion { return lc.Client.APIVersion() }
