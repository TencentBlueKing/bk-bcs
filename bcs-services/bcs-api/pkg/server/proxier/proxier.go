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

package proxier

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/metric"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/auth"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/credentials"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/utils"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

// ReverseProxyDispatcher is the handler which dispatch and proxy the incoming requests to external
// apiservers.
type ReverseProxyDispatcher struct {
	// ClusterVarName is the path parameter name of cluster identifier
	ClusterVarName string
	// ClusterVarName is the path parameter name of sub-path needs to be forwarded
	SubPathVarName string

	handlerStore      map[string]*ClusterHandlerInstance
	handlerMutateLock sync.RWMutex
	// Credential backend storages
	credentialBackends []credentials.CredentialBackend

	availableSrvStore map[string]*UpstreamServer

	wsTunnelStore      map[string]*WsTunnel
	wsTunnelMutateLock sync.RWMutex
}

type ClusterHandlerInstance struct {
	ServerAddress string
	Handler       http.Handler
}

func NewReverseProxyDispatcher(clusterVarName, subPathVarName string) *ReverseProxyDispatcher {
	return &ReverseProxyDispatcher{
		ClusterVarName:    clusterVarName,
		SubPathVarName:    subPathVarName,
		handlerStore:      make(map[string]*ClusterHandlerInstance),
		availableSrvStore: make(map[string]*UpstreamServer),
		wsTunnelStore:     make(map[string]*WsTunnel),
	}
}

var DefaultReverseProxyDispatcher = NewReverseProxyDispatcher("cluster_identifier", "sub_path")

// Initialize the required components for dispatcher
func (f *ReverseProxyDispatcher) Initialize() {
	credentials.GFixtureCredentialBackend.ExtractCredentialsFixtures()
	// Load default backends for credentials
	f.credentialBackends = append(f.credentialBackends, credentials.GDatabaseCrendentialBackend)
	f.credentialBackends = append(f.credentialBackends, credentials.GFixtureCredentialBackend)
}

func (f *ReverseProxyDispatcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	start := time.Now()

	vars := mux.Vars(req)
	// Get current cluster object
	clusterIdentifier := vars[f.ClusterVarName]
	if clusterIdentifier == "" {
		metric.RequestErrorCount.WithLabelValues("k8s_native", req.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_native", req.Method).Observe(time.Since(start).Seconds())
		err := fmt.Errorf("cluster_id is required in path parameters")
		status := utils.NewInvalid(utils.ClusterGroupKind, "cluster", f.ClusterVarName, err)
		utils.WriteKubeAPIError(rw, status)
		return
	}

	// Try to get the clusterId by given clusterIdentifier
	cluster := f.GetCluster(clusterIdentifier)
	if cluster == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_native", req.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_native", req.Method).Observe(time.Since(start).Seconds())
		message := "no cluster can be found using given cluster identifier"
		status := utils.NewNotFound(utils.ClusterResource, clusterIdentifier, message)
		utils.WriteKubeAPIError(rw, status)
		return
	}
	clusterId := cluster.ID

	// Authenticate user
	var authenticater *auth.TokenAuthenticater
	authenticater = auth.NewTokenAuthenticater(req, &auth.TokenAuthConfig{
		SourceBearerEnabled: true,
	})

	user, hasExpired := authenticater.GetUser()
	if user == nil {
		metric.RequestErrorCount.WithLabelValues("k8s_native", req.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_native", req.Method).Observe(time.Since(start).Seconds())
		status := utils.NewUnauthorized("anonymous requests is forbidden")
		utils.WriteKubeAPIError(rw, status)
		return
	}
	if hasExpired {
		metric.RequestErrorCount.WithLabelValues("k8s_native", req.Method).Inc()
		metric.RequestErrorLatency.WithLabelValues("k8s_native", req.Method).Observe(time.Since(start).Seconds())
		reason := fmt.Sprintf("this token has expired for user: %s", user.Name)
		status := utils.NewUnauthorized(reason)
		utils.WriteKubeAPIError(rw, status)
		return
	}

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")

	var proxyHandler *ClusterHandlerInstance
	// 先从websocket dialer缓存中查找websocket链
	websocketHandler, found, err := f.lookupWsHandler(clusterId, req)
	if err != nil {
		blog.Errorf("error when lookup websocket conn: %s", err.Error())
		err := fmt.Errorf("error when lookup websocket conn: %s", err.Error())
		status := utils.NewInternalError(err)
		status.ErrStatus.Reason = "CREATE_TUNNEL_ERROR"
		utils.WriteKubeAPIError(rw, status)
		return
	}
	if found {
		blog.Info("found websocket conn for cluster %s", clusterId)
		handlerServer := stripLeaveSlash(f.ExtractPathPrefix(req), websocketHandler)
		proxyHandler = &ClusterHandlerInstance{
			Handler: handlerServer,
		}
		credentials := sqlstore.GetWsCredentials(clusterId)
		bearerToken := "Bearer " + credentials.UserToken
		req.Header.Set("Authorization", bearerToken)
	} else {
		// Try not to initialize the handler everytime by using a map to store all the initialized handler
		// Use RWLock to fix race condition
		f.handlerMutateLock.Lock()
		if f.handlerStore[clusterId] == nil {
			handlerServer, err := f.InitializeHandlerForCluster(clusterId, req)
			if err != nil {
				err = fmt.Errorf("error when creating proxy channel: %s", err.Error())
				status := utils.NewInternalError(err)
				status.ErrStatus.Reason = "CREATE_TUNNEL_ERROR"
				utils.WriteKubeAPIError(rw, status)
				f.handlerMutateLock.Unlock()
				return
			}
			f.handlerStore[clusterId] = handlerServer
		}
		proxyHandler = f.handlerStore[clusterId]
		f.handlerMutateLock.Unlock()
	}

	// Add the user name to Header to pass to k8s cluster, implement the user Impersonate feature
	// Because k8s rbac doesn't allow label to contain ":", so replaced by "."
	turnOnRbac := config.TurnOnRBAC
	if turnOnRbac {
		if !user.IsSuperUser {
			req.Header.Set("Impersonate-User", strings.Replace(user.Name, ":", ".", 1))
		}
	}

	// TODO: How to modify the rev-proxy request to allow user pass the ORIGINAL CLUSTER CA instead of the ca of current
	// bke-server instance?
	req.URL.Scheme = "https"

	if websocket.IsWebSocketUpgrade(req) {
		metric.RequestCount.WithLabelValues("k8s_native", "websocket").Inc()
		metric.RequestLatency.WithLabelValues("k8s_native", "websocket").Observe(time.Since(start).Seconds())
	}
	proxyHandler.Handler.ServeHTTP(rw, req)
	if !websocket.IsWebSocketUpgrade(req) {
		metric.RequestCount.WithLabelValues("k8s_native", req.Method).Inc()
		metric.RequestLatency.WithLabelValues("k8s_native", req.Method).Observe(time.Since(start).Seconds())
	}
	return
}

