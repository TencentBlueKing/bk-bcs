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
 */

// Package http xxx
package http

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	syshttp "net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/meta"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/watch"
)

const (
	// http api default prefix
	defaultAPIPrefix = "apis"
)

var (
	// PrevDataErr this error means operation success, but got previous data failed
	PrevDataErr = errors.New("Previous data err") // nolint
)

// Config etcd storage config
type Config struct {
	Hosts          []string             // http api host link, http://ip:host
	User           string               // user name for http basic authentication
	Passwd         string               // password relative to user
	Codec          meta.Codec           // Codec for encoder & decoder
	ObjectNewFunc  meta.ObjectNewFn     // object pointer for serialization
	ObjectListFunc meta.ObjectListNewFn // decode raw json data to object list
	TLS            *tls.Config          // tls config for https
}

// NewStorage create etcd accessor implemented storage interface
func NewStorage(config *Config) (storage.Storage, error) {
	if config == nil {
		return nil, fmt.Errorf("lost client configuration")
	}
	return NewClient(config)
}

// NewClient create new client for http event apis
func NewClient(config *Config) (*Client, error) {
	if len(config.Hosts) == 0 {
		return nil, fmt.Errorf("Lost http api hosts info")
	}
	c := &syshttp.Client{
		Transport: &syshttp.Transport{
			TLSClientConfig:     config.TLS,
			TLSHandshakeTimeout: time.Second * 3,
			IdleConnTimeout:     time.Second * 300,
		},
	}
	s := &Client{
		client:       c,
		codec:        config.Codec,
		objectNewFn:  config.ObjectNewFunc,
		objectListFn: config.ObjectListFunc,
		servers:      config.Hosts,
	}
	return s, nil
}

// Client implementation storage interface with etcd client
type Client struct {
	client       *syshttp.Client      // http client
	codec        meta.Codec           // json Codec for object
	objectNewFn  meta.ObjectNewFn     // create new object for json Decode
	objectListFn meta.ObjectListNewFn // decode json list objects
	servers      []string             // server http api prefix
}

// Create implements storage interface
// param cxt: context for use decline Creation, not used
// param key: http full api path
// param obj: object for creation
// param ttl: second for time-to-live, not used
// return out: exist object data
func (s *Client) Create(_ context.Context, key string, obj meta.Object, _ int) (meta.Object, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("http client lost object key")
	}
	if obj == nil {
		blog.V(3).Infof("http storage client lost object data for %s", key)
		return nil, fmt.Errorf("lost object data")
	}
	// serialize object
	data, err := s.codec.Encode(obj)
	if err != nil {
		blog.V(3).Infof("http storage client encode %s/%s for %s failed, %s", obj.GetNamespace(), obj.GetName(), key, err)
		return nil, fmt.Errorf("encode %s/%s: %s", obj.GetNamespace(), obj.GetName(), err)
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.selectServers(), defaultAPIPrefix, key)
	// check path for http method
	method := "POST"
	if strings.Contains(key, "namespace") {
		// this means updating with detail url
		method = "PUT"
	}
	blog.V(3).Infof("obj encoded data %s", string(data))
	request, err := syshttp.NewRequest(method, fullPath, bytes.NewBuffer(data))
	if err != nil {
		blog.V(3).Infof("http storage client create %s request for %s failed, %s", method, fullPath, err)
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		blog.V(3).Infof("http storage client %s with %s for %s/%s failed, %s", method, fullPath, obj.GetNamespace(),
			obj.GetName(), err)
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)
	if response.StatusCode != syshttp.StatusOK {
		blog.V(3).Infof("http storage %s with %s failed, http response code: %d, status: %s", method, fullPath,
			response.StatusCode, response.Status)
		return nil, fmt.Errorf("response: %d, message: %s", response.StatusCode, response.Status)
	}
	rawData, err := io.ReadAll(response.Body)
	if err != nil {
		blog.V(3).Infof("http storage client read %s %s response body failed, %s. Operation status unknown.", method,
			fullPath, err)
		return nil, err
	}
	// format data
	standarResponse := &Response{}
	if err := json.Unmarshal(rawData, standarResponse); err != nil {
		blog.V(3).Infof("http storage client parse %s %s response failed, %s. Operation status unknown", method, fullPath,
			err)
		return nil, err
	}
	if standarResponse.Code != 0 {
		blog.V(3).Infof("http storage %s %s failed, %s", method, fullPath, standarResponse.Message)
		return nil, fmt.Errorf("%s", standarResponse.Message)
	}
	// check exist data
	if len(standarResponse.Data) == 0 {
		blog.V(5).Infof("http storage %s %s got no response data", method, fullPath)
		return nil, nil
	}
	target := s.objectNewFn()
	if err := s.codec.Decode(standarResponse.Data, target); err != nil {
		blog.V(3).Infof("http storage decode %s %s Previous value failed, %s", method, fullPath, err)
		// even got previous data failed, we still consider Create successfully
		return nil, PrevDataErr
	}
	blog.V(3).Infof("etcd storage %s %s & got previous kv success", method, fullPath)
	return target, nil
}

