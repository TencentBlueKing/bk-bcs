/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/gse"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/metrics"
	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules/session"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
	"bk-bscp/pkg/ssl"
)

const (
	// defaultReInitWaitDuration is default re-init wait duration.
	defaultReInitWaitDuration = time.Second
)

// GSEClient is gse taskserver client.
type GSEClient struct {
	// config viper as context here.
	viper *viper.Viper

	// serviceID is given by gse admin.
	serviceID uint32

	// endpoint is gse taskserver endpoint.
	endpoint string

	// sessionID is gse taskserver session id.
	sessionID uint64

	// plat service protocol factory.
	platService *gse.PlatService

	// gse taskserver client.
	conn     *tls.Conn
	mu       sync.Mutex
	isClosed bool

	// plugin send message channel.
	sendMessageChan chan *gse.Message

	// plugin recv message channel.
	recvMessageChan chan *gse.Message

	// session manager, handles gse plugin sessions.
	sessionMgr *session.Manager

	// gseProcesser is gse message processer.
	gseProcesser *GSEProcesser

	// tunnelserver handles normal request from plugin.
	tunnelServer *TunnelServer

	// prometheus metrics collector.
	collector *metrics.Collector
}

// NewGSEClient creates a new GSEClient.
func NewGSEClient(viper *viper.Viper, tunnelServer *TunnelServer, serviceID uint32, endpoint string) *GSEClient {
	client := &GSEClient{
		viper:           viper,
		serviceID:       serviceID,
		endpoint:        endpoint,
		platService:     gse.NewPlatService(),
		sendMessageChan: make(chan *gse.Message, viper.GetInt("gseTaskServer.sendMessageChanSize")),
		recvMessageChan: make(chan *gse.Message, viper.GetInt("gseTaskServer.recvMessageChanSize")),
		sessionMgr:      tunnelServer.sessionMgr,
		tunnelServer:    tunnelServer,
		collector:       tunnelServer.collector,
	}

	client.gseProcesser = NewGSEProcesser(viper, client.platService, tunnelServer, client.pushToPlugin)
	client.gseProcesser.Run()

	return client
}

// Init inits a new GSEClient.
func (c *GSEClient) Init() error {
	// init GSE plat service.
	if err := c.newGSEPlatService(); err != nil {
		return err
	}

	// keep processing send message.
	for i := 0; i < c.viper.GetInt("gseTaskServer.processerNum"); i++ {
		go c.processSend()
	}

	// keep processing recv message.
	for i := 0; i < c.viper.GetInt("gseTaskServer.processerNum"); i++ {
		go c.processRecv()
	}

	// keep recving message, and handle request to target channels.
	go c.process()

	return nil
}

func (c *GSEClient) register() error {
	header := &gse.PlatServiceMessageHeader{
		MainHeader: &gse.PlatServiceMessageMainHeader{
			MessageType: gse.PlatServiceCMDRegisterReq,
		},
		ExtraHeader: &gse.PlatServiceMessageExtraHeader{
			ServiceID:         c.serviceID,
			SessionID:         c.sessionID,
			MessageSequenceID: common.SequenceNum(),
		},
	}

	message, err := c.platService.Encode(header, nil)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(message); err != nil {
		return err
	}

	respHeader, body, err := c.platService.DecodeHeader(c.conn)
	if err != nil {
		return err
	}
	if respHeader.MainHeader.MessageType != gse.PlatServiceCMDRegisterResp {
		return fmt.Errorf("register failed, response message type[%d] not matched", respHeader.MainHeader.MessageType)
	}

	respBody := &gse.PlatServiceSimpleBody{}
	if err := c.platService.DecodeJSONBody(body, respBody); err != nil {
		return fmt.Errorf("decode register response message body failed, %+v", err)
	}
	if respBody.ErrCode != gse.PlatServiceErrCodeOK {
		return fmt.Errorf("register failed, response errcode: %d, errmsg: %+v", respBody.ErrCode, respBody.ErrMsg)
	}

	// reset session id for the new connection.
	c.sessionID = respHeader.ExtraHeader.SessionID

	return nil
}