// InitializeUpstreamServer initialize the upstreamServer instance for cluster
func (f *ReverseProxyDispatcher) InitializeUpstreamServer(clusterId string, serverAddresses []string) {
	// Only create the upstremServer instance for once
	if _, ok := f.availableSrvStore[clusterId]; ok {
		return
	}

	upstreamServer := NewUpstreamServer(clusterId, serverAddresses, func() {
		blog.Infof("endpoints availablility changes, delete cached proxy handler instance for cluster<%s>", clusterId)
		f.DelHandlerStoreByClusterId(clusterId)
	})
	upstreamServer.Initialize()
	f.availableSrvStore[clusterId] = upstreamServer

	// Starts a new period checker to notify the upstreamServer when cluster's apiservers have been majorly changed
	go f.StartClusterAddressesPoller(clusterId)
}

// InitializeHandlerForCluster was called when a cluster channel is requested for the first time. There are also
// other cases when we may also need to re-establish the apiserver connection. This includes apiserver connection
// failure or apiserver addresses's major changes.
func (f *ReverseProxyDispatcher) InitializeHandlerForCluster(clusterId string, req *http.Request) (*ClusterHandlerInstance, error) {

	// Query for the cluster credentials
	clusterCredentials := f.GetClusterCredentials(clusterId)
	if clusterCredentials == nil || clusterCredentials.ServerAddresses == "" {
		blog.Error("cluster has no credentials or its apiserver addresses field is empty")
		return nil, errors.New("cluster has no credentials or its apiserver addresses field is empty")
	}

	f.InitializeUpstreamServer(clusterId, clusterCredentials.GetServerAddressesList())

	// Pick one available apiserver address
	clusterCredentials.ServerAddresses = f.availableSrvStore[clusterId].GetAvailableServer()
	blog.Infof("Init new proxy handler for %s, using address: %s", clusterId, clusterCredentials.ServerAddresses)
	restConfig, err := TurnCredentialsIntoConfig(clusterCredentials)
	if err != nil {
		blog.Errorf("TurnCredentialsIntoConfig failed: %s", err.Error())
		return nil, fmt.Errorf("error when turning credentials into restconfig: %s", err.Error())
	}

	handler, err := NewProxyHandlerFromConfig(restConfig)
	if err != nil {
		blog.Errorf("NewProxyHandlerFromConfig failed: %s \n restConfig is: %+v", err.Error(), restConfig)
		return nil, err
	}
	// Strip the path prefix to make sure the proxy works
	handlerServer := stripLeaveSlash(f.ExtractPathPrefix(req), handler)
	return &ClusterHandlerInstance{
		ServerAddress: clusterCredentials.ServerAddresses,
		Handler:       handlerServer,
	}, nil

}