// Delete implements storage interface
// for http api operation, there are three situations for key
// * if key likes apis/v1/dns, clean all dns data under version v1
// * if key likes apis/v1/dns/cluster/$clustername, delete all data under cluster
// * if key likes apis/v1/dns/cluster/$clustername/namespace/bmsf-system, delete all data under namespace
// * if key likes apis/v1/dns/cluster/$clustername/namespace/bmsf-system/data, delete detail data
// in this version, no delete objects reply
func (s *Client) Delete(_ context.Context, key string) (obj meta.Object, err error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("empty key")
	}
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error format key, cannot end with /")
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.selectServers(), defaultAPIPrefix, key)
	request, err := syshttp.NewRequest("DELETE", fullPath, nil)
	if err != nil {
		blog.V(3).Infof("http storage client create request for %s failed, %s", fullPath, err)
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		blog.V(3).Infof("http storage client DELETE request to %s failed, %s", fullPath, err)
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)
	if response.StatusCode < syshttp.StatusOK || response.StatusCode >= syshttp.StatusMultipleChoices {
		blog.V(3).Infof("http storage client delete to %s failed, code: %d, message: %s", fullPath, response.StatusCode,
			response.Status)
		return nil, fmt.Errorf("delete response failed, code: %d, status: %s", response.StatusCode, response.Status)
	}
	rawByets, err := io.ReadAll(response.Body)
	if err != nil {
		blog.V(3).Infof("http storage delete %s http status success, but read response body failed, %s", fullPath, err)
		return nil, err
	}
	// format http response
	standarResponse := &Response{}
	if err := json.Unmarshal(rawByets, standarResponse); err != nil {
		blog.V(3).Infof("http storage decode %s http response failed, %s", fullPath, err)
		return nil, err
	}
	if standarResponse.Code != 0 {
		blog.V(3).Infof("http storage delete %s failed, %s", fullPath, standarResponse.Message)
		return nil, fmt.Errorf("remote err: %s", standarResponse.Message)
	}
	blog.V(3).Infof("http storage delete %s success, status: %s", fullPath, standarResponse.Message)
	return nil, nil
}

// Watch implements storage interface
// * if key empty, watch all data
// * if key is namespace, watch all data under namespace
// * if key is namespace/name, watch detail data
// watch is Stopped when any error occure, close event channel immediately
// param cxt: context for background running, not used, only reserved now
// param version: data version, not used, reserved
// param selector: labels selector
// return:
//
//	watch: watch implementation for changing event, need to Stop manually
func (s *Client) Watch(_ context.Context, key, _ string, selector storage.Selector) (watch.Interface, error) {
	if len(key) == 0 || strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key formate")
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.selectServers(), defaultAPIPrefix, key)
	if selector != nil {
		// fullPath = fullPath + "?labelSelector=" + selector.String() + "&watch=true"
		fullPath = fullPath + "?" + selector.String() + "&watch=true"
	} else {
		fullPath += "?watch=true"
	}
	request, err := syshttp.NewRequest("GET", fullPath, nil)
	if err != nil {
		blog.V(3).Infof("http storage create watch request %s failed, %s", fullPath, err)
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		blog.V(3).Infof("http storage do watch request %s failed, %s", fullPath, err)
		return nil, err
	}
	// defer response.Body.Close()
	proxy := newHTTPWatch(fullPath, response, s.objectNewFn)
	go proxy.eventProxy()
	if selector != nil {
		blog.V(3).Infof("http storage client is ready to watch %s, selector %s", fullPath, selector.String())
	} else {
		blog.V(3).Infof("http storage client is ready to watch %s, selector null", fullPath)
	}
	return proxy, nil
}

