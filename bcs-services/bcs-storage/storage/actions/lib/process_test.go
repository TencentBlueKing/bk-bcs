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

package lib

// import (
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"testing"

// 	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
// 	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"

// 	"github.com/emicklei/go-restful"
// )

// func TestMarkProcess(t *testing.T) {
// 	bodyStr := "hello world"
// 	req, _ := http.NewRequest("GET", "/", ioutil.NopCloser(strings.NewReader(bodyStr)))
// 	apiserver.GetAPIResource().Conf = &options.StorageOptions{PrintBody: true, QueryMaxNum: 100}

// 	MarkProcess(func(req *restful.Request, resp *restful.Response) {
// 		body, err := ioutil.ReadAll(req.Request.Body)
// 		if err != nil || string(body) != bodyStr {
// 			t.Errorf("MarkProcess() do not pass the correct body!")
// 		}
// 	})(restful.NewRequest(req), &restful.Response{})
// }
