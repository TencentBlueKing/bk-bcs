/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/plugins/repo-sidecar/client/pack"
)

var (
	rx = regexp.MustCompile(`replicas: (?P<replicasNum>\d)`)
)

func main() {
	// 参数指定HTTP服务监听的地址端口
	address := flag.String("address", "0.0.0.0", "bind address")
	port := flag.Uint("port", 8080, "listen port")
	flag.Parse()

	// 定义render处理函数
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {

		var code int32
		var message, data string

		// 处理返回的内容
		defer func() {
			var responseData []byte
			_ = codec.EncJson(&proto.PluginRenderResp{
				Code:    &code,
				Message: &message,
				Data:    &data,
			}, &responseData)
			_, _ = w.Write(responseData)
		}()

		// 从请求BODY中获取标准协议内容 proto.PluginRenderParam
		var param proto.PluginRenderParam
		if err := codec.DecJsonReader(req.Body, &param); err != nil {

			// 如果解析json出现错误, 则返回code=1
			code = 1
			message = fmt.Sprintf("decode body data failed, %v", err)
			return
		}

		// 可以获取到render环境的环境变量env
		_ = param.GetEnv()
		// 可以获取到需要render的整个目录的tgz格式数据
		_ = param.GetData()

		// 文件data是用base64编码的
		tgzData, err := base64.StdEncoding.DecodeString(param.GetData())
		if err != nil {

			// 如果base64解码出现错误, 则返回code=1
			code = 1
			message = fmt.Sprintf("decode tgz data from base64 failed, %v", err)
			return
		}

		// 解压获得需要render的所有文件
		files, err := pack.UnpackFromTgz(tgzData)
		if err != nil {

			// 如果解压出现错误, 则返回code=1
			code = 1
			message = fmt.Sprintf("unpack from tgz data failed, %v", err)
			return
		}

		// 处理所有文件的render
		// 在这里的示例中, render把所有workload的replicas+1
		for _, f := range files {
			// 跳过了不是yaml的文件
			if !strings.HasSuffix(f.Name, ".yaml") {
				continue
			}

			// f.Content 是当前文件的真实内容
			// 找出 replicas: $NUM 的内容, 并把$NUM+1
			rs := rx.FindSubmatch(f.Content)
			if len(rs) >= 2 {
				replicaNum, _ := strconv.Atoi(string(rs[1]))
				f.Content = rx.ReplaceAll(f.Content, []byte(fmt.Sprintf("replicas: %d", replicaNum+1)))
			}

			data += "\n---\n" + string(f.Content)
		}
	})

	// 拉起HTTP服务
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "brings server up failed, %v", err)
	}
}
