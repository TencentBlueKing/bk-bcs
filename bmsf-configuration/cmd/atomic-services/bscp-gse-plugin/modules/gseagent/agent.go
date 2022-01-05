/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.  */

package gseagent

import "C"
import (
	"errors"
	"time"

	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/logger"
)

const (
	// default message channel size.
	defaultMessageChannelSize = 100
)

// Message message on gse tunnel.
type Message struct {
	// Meta is agent message meta.
	Meta []byte

	// Data us agent message data.
	Data []byte
}

// GSEAgent is gse agent wrapper.
type GSEAgent struct {
	viper *safeviper.SafeViper

	// recvFunc is message recv handle func.
	recvFunc func(*Message)

	// messageChannel is message channel.
	messageChannel chan *Message

	// stop channel.
	stopCh chan struct{}
}

// NewGSEAgent creates a new GSEAgent object.
func NewGSEAgent(viper *safeviper.SafeViper) *GSEAgent {
	return &GSEAgent{
		viper:          viper,
		messageChannel: make(chan *Message, defaultMessageChannelSize),
		stopCh:         make(chan struct{}),
	}
}

// SendMessage sends message base on gse tunnel.
func (agent *GSEAgent) SendMessage(data []byte, sessionID int64, messageID uint64, transmitType TransmitType) int {
	return SendMessage(string(data), sessionID, int64(messageID), transmitType)
}

// Run inits gse agent and keep recving message here..
func (agent *GSEAgent) Run(recvFunc func(*Message)) error {
	if recvFunc == nil {
		return errors.New("can't register recv func: nil")
	}

	// registers func for recving message.
	agent.recvFunc = recvFunc

	// recving in a loop.
	go agent.recvLoop()

	logger.Infof("GSE Agent| init success and keep recving message now!")
	return nil
}

// RecvMessage recvs message fromgse agent.
func (agent *GSEAgent) RecvMessage(msg *Message) {
	select {
	case agent.messageChannel <- msg:
	case <-time.After(time.Second):
		logger.Warnf("GSE Agent| add agent message to msg channel timeout, meta[%s], len[%d]", msg.Meta, len(msg.Data))
	}
}

// keep recving message.
func (agent *GSEAgent) recvLoop() {
	for i := 0; i < agent.viper.GetInt("server.gseTunnelProcesserNum"); i++ {
		go agent.recv()
	}
}

func (agent *GSEAgent) recv() {
	for {
		select {
		case msg := <-agent.messageChannel:
			agent.recvFunc(msg)

		case <-agent.stopCh:
			logger.Infof("GSE Agent| stop recving message now!")
			return
		}
	}
}

// Stop stops recving message.
func (agent *GSEAgent) Stop() {
	close(agent.stopCh)
}