// build or re-build connection.
func (c *GSEClient) newGSEPlatService() error {
	// TLS.
	tlsConf, err := ssl.ClientTLSConfVerify(
		c.viper.GetString("gseTaskServer.tls.caFile"),
		c.viper.GetString("gseTaskServer.tls.certFile"),
		c.viper.GetString("gseTaskServer.tls.keyFile"),
		c.viper.GetString("gseTaskServer.tls.certPassword"))
	if err != nil {
		return fmt.Errorf("setup gse taskserver client TLS failed, %+v", err)
	}

	setTCPConfig := func(clientHello *tls.ClientHelloInfo) (*tls.Config, error) {
		tcpConn, ok := clientHello.Conn.(*net.TCPConn)
		if !ok {
			return nil, errors.New("can't set tcp config, this is not TCP protocol")
		}
		// tcp buffer.
		tcpConn.SetWriteBuffer(c.viper.GetInt("gseTaskServer.writeBufferSize"))
		tcpConn.SetReadBuffer(c.viper.GetInt("gseTaskServer.readBufferSize"))

		// keepalive.
		tcpConn.SetKeepAlive(c.viper.GetDuration("gseTaskServer.keepAlivePeriod") != 0)
		tcpConn.SetKeepAlivePeriod(c.viper.GetDuration("gseTaskServer.keepAlivePeriod"))
		return nil, nil
	}
	tlsConf.GetConfigForClient = setTCPConfig

	// connect.
	conn, err := tls.Dial("tcp", c.endpoint, tlsConf)
	if err != nil {
		return err
	}

	if c.conn != nil {
		// close connection for re-connect, GSE would unregister session for
		// this connection automatically.
		c.conn.Close()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// reset new connection.
	c.conn = conn

	// register session for new connection.
	if err := c.register(); err != nil {
		return err
	}
	logger.V(2).Infof("GSECLIENT sessionid[%d]| register session for new connection success", c.sessionID)

	return nil
}

// close old conn/fd to re-connect, and then register session again.
func (c *GSEClient) reconnect() {
	for {
		err := c.newGSEPlatService()
		if err == nil {
			break
		}

		logger.Warnf("GSECLIENT sessionid[%d]| re-connect, %+v", c.sessionID, err)
		c.collector.StatGSEPlatServiceReInit(false)

		// do not reconnect too frequently.
		time.Sleep(defaultReInitWaitDuration)
	}

	logger.Warnf("GSECLIENT sessionid[%d]| re-connect success!", c.sessionID)
	c.collector.StatGSEPlatServiceReInit(true)
}

func (c *GSEClient) processRecv() {
	startTime := time.Now()

	for {
		if c.isClosed {
			return
		}
		c.collector.StatProcess("recv", startTime, time.Now())

		message := <-c.recvMessageChan

		startTime = time.Now()

		// stat message recv processed num.
		c.collector.StatGSEPlatMessageProcessedCount("recv")

		// process recved message.
		c.gseProcesser.Handle(message)

		// send message response ack to plat service.
		respHeader := &gse.PlatServiceMessageHeader{
			MainHeader: &gse.PlatServiceMessageMainHeader{
				MessageType: gse.PlatServiceCMDSubscribeFromPluginResp,
			},
			ExtraHeader: &gse.PlatServiceMessageExtraHeader{
				ServiceID:         c.serviceID,
				SessionID:         c.sessionID,
				MessageSequenceID: message.Header.ExtraHeader.MessageSequenceID,
			},
		}

		respBody, err := c.platService.EncodeJSONBody(&gse.PlatServiceSimpleBody{
			ErrCode: gse.PlatServiceErrCodeOK,
			ErrMsg:  gse.PlatServiceErrMsgSuccess,
		})
		if err != nil {
			logger.Errorf("GSECLIENT sessionid[%d]| encode plugin message response failed, %+v, %+v, %+v",
				c.sessionID, message.Header.MainHeader, message.Header.ExtraHeader, err)
			continue
		}

		if err := c.addMessageToSendChan(&gse.Message{Header: respHeader, Body: respBody}); err != nil {
			logger.Warnf("GSECLIENT sessionid[%d]| plugin ack message to chan fusing now, %+v, %+v, len[%d], %+v",
				c.sessionID, respHeader.MainHeader, respHeader.ExtraHeader, len(respBody), err)
		}
	}
}

func (c *GSEClient) addMessageToRecvChan(message *gse.Message) error {
	// stat recv message channel runtime size.
	c.collector.StatGSEPlatMessageChanRuntime("recv", int64(len(c.recvMessageChan)))

	select {
	case c.recvMessageChan <- message:
		c.collector.StatGSEPlatMessageCount("recv")

	case <-time.After(c.viper.GetDuration("gseTaskServer.recvMessageChanTimeout")):
		c.collector.StatGSEPlatMessageFuseCount("recv")
		return fmt.Errorf("add to gse recv channel timeout, current size[%d]", len(c.recvMessageChan))
	}
	return nil
}

func (c *GSEClient) addMessageToSendChan(message *gse.Message) error {
	// stat send message channel runtime size.
	c.collector.StatGSEPlatMessageChanRuntime("send", int64(len(c.sendMessageChan)))

	select {
	case c.sendMessageChan <- message:
		c.collector.StatGSEPlatMessageCount("send")

	case <-time.After(c.viper.GetDuration("gseTaskServer.sendMessageChanTimeout")):
		c.collector.StatGSEPlatMessageFuseCount("send")
		return fmt.Errorf("add to gse send channel timeout, current size[%d]", len(c.sendMessageChan))
	}
	return nil
}

func (c *GSEClient) handlePushMessage(message *gse.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// reset new session id.
	message.Header.ExtraHeader.SessionID = c.sessionID

	// encode push message.
	pushMessage, err := c.platService.Encode(message.Header, message.Body)
	if err != nil {
		return err
	}
	if _, err := c.conn.Write(pushMessage); err != nil {
		return err
	}
	return nil
}

func (c *GSEClient) processSend() {
	startTime := time.Now()

	for {
		if c.isClosed {
			return
		}
		c.collector.StatProcess("send", startTime, time.Now())

		message := <-c.sendMessageChan

		startTime = time.Now()

		// stat message send processed num.
		c.collector.StatGSEPlatMessageProcessedCount("send")

		logger.V(2).Infof("GSECLIENT sessionid[%d]| send new message, %+v %+v, len[%d]",
			c.sessionID, message.Header.MainHeader, message.Header.ExtraHeader, len(message.Body))

		if err := c.handlePushMessage(message); err != nil {
			logger.Errorf("GSECLIENT sessionid[%d]| handle push response message, %+v, %+v, %+v",
				c.sessionID, message.Header.MainHeader, message.Header.ExtraHeader, err)
		}
	}
}

// keep processing recv message.
func (c *GSEClient) process() error {
	startTime := time.Now()

	for {
		if c.isClosed {
			return nil
		}

		c.collector.StatProcessProtocol(fmt.Sprintf("%d", c.sessionID), startTime, time.Now())
		startTime = time.Now()

		header, body, err := c.platService.DecodeHeader(c.conn)
		if err != nil {
			if err == io.EOF {
				// connection reset by peer, need re-connect.
				logger.Warnf("GSECLIENT sessionid[%d]| connection reset by peer, %+s",
					c.sessionID, c.conn.RemoteAddr().String())
			} else if err == syscall.EINVAL {
				// not connected, need re-connect.
				logger.Warnf("GSECLIENT sessionid[%d]| connection missing, re-connect now", c.sessionID)
			} else {
				// decode message stream failed, re-connect to recv matched package length.
				logger.Errorf("GSECLIENT sessionid[%d]| process recv and decode header failed, %+v", c.sessionID, err)
			}

			// re-connect.
			c.reconnect()
			continue
		}

		logger.V(4).Infof("GSECLIENT sessionid[%d]| recv new message, %+v %+v, len[%d]",
			c.sessionID, header.MainHeader, header.ExtraHeader, len(body))

		if header.ExtraHeader.ServiceID != c.serviceID {
			logger.Errorf("GSECLIENT sessionid[%d]| recv not matched service id[%d]",
				c.sessionID, header.ExtraHeader.ServiceID)
			continue
		}
		if header.ExtraHeader.SessionID != c.sessionID {
			logger.Errorf("GSECLIENT sessionid[%d]| recv not matched session id[%d]",
				c.sessionID, header.ExtraHeader.SessionID)
			continue
		}
		if err := c.addMessageToRecvChan(&gse.Message{Header: header, Body: body}); err != nil {
			logger.Warnf("GSECLIENT sessionid[%d]| recv message channel fusing now, %+v, %+v, len[%d], %+v",
				c.sessionID, header.MainHeader, header.ExtraHeader, len(body), err)
		}
	}
}

func (c *GSEClient) pushToPlugin(sendProcesserMessage *modules.GSESendProcesserMessage) error {
	data, err := proto.Marshal(sendProcesserMessage.UpStream)
	if err != nil {
		return err
	}

	message := &gse.PlatServicePushToPluginBody{
		Agents:  sendProcesserMessage.Agents,
		Message: base64.StdEncoding.EncodeToString(data),
	}

	body, err := c.platService.EncodeJSONBody(message)
	if err != nil {
		return err
	}

	header := &gse.PlatServiceMessageHeader{
		MainHeader: &gse.PlatServiceMessageMainHeader{
			MessageType: gse.PlatServiceCMDPushToPluginReq,
		},
		ExtraHeader: &gse.PlatServiceMessageExtraHeader{
			ServiceID:         c.serviceID,
			SessionID:         c.sessionID,
			MessageSequenceID: common.SequenceNum(),
		},
	}

	if err := c.addMessageToSendChan(&gse.Message{Header: header, Body: body}); err != nil {
		logger.Warnf("GSECLIENT sessionid[%d]| send message channel fusing now, %+v, %+v, len[%d], %+v",
			c.sessionID, header.MainHeader, header.ExtraHeader, len(body), err)
		return err
	}

	return nil
}

// Close closes plat service client.
func (c *GSEClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}
	c.isClosed = true
}

