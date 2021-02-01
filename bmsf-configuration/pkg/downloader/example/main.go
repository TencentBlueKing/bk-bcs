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

	dl "bk-bscp/pkg/downloader"
)

var (
	// NOTE: add your test source file url here, build your own request header
	// and setup the download rate limitations.

	target             = "http://source.com/api/generic/my-project/my-repo/path/source.tgz"
	newFile            = "./local-source.tgz"
	headers            = map[string]string{}
	timeout            = 30 * time.Second
	concurrent         = 5
	limitBytesInSecond = int64(1024 * 1024 * 10) // 10MB
)

func main() {
	// create a downloader instance.
	downloader := dl.NewDownloader(target, concurrent, headers, newFile)

	// setup downloader rate limitations.
	downloader.SetRateLimiterOption(dl.NewSimpleRateLimiter(limitBytesInSecond))

	// download now.
	timenow := time.Now()
	if err := downloader.Download(timeout); err != nil {
		// clean temp file.
		downloader.Clean()
		panic(err)
	}
	fmt.Println("download success, cost: ", time.Since(timenow))
}