// WatchList implements storage interface
// Watch & WatchList are the same for http api
func (s *Client) WatchList(ctx context.Context, key, version string, selector storage.Selector) (watch.Interface,
	error) {
	return s.Watch(ctx, key, version, selector)
}

// Get implements storage interface
// get exactly data object from http event storage. so key must be resource fullpath
// param cxt: not used
// param version: reserved for future
func (s *Client) Get(_ context.Context, key, _ string, ignoreNotFound bool) (meta.Object, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("lost object key")
	}
	if !strings.Contains(key, "cluster") {
		return nil, fmt.Errorf("lost cluster parameter")
	}
	if !strings.Contains(key, "namespace") {
		return nil, fmt.Errorf("lost namespace parameter")
	}
	if strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("err key format, no / ends")
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.selectServers(), defaultAPIPrefix, key)
	request, err := syshttp.NewRequest("GET", fullPath, nil)
	if err != nil {
		blog.V(3).Infof("http storage create GET %s failed, %s", fullPath, err)
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		blog.V(3).Infof("http storage Do %s request failed, %s", fullPath, err)
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)
	if response.StatusCode < syshttp.StatusOK || response.StatusCode >= syshttp.StatusMultipleChoices {
		blog.V(3).Infof("http storage get %s failed, code: %d, message: %s", fullPath, response.StatusCode, response.Status)
		return nil, fmt.Errorf("remote err, code: %d, status: %s", response.StatusCode, response.Status)
	}
	rawData, err := io.ReadAll(response.Body)
	if err != nil {
		blog.V(3).Infof("http storage get %s http status success, but read response body failed, %s", fullPath, err)
		return nil, err
	}
	// format http response
	standarResponse := &Response{}
	if err := json.Unmarshal(rawData, standarResponse); err != nil {
		blog.V(3).Infof("http storage decode GET %s http response failed, %s", fullPath, err)
		return nil, err
	}
	if standarResponse.Code != 0 {
		blog.V(3).Infof("http storage GET %s failed, %s", fullPath, standarResponse.Message)
		return nil, fmt.Errorf("remote err: %s", standarResponse.Message)
	}
	if len(standarResponse.Data) == 0 {
		blog.V(3).Infof("http storage GET %s success, but got no data", fullPath)
		if ignoreNotFound {
			return nil, nil
		}
		return nil, PrevDataErr
	}
	target := s.objectNewFn()
	if err := s.codec.Decode(standarResponse.Data, target); err != nil {
		blog.V(3).Infof("http storage decode data object %s failed, %s", fullPath, err)
		return nil, fmt.Errorf("json decode: %s", err)
	}
	blog.V(3).Infof("http storage client got %s success", fullPath)
	return target, nil
}

