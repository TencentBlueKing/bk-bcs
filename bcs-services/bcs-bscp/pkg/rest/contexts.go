/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"strconv"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/jsoni"

	"github.com/emicklei/go-restful/v3"
)

// Contexts NOTES
type Contexts struct {
	Kit            *kit.Kit
	Request        *restful.Request
	resp           *restful.Response
	respStatusCode int

	// request meta info
	bizID string
	appID string
}

// DecodeInto decode request body to a struct, if failed, then return the
// response with an error
func (c *Contexts) DecodeInto(to interface{}) error {

	err := jsoni.Decoder(c.Request.Request.Body).Decode(to)
	if err != nil {
		logs.ErrorDepthf(1, "decode request body failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return errf.New(errf.InvalidParameter, err.Error())
	}

	return nil
}

// WithMeta set the request meta which is decoded from the request.
func (c *Contexts) WithMeta(bizID, appID uint32) {
	c.bizID = strconv.FormatUint(uint64(bizID), 10)
	c.appID = strconv.FormatUint(uint64(appID), 10)
}

// WithStatusCode set the response status header code
func (c *Contexts) WithStatusCode(statusCode int) *Contexts {
	c.respStatusCode = statusCode
	return c
}

// respEntity response request with a success response.
func (c *Contexts) respEntity(data interface{}) {
	if c.respStatusCode != 0 {
		c.resp.WriteHeader(c.respStatusCode)
	}

	c.resp.Header().Set(constant.RidKey, c.Kit.Rid)

	resp := &Response{
		Code:    errf.OK,
		Message: "",
		Data:    data,
	}

	if err := jsoni.Encoder(c.resp.ResponseWriter).Encode(resp); err != nil {
		logs.ErrorDepthf(1, "do response failed, err: %s, rid: %s", err.Error(), c.Kit.Rid)
		return
	}

	return
}

// respError response request with error response.
func (c *Contexts) respError(err error) {
	if c.respStatusCode > 0 {
		c.resp.WriteHeader(c.respStatusCode)
	}

	if c.Kit != nil {
		c.resp.Header().Set(constant.RidKey, c.Kit.Rid)
	}

	parsed := errf.Error(err)
	resp := &Response{
		Code:    parsed.Code,
		Message: parsed.Message,
		Data:    nil,
	}

	encodeErr := jsoni.Encoder(c.resp.ResponseWriter).Encode(resp)
	if encodeErr != nil {
		logs.ErrorDepthf(1, "response with error failed, err: %v, rid: %s", encodeErr, c.Kit.Rid)
		return
	}

	return
}
