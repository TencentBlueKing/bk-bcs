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

package main

import (
	"fmt"
	"time"

	redis "gopkg.in/redis.v5"

	"bk-bscp/pkg/redisc"
)

var (
	addrs = []string{"127.0.0.1:6379", "127.0.0.1:6380"}

	pwd = ""
)

func main() {
	cli, err := redisc.NewRedisCli(addrs, time.Second, redis.Options{
		Network:      "tcp",
		Password:     "",
		DB:           0,
		MaxRetries:   3,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  10 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	for i := 0; i < 100000; i++ {
		c := cli.Cli()
		if c == nil {
			panic("can't get client.")
		}

		if err := c.Set(fmt.Sprintf("test-%d", i), i, 60*time.Second).Err(); err != nil {
			panic(err)
		}

		time.Sleep(time.Second)
	}
}
