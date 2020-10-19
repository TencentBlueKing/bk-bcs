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
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	exe "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/mesos/executor"

	"github.com/golang/protobuf/proto"
)

//NewFakeConnection return Connection
func NewFakeConnection() Connection {
	c := NewConnection()
	return &FakeConnection{
		con:        c.(*HTTPConnection),
		stopCh:     make(chan struct{}, 1),
		dataReader: nil,
		dataWriter: nil,
		dataCh:     make(chan *exe.Call, 10),
	}
}

//FakeConnection is a proxy Connection for testing
//Executor message register, runTaskGroup, shutdown
type FakeConnection struct {
	con        *HTTPConnection //connection for HTTPConenction
	stopCh     chan struct{}   //stop channel for data loop
	dataReader *io.PipeReader  //reader for http.Response
	dataWriter *io.PipeWriter  //writer for reader
	dataCh     chan *exe.Call  //data channel for handling message sent by Executor
}

//Install mount an handler based on incoming message name.
func (httpConn *FakeConnection) Install(eventType exe.Event_Type, dispatch DispatchFunc) error {
	return httpConn.con.Install(eventType, dispatch)
}

//TLSConfig setting TLS Connection to remote server
func (httpConn *FakeConnection) TLSConfig(config *tls.Config, handshakeTimeout time.Duration) {
	//https transport
	httpConn.con.TLSConfig(config, handshakeTimeout)
}

//Start starts the Connection to remote server.
//endpoint: remote endpoint url info, like mesos-slave:8080, mesos-slave:8080
//path: remote, like /v1/executor
func (httpConn *FakeConnection) Start(endpoint string, path string) error {
	httpConn.con.Start(endpoint, path)
	//start event message simulation
	go httpConn.messageSimulation()
	//active test message sending
	go httpConn.slaveSimulation()
	return nil
}

func (httpConn *FakeConnection) slaveSimulation() {
	//simulate slave, sending message like TaskGroupInfo,
	//Shutdown, KillTask to Executor
	//send TaskGroupInfo in 1 second
	//send Shutdown in 10 second
	//send KillTask in 10 second ?
	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	fmt.Fprintln(os.Stdout, "enter slave message sending loop")
	i := 0
	for {
		select {
		case <-httpConn.stopCh:
			fmt.Fprintln(os.Stdout, "FakeConnection slave simulation exit")
			return
		case <-tick.C:
			fmt.Fprintf(os.Stdout, "slaveSimulation tick: %s\n", time.Now())
			if i == 1 {
				response := httpConn.getTaskGroupInfoResponse()
				httpConn.recvLoop(response)
			}
			if i == 20 {
				response := httpConn.getShutdownResponse()
				httpConn.recvLoop(response)
			}
			i++
		}
	}
}

func (httpConn *FakeConnection) messageSimulation() {
	fmt.Fprintln(os.Stdout, "enter Mesos message handling loop")
	for {
		select {
		case <-httpConn.stopCh:
			fmt.Fprintln(os.Stdout, "FakeConnection message simulation exit")
			return
		case event := <-httpConn.dataCh:
			//handle message according EventType
			var response *http.Response
			switch event.GetType() {
			case exe.Call_SUBSCRIBE:
				response = httpConn.getRegisteredResponse(event)
			case exe.Call_UPDATE:
				response = httpConn.getUpdateResponse(event)
			case exe.Call_MESSAGE:
				response = httpConn.getMessageResponse(event)
			}
			httpConn.recvLoop(response)
		}
	}
}

func (httpConn *FakeConnection) getRegisteredResponse(call *exe.Call) *http.Response {
	//response registered event
	//create data for json
	event := &exe.Event{
		Type: exe.Event_SUBSCRIBED.Enum(),
		Subscribed: &exe.Event_Subscribed{
			ExecutorInfo: &mesos.ExecutorInfo{
				Type:       mesos.ExecutorInfo_CUSTOM.Enum(),
				ExecutorId: call.GetExecutorId(),
			},
			FrameworkInfo: &mesos.FrameworkInfo{
				Id: call.GetFrameworkId(),
			},
			AgentInfo: &mesos.AgentInfo{
				Hostname: proto.String("developerJimTestAgent"),
				Id:       &mesos.AgentID{Value: proto.String("developerJim-testing-agentID")},
			},
		},
	}
	response := new(http.Response)
	response.Status = "200 OK"
	response.StatusCode = http.StatusOK
	response.Header = make(map[string][]string)
	response.Header.Set("Content-Type", "application/json")
	resBytes, _ := json.Marshal(event)
	length := len(resBytes)
	response.Header.Set("Transfer-Encoding", "chunked")
	if httpConn.dataWriter != nil {
		httpConn.dataWriter.Close()
		httpConn.dataReader.Close()
	}
	httpConn.dataReader, httpConn.dataWriter = io.Pipe()
	go func() {
		httpConn.dataWriter.Write([]byte(strconv.Itoa(length) + "\n"))
		httpConn.dataWriter.Write(resBytes)
	}()
	fmt.Fprintf(os.Stdout, "Register message created.\n")
	response.Body = ioutil.NopCloser(httpConn.dataReader)
	return response
}

