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

package connection

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	exe "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/mesos/executor"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/upid"
)

// Message defines the type that passes in the ExecutorDriver.
type Message struct {
	UPID  *upid.UPID //remote server
	URI   string     //remote uri
	Name  string     //message name
	Bytes []byte     //json bytes
}

//HandleFunc handlefunc is call back for receiving message
//from: Message from server
//msg: receiving message
type HandleFunc func(from *upid.UPID, msg proto.Message)

//DispatchFunc dispatch event message to HandleFunc
type DispatchFunc func(from *upid.UPID, event *exe.Event)

const (
	//Event_CONNECTION_CLOSE connection close event type
	Event_CONNECTION_CLOSE exe.Event_Type = 1024
)

//Status status for Connection
type Status int

const (
	//NOTCONNECT init status for Connection
	NOTCONNECT Status = iota
	//DISCONNECT if connection is broken, set to DISCONNECT
	DISCONNECT
	//CONNECTED connection is OK
	CONNECTED
	//STOPPED connection is stop
	STOPPED
)

//CONNKEEPALIVE only for timeout for http connection, one year for timeout
//const CONNKEEPALIVE = 86400 * 365 * time.Second

//Connection maintain http connection with Mesos slave
type Connection interface {
	//Install mount an handler based on incoming message name.
	Install(eventType exe.Event_Type, dispatch DispatchFunc) error

	//TLSConfig setting TLS Connection to remote server
	TLSConfig(config *tls.Config, handshakeTimeout time.Duration)

	//Start starts the Connection to remote server.
	Start(endpoint string, path string) error

	//Stop kills the transporter.
	Stop(graceful bool)

	//Send sends message to remote process. Will stop sending when Connection is stopped.
	//if aliveLoop is true, keep connection alive for incomming package
	Send(call *exe.Call, aliveLoop bool) error

	//Recv receives message initialtively
	Recv() (*exe.Event, error)

	//GetConnStatus return connection status
	GetConnStatus() Status
}

//NewConnection return Connection
func NewConnection() Connection {
	//http transport
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: 5 * time.Second,
		//DisableKeepAlives:true,
	}
	return &HTTPConnection{
		transport: httpTransport,
		client: &http.Client{
			//Timeout:   CONNKEEPALIVE,
			Transport: httpTransport,
		},
		status:        NOTCONNECT,
		dispatcherMap: make(map[exe.Event_Type]DispatchFunc),
	}
}

//HTTPConnection implement Connection interface communicate
//mesos slave
type HTTPConnection struct {
	endpoint      string                          //remote http endpoint info
	uri           string                          //remote http endpoint uri
	streamID      string                          //http header Mesos-Stream-Id
	protocol      string                          //https or http
	transport     *http.Transport                 //connection transport
	client        *http.Client                    //http client for connection
	status        Status                          //Connection status
	dispatchLock  sync.RWMutex                    //lock dispatchMap read & write
	dispatcherMap map[exe.Event_Type]DispatchFunc //store for DispatchFunc
}

//Install mount an handler based on incoming message name.
func (httpConn *HTTPConnection) Install(eventType exe.Event_Type, dispatch DispatchFunc) error {
	//register handler with message name
	httpConn.dispatchLock.Lock()
	defer httpConn.dispatchLock.Unlock()
	if _, exist := httpConn.dispatcherMap[eventType]; exist {
		return fmt.Errorf("Message %s is already installed", eventType.String())
	}
	httpConn.dispatcherMap[eventType] = dispatch
	return nil
}

//TLSConfig setting TLS Connection to remote server
func (httpConn *HTTPConnection) TLSConfig(config *tls.Config, handshakeTimeout time.Duration) {
	//https transport
	httpsTransport := &http.Transport{
		TLSHandshakeTimeout: handshakeTimeout,
		TLSClientConfig:     config,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: 5 * time.Second,
		//DisableKeepAlives:true,
	}
	//refresh http transport & client
	httpConn.transport = httpsTransport
	httpConn.client = &http.Client{
		//Timeout:   CONNKEEPALIVE,
		Transport: httpsTransport,
	}
}

//Start starts the Connection to remote server.
//endpoint: remote endpoint url info, like mesos-slave:8080, mesos-slave:8080
//path: remote, like /v1/executor
func (httpConn *HTTPConnection) Start(endpoint string, path string) error {
	httpConn.endpoint = endpoint
	httpConn.uri = path
	return nil
}

//Stop kills the transporter.
func (httpConn *HTTPConnection) Stop(graceful bool) {
	fmt.Fprintln(os.Stdout, "HttpConnection is asked to stop")
	httpConn.status = STOPPED
}

