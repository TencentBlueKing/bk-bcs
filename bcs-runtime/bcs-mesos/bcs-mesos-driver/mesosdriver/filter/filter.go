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

package filter

import (
	bcshttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/emicklei/go-restful"
	"net/http"
)

// NewFilter create general filter
func NewFilter() *GeneralFilter {

	return &GeneralFilter{
		filterFunctions: []RequestFilterFunction{},
	}
}

// GeneralFilter filter slice wrapper
type GeneralFilter struct {
	filterFunctions []RequestFilterFunction
}

// RequestFilterFunction filter function definition
type RequestFilterFunction interface {
	//Execute check http request
	Execute(req *restful.Request) (int, error)
}

// Filter slice implementation
func (gf *GeneralFilter) Filter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	for _, filterFunction := range gf.filterFunctions {
		errCode, err := filterFunction.Execute(req)
		if err != nil {
			resp.WriteHeaderAndEntity(http.StatusBadRequest, bcshttp.APIRespone{
				Result:  false,
				Code:    errCode,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}
	}
	chain.ProcessFilter(req, resp)
}

// AppendFilter add filter
func (gf *GeneralFilter) AppendFilter(filterFunc RequestFilterFunction) {
	gf.filterFunctions = append(gf.filterFunctions, filterFunc)
}