func (httpConn *FakeConnection) getTaskGroupInfoResponse() *http.Response {
	fmt.Fprintln(os.Stdout, "ready to create task group info send to driver")
	var tasks []*mesos.TaskInfo
	//create task for tasks
	task := &mesos.TaskInfo{
		Name:    proto.String("developerJimTestTaskBox"),
		TaskId:  &mesos.TaskID{Value: proto.String("d40f3f3e-bbe3-44af-a230-4cb1eae72f67")},
		AgentId: &mesos.AgentID{Value: proto.String("developerJim-testing-agentID")},
		Command: &mesos.CommandInfo{
			Shell:     proto.Bool(false),
			Value:     proto.String("/bin/sh"),
			Arguments: []string{"-c", "while true; do echo hello developerJim; sleep 20; done"},
		},
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_MESOS.Enum(),
			Mesos: &mesos.ContainerInfo_MesosInfo{
				Image: &mesos.Image{
					Type: mesos.Image_DOCKER.Enum(),
					Docker: &mesos.Image_Docker{
						Name: proto.String("hub.o.com/system/busybox"),
					},
				},
			},
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image:          proto.String("hub.o.com/calico/busybox"),
				ForcePullImage: proto.Bool(true),
				PortMappings:   []*mesos.ContainerInfo_DockerInfo_PortMapping{},
				Parameters:     []*mesos.Parameter{},
			},
		},
	}
	tasks = append(tasks, task)
	event := &exe.Event{
		Type: exe.Event_LAUNCH_GROUP.Enum(),
		LaunchGroup: &exe.Event_LaunchGroup{
			TaskGroup: &mesos.TaskGroupInfo{
				Tasks: tasks,
			},
		},
	}
	response := new(http.Response)
	response.Status = "200 OK"
	response.StatusCode = http.StatusOK
	response.Header = make(map[string][]string)
	response.Header.Set("Content-Type", "application/json")
	resBytes, _ := json.Marshal(event)
	length := len(resBytes)
	lengthBytes := strconv.Itoa(length)
	var allbytes []byte
	allbytes = append(allbytes, []byte(lengthBytes+"\n")...)
	allbytes = append(allbytes, resBytes...)
	response.Header.Set("Content-Length", strconv.Itoa(len(allbytes)))
	response.Body = ioutil.NopCloser(bytes.NewReader(allbytes))
	if httpConn.dataWriter != nil {
		fmt.Fprintf(os.Stdout, "taskgroup message created.")
		go httpConn.dataWriter.Write(allbytes)
		fmt.Fprintf(os.Stdout, "taskgroup message sending succ.")
		return nil
	}
	return response
}

func (httpConn *FakeConnection) getShutdownResponse() *http.Response {
	event := &exe.Event{
		Type: exe.Event_SHUTDOWN.Enum(),
	}
	response := new(http.Response)
	response.Status = "200 OK"
	response.StatusCode = http.StatusOK
	response.Header = make(map[string][]string)
	response.Header.Set("Content-Type", "application/json")
	resBytes, _ := json.Marshal(event)
	length := len(resBytes)
	lengthBytes := strconv.Itoa(length)
	var allbytes []byte
	allbytes = append(allbytes, []byte(lengthBytes+"\n")...)
	allbytes = append(allbytes, resBytes...)
	response.Header.Set("Content-Length", strconv.Itoa(len(allbytes)))
	response.Body = ioutil.NopCloser(bytes.NewReader(allbytes))
	if httpConn.dataWriter != nil {
		go func() {
			httpConn.dataWriter.Write(allbytes)
			httpConn.dataWriter.Close()
		}()
		return nil
	}
	return response
}

func (httpConn *FakeConnection) getUpdateResponse(call *exe.Call) *http.Response {
	event := &exe.Event{
		Type: exe.Event_ACKNOWLEDGED.Enum(),
		Acknowledged: &exe.Event_Acknowledged{
			TaskId: call.Update.Status.TaskId,
			Uuid:   call.Update.Status.Uuid,
		},
	}
	response := new(http.Response)
	response.Status = "200 OK"
	response.StatusCode = http.StatusOK
	response.Header = make(map[string][]string)
	response.Header.Set("Content-Type", "application/json")
	resBytes, _ := json.Marshal(event)
	length := len(resBytes)
	response.Header.Set("Content-Length", strconv.Itoa(length))
	response.Body = ioutil.NopCloser(bytes.NewReader(resBytes))
	if httpConn.dataWriter != nil {
		go func() {
			httpConn.dataWriter.Write([]byte(strconv.Itoa(length) + "\n"))
			httpConn.dataWriter.Write(resBytes)
		}()
		return nil
	}
	return response
}

func (httpConn *FakeConnection) getMessageResponse(call *exe.Call) *http.Response {
	return nil
}

//Stop kills the transporter.
func (httpConn *FakeConnection) Stop(graceful bool) {
	fmt.Fprintln(os.Stdout, "FakeConnection stop is Called")
	httpConn.con.Stop(graceful)
	close(httpConn.stopCh)
}

//Send sends message to remote process. Will stop sending when Connection is stopped.
//if aliveLoop is true, keep connection alive for incomming package,
//when new message coming, dispatch message to installed callback handlerFunc
func (httpConn *FakeConnection) Send(call *exe.Call, keepAlive bool) error {
	if call != nil {
		httpConn.dataCh <- call
	}
	return nil
}

//Recv receives message initialtively
func (httpConn *FakeConnection) Recv() (*exe.Event, error) {
	return nil, fmt.Errorf("NotImplement")
}

//GetConnStatus return connection status
func (httpConn *FakeConnection) GetConnStatus() Status {
	return httpConn.con.status
}

//reconnect if connection to mesos slave down, reconnect to slave
func (httpConn *FakeConnection) reconnect() error {
	return fmt.Errorf("NotImplement")
}

//recvLoop recv message from mesos & post to HandleFunc
func (httpConn *FakeConnection) recvLoop(response *http.Response) {
	if response == nil {
		fmt.Fprintln(os.Stdout, "Get Nil Response")
		return
	}
	//setting connected before handle response
	httpConn.con.status = CONNECTED
	go httpConn.con.recvLoop(response)
}
