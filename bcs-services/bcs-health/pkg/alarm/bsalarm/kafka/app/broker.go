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

package app

//
// func NewBroker(sock string) (*Broker, error) {
// 	listen, err := net.Listen("unix", sock)
// 	if err != nil {
// 		return nil, fmt.Errorf("listen socket failed, err: %v", err)
// 	}
//
// 	return &Broker{
// 		socket: listen,
// 	}, nil
// }
//
// type Broker struct {
// 	socket net.Listener
// }
//
// func (b Broker) GetData() <-chan io.ReadWriteCloser {
// 	ch := make(chan io.ReadWriteCloser, 20)
// 	go func() {
// 		for {
// 			fd, err := b.socket.Accept()
// 			if err != nil {
// 				blog.Errorf("accept unix socket request failed, err: %v", err)
// 				continue
// 			}
// 			ch <- fd
// 		}
// 	}()
// 	return ch
// }
//
// func (b Broker) Log(log string) {
//
// }
