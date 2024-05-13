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

package netcheck

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

func TestCheckClusterNet(t *testing.T) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("Hello, world!"),
		},
	}
	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	host := "127.0.0.1"
	start := time.Now()
	conn.SetDeadline(start.Add(5 * time.Second))
	_, err = conn.WriteTo(msgBytes, &net.IPAddr{IP: net.ParseIP(host)})
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	reply := make([]byte, 1500)
	_, _, err = conn.ReadFrom(reply)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	duration := time.Since(start)
	fmt.Printf("Ping successful, time=%v\n", duration)

}