//Send sends message to remote process. Will stop sending when Connection is stopped.
//if aliveLoop is true, keep connection alive for incomming package,
//when new message coming, dispatch message to installed callback handlerFunc
func (httpConn *HTTPConnection) Send(call *exe.Call, keepAlive bool) error {
	if httpConn.status == CONNECTED && keepAlive {
		//only one long connection is permitted
		fmt.Fprintf(os.Stderr, "Only one long connection with one HTTPConnection")
		return fmt.Errorf("long connection is exist")
	}
	//create targetURL
	targetURL := fmt.Sprintf("%s%s", httpConn.endpoint, httpConn.uri)
	//json serialization
	/*
		jsonBytes, jsonErr := json.Marshal(call)
		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Format executor.Call to json failed: %s", jsonErr.Error())
			return jsonErr
		}
	*/
	// proto serialization
	payLoad, payLoadErr := proto.Marshal(call)
	if payLoadErr != nil {
		fmt.Fprintf(os.Stderr, "Format executor.Call to proto failed: %s", payLoadErr.Error())
		return payLoadErr
	}

	//create http request
	request, err := http.NewRequest("POST", targetURL, bytes.NewReader(payLoad))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create Post request to %s failed: %s", targetURL, err.Error())
		return err
	}

	request.Header.Set("Content-Type", "application/x-protobuf")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "bcs-container-executor/1.0")
	//request.Header.Set("Connection", "Close")
	//if keepAlive {
	//starting long connection for incoming message
	request.Header.Set("Connection", "Keep-Alive")
	//}
	response, resErr := httpConn.client.Do(request)
	if resErr != nil {
		fmt.Fprintf(os.Stderr, "POST TO %s failed: %s\n", targetURL, resErr.Error())
		return resErr
	}
	if !(response.StatusCode == http.StatusAccepted || response.StatusCode == http.StatusOK) {
		reply, _ := ioutil.ReadAll(response.Body)
		return fmt.Errorf("Connect to %s failed, [%d]%s, reply(%s)", targetURL, response.StatusCode, response.Status, string(reply))
	}
	//fmt.Fprintf(os.Stdout, "POST to %s %s message success, code: %s\n", targetURL, call.GetType().String(), response.Status)
	if keepAlive {
		httpConn.status = CONNECTED
		go httpConn.recvLoop(response)
		return nil
	}
	defer response.Body.Close()
	//httpConn.status = DISCONNECT
	return nil
}

//Recv receives message initialtively
func (httpConn *HTTPConnection) Recv() (*exe.Event, error) {
	return nil, fmt.Errorf("NotImplement")
}

//GetConnStatus return connection status
func (httpConn *HTTPConnection) GetConnStatus() Status {
	return httpConn.status
}

//reconnect if connection to mesos slave down, reconnect to slave
func (httpConn *HTTPConnection) reconnect() error {
	return fmt.Errorf("NotImplement")
}

//recvLoop recv message from mesos & post to DispatchFunc
func (httpConn *HTTPConnection) recvLoop(response *http.Response) {
	defer response.Body.Close()
	jsonDecoder := json.NewDecoder(NewReader(response.Body))
	for {
		if httpConn.status == STOPPED {
			fmt.Fprintln(os.Stdout, "HTTPConnection is stopped, skip recvLoop")
			return
		}
		if httpConn.status == DISCONNECT {
			//done in 2017-01-10(developerJim): connection broken, notify ExecutorDriver for reconnect
			//this recvLoop must return for releasing response.Body resource
			fmt.Fprintln(os.Stderr, "HTTPConnection is disconnected, recvLoop exit, ready to reconnect")
			httpConn.dispatchLock.RLock()
			defer httpConn.dispatchLock.RUnlock()
			dispatchFunc, ok := httpConn.dispatcherMap[Event_CONNECTION_CLOSE]
			if ok {
				go dispatchFunc(nil, nil)
			} else {
				fmt.Fprintln(os.Stderr, "HTTPConnection do not register message CONNECTION_CLOSE HandlerFunc")
			}
			return
		}
		event := new(exe.Event)
		if err := jsonDecoder.Decode(event); err != nil {
			//no matter what error is, may be decode json error, EOF,
			//http connection timeout, we set status to DISCONNECTED,
			//and try reconnect later.
			fmt.Fprintf(os.Stderr, "HTTPConnection decode failed: %s\n", err.Error())
			httpConn.status = DISCONNECT
		}
		//check if decode object. if error is EOF,
		//event may be decode successfully
		if event.GetType() != exe.Event_UNKNOWN {
			httpConn.dispatchLock.RLock()
			eType := event.GetType()
			dispatchFunc, ok := httpConn.dispatcherMap[eType]
			if ok {
				go dispatchFunc(nil, event)
			} else {
				fmt.Fprintf(os.Stderr, "HTTPConnection do not register message %s HandlerFunc\n", eType.String())
			}
			httpConn.dispatchLock.RUnlock()
		}
	}
}