// List implements storage interface
// list namespace-based data or all data
func (s *Client) List(_ context.Context, key string, selector storage.Selector) ([]meta.Object, error) {
	if len(key) == 0 || strings.HasSuffix(key, "/") {
		return nil, fmt.Errorf("error key format")
	}
	fullPath := fmt.Sprintf("%s/%s/%s", s.selectServers(), defaultAPIPrefix, key)
	if selector != nil {
		filter := selector.String()
		if len(filter) != 0 {
			fullPath = fullPath + "?" + filter
		}
	}
	request, err := syshttp.NewRequest("GET", fullPath, nil)
	if err != nil {
		blog.V(3).Infof("http storage create requestfor GET %s failed, %s", fullPath, err)
		return nil, err
	}
	request.Header.Set("Content-type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		blog.V(3).Infof("http storage get %s failed, %s", fullPath, err)
		return nil, err
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)
	rawData, err := io.ReadAll(response.Body)
	if err != nil {
		blog.V(3).Infof("http storage read %s failed, %s", fullPath, err)
		return nil, err
	}
	standardResponse := &Response{}
	if jsonErr := json.Unmarshal(rawData, standardResponse); jsonErr != nil {
		blog.V(3).Infof("http storage decode %s response failed, %s", fullPath, jsonErr)
		return nil, jsonErr
	}
	if standardResponse.Code != 0 {
		blog.V(3).Infof("http storage List %s failed, %s", fullPath, err)
		return nil, fmt.Errorf("remote err, code: %d, %s", standardResponse.Code, standardResponse.Message)
	}
	if len(standardResponse.Data) == 0 {
		blog.V(3).Infof("http storage list empty data with %s", fullPath)
		return nil, nil
	}
	objs, err := s.objectListFn(standardResponse.Data)
	if err != nil {
		blog.V(3).Infof("http storage list %s success, but parse json list failed, %s", fullPath, err)
		return nil, err
	}
	blog.V(3).Infof("http storage list %s success, got %d objects", fullPath, len(objs))
	return objs, nil
}

// Close storage connection, clean resource
func (s *Client) Close() {
	blog.V(3).Infof("http api event storage %v exit.", s.servers)
}

func (s *Client) selectServers() string {
	index := rand.Intn(len(s.servers))
	return s.servers[index]
}

// newHTTPWatch create http watch
func newHTTPWatch(url string, response *syshttp.Response, objectFn meta.ObjectNewFn) *Watch {
	proxy := &Watch{
		url:           url,
		response:      response,
		objectNewFn:   objectFn,
		filterChannel: make(chan watch.Event, watch.DefaultChannelBuffer),
		isStop:        false,
	}
	return proxy
}

// Watch wrapper for http chunk response
type Watch struct {
	url           string
	response      *syshttp.Response
	objectNewFn   meta.ObjectNewFn
	filterChannel chan watch.Event
	isStop        bool
}

// Stop watch channel
func (e *Watch) Stop() {
	_ = e.response.Body.Close()
	e.isStop = true
}

// WatchEvent get watch events, if watch stopped/error, watch must close
// channel and exit, watch user must read channel like
// e, ok := <-channel
func (e *Watch) WatchEvent() <-chan watch.Event {
	return e.filterChannel
}

// eventProxy read all event json data from http response
// end then dispatch to use by Watch.Interface channel
func (e *Watch) eventProxy() {
	defer func() {
		close(e.filterChannel)
	}()
	buf := bufio.NewReader(e.response.Body)
	for {
		if e.isStop {
			blog.V(3).Infof("http watch is asked stopped")
			return
		}
		// reading all data from response connection
		rawStr, err := buf.ReadSlice('\n')
		if err != nil {
			blog.V(3).Infof("http watch %s read continue response failed, %s", e.url, err)
			return
		}
		// parse data
		watchRes := &WatchResponse{}
		if err := json.Unmarshal(rawStr, watchRes); err != nil {
			blog.V(3).Infof("http watch %s parse json %s failed, %s", string(rawStr), e.url, err)
			return
		}
		if watchRes.Code != 0 {
			// Note(DeveloperJim): error code classification for recovery
			blog.V(3).Infof("http watch %s failed, code: %d, message: %s", e.url, watchRes.Code, watchRes.Message)
			return
		}
		target := e.objectNewFn()
		if err := json.Unmarshal(watchRes.Data.Data, target); err != nil {
			blog.V(3).Infof("http watch %s got unexpect json parsing err, %s", e.url, err)
			return
		}
		targetEvent := watch.Event{
			Type: watchRes.Data.Type,
			Data: target,
		}
		e.filterChannel <- targetEvent
	}
}