// GSEClientManager is gse client manager.
type GSEClientManager struct {
	// config viper as context here.
	viper *viper.Viper

	// serviceID is gse service id.
	serviceID uint32

	// handle normal request from plugin.
	tunnelServer *TunnelServer

	// gse target servers endpoints.
	endpoints []string

	// gse task server clients.
	gseClients []*GSEClient
}

// NewGSEClientManager creates a new GSEClientManager.
func NewGSEClientManager(viper *viper.Viper, tunnelServer *TunnelServer) *GSEClientManager {
	return &GSEClientManager{
		viper:        viper,
		tunnelServer: tunnelServer,
		serviceID:    viper.GetUint32("gseTaskServer.gseServiceID"),
		endpoints:    viper.GetStringSlice("gseTaskServer.endpoints"),
	}
}

// Init inits new gse client manager.
func (mgr *GSEClientManager) Init() error {
	if len(mgr.endpoints) == 0 {
		return errors.New("empty endpoints")
	}

	for _, endpoint := range mgr.endpoints {
		gseClient := NewGSEClient(mgr.viper, mgr.tunnelServer, mgr.serviceID, endpoint)
		if err := gseClient.Init(); err != nil {
			return err
		}
		mgr.gseClients = append(mgr.gseClients, gseClient)
	}

	return nil
}

// Close closes target gse client manager.
func (mgr *GSEClientManager) Close() {
	for _, gseClient := range mgr.gseClients {
		gseClient.Close()
	}
}
