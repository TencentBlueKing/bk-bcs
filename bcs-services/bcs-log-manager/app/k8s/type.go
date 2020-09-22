package k8s

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
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
