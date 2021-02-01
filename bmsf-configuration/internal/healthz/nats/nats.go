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

package natsz

import (
	"crypto/tls"
	"time"

	mq "bk-bscp/pkg/natsmq"
	"bk-bscp/pkg/ssl"
)

// Healthz checks the nats server health state.
func Healthz(host, caFile, certFile, keyFile, passwd string) (bool, error) {
	var err error
	var tlsConf *tls.Config

	var publisher *mq.Publisher
	var subscriber *mq.Subscriber

	if len(caFile) != 0 || len(certFile) != 0 || len(keyFile) != 0 {
		if tlsConf, err = ssl.ClientTLSConfVerify(caFile, certFile, keyFile, passwd); err != nil {
			return false, err
		}
		publisher = mq.NewPublisher([]string{host}, tlsConf)
		subscriber = mq.NewSubscriber([]string{host}, tlsConf)
	} else {
		publisher = mq.NewPublisher([]string{host}, nil)
		subscriber = mq.NewSubscriber([]string{host}, nil)
	}

	if err := publisher.Init(time.Second, time.Second, 1); err != nil {
		return false, err
	}
	if err := subscriber.Init(time.Second, time.Second, 1); err != nil {
		return false, err
	}

	isHealthy := false
	if err := subscriber.Subscribe("healthz", func(bytes []byte) {
		isHealthy = true
	}); err != nil {
		return false, err
	}
	if err := publisher.Publish("healthz", []byte("healthz")); err != nil {
		return false, err
	}

	publisher.Close()
	subscriber.Close()

	return isHealthy, nil
}