func (f *ReverseProxyDispatcher) StartClusterAddressesPoller(clusterId string) {
	refreshTicker := time.NewTicker(60 * time.Second)
	defer refreshTicker.Stop()
	upstreamServer := f.availableSrvStore[clusterId]
	for {
		select {
		case <-refreshTicker.C:
			existedHander := f.handlerStore[clusterId]
			if existedHander == nil {
				continue
			}
			// If cluster's apiserver addresses have been updated, we will notify the upstreamServer to
			// update the servers.
			clusterCredentials := f.GetClusterCredentials(clusterId)
			if clusterCredentials == nil {
				blog.Infof("no credentials for cluster[%s], so stop monitors for it", clusterId)
				upstreamServer.Stop()
				return
			}
			currentAddresses := clusterCredentials.GetServerAddressesList()
			if !cmp.Equal(currentAddresses, upstreamServer.servers) {
				blog.Infof("update server addresses for cluster[%s], new value: %s", clusterId, currentAddresses)
				upstreamServer.UpdateServerAddresses(currentAddresses)
			}
		}
	}
}

// delHandlerStoreByClusterId used when delete the cluster or switch available server
func (f *ReverseProxyDispatcher) DelHandlerStoreByClusterId(clusterId string) {
	defer f.handlerMutateLock.Unlock()
	f.handlerMutateLock.Lock()
	delete(f.handlerStore, clusterId)
}

// GetCluster loop over all available storage backends to find the cluster for given identifier
func (f *ReverseProxyDispatcher) GetCluster(clusterIdentifier string) *m.Cluster {
	for _, storage := range f.credentialBackends {
		result, _ := storage.GetClusterByIdentifier(clusterIdentifier)
		if result != nil {
			return result
		}
	}
	return nil
}

// GetClusterCredentials loop over all available storage backends to find the credentials for given clusterId
func (f *ReverseProxyDispatcher) GetClusterCredentials(clusterId string) *m.ClusterCredentials {
	for _, storage := range f.credentialBackends {
		result, _ := storage.GetCredentials(clusterId)
		if result != nil {
			return result
		}
	}
	return nil
}

// ExtractPathPrefix extracts the path prefix which needs to be stripped when the request is forwarded to the reverse
// proxy handler.
func (f *ReverseProxyDispatcher) ExtractPathPrefix(req *http.Request) string {
	subPath := mux.Vars(req)[f.SubPathVarName]
	fullPath := req.URL.Path
	// We need to strip the prefix string before the request can be forward to apiserver, so we will walk over the full
	// request path backwards, everything before the `sub_path` will be the prefix we need to strip
	return fullPath[:len(fullPath)-len(subPath)]
}

type responder struct{}

func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// NewProxyHandler creates a new proxy handler to a single api server based on the given kube config object
func NewProxyHandlerFromConfig(config *rest.Config) (*proxy.UpgradeAwareHandler, error) {

	host := config.Host
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	target, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	responder := &responder{}
	apiTransport, err := rest.TransportFor(config)
	if err != nil {
		return nil, err
	}

	keepalive := 0 * time.Second
	upgradeTransport, err := makeUpgradeTransport(config, keepalive)
	if err != nil {
		return nil, err
	}

	apiProxy := proxy.NewUpgradeAwareHandler(target, apiTransport, false, false, responder)
	apiProxy.UpgradeTransport = upgradeTransport
	apiProxy.UseRequestLocation = true
	return apiProxy, nil
}

// makeUpgradeTransport creates a transport that explicitly bypasses HTTP2 support
// for proxy connections that must upgrade.
func makeUpgradeTransport(config *rest.Config, keepalive time.Duration) (proxy.UpgradeRequestRoundTripper, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: keepalive,
	}
	ipAddress, err := ExtractIpAddress(config.Host)
	if err != nil {
		return nil, err
	}
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		TLSClientConfig: tlsConfig,
		Dial: func(network, addr string) (net.Conn, error) {
			// resolve domain to real apiserver address
			addr = ipAddress.Host
			return dialer.Dial(network, addr)
		},
	})

	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

// like http.StripPrefix, but always leaves an initial slash. (so that our
// regexps will work.)
func stripLeaveSlash(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		blog.Debug(fmt.Sprintf("begin proxy for: %s", req.URL.Path))
		p := strings.TrimPrefix(req.URL.Path, prefix)
		if len(p) >= len(req.URL.Path) {
			http.NotFound(w, req)
			return
		}
		if len(p) > 0 && p[:1] != "/" {
			p = "/" + p
		}
		req.URL.Path = p
		h.ServeHTTP(w, req)
	})
}
