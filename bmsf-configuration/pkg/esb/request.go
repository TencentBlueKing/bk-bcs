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

package esb

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Request is request struct to esb.
type Request struct {
	// Ctx is esb context.
	Ctx *Context

	// Content request content.
	Content interface{}

	// Extra request extra content.
	Extra map[string]interface{}
}

// Marshal marshals esb request data to json including content and extra part.
func (r *Request) Marshal() ([]byte, error) {
	data := bytes.Buffer{}

	// build request data.
	data.WriteRune('{')
	data.WriteString(fmt.Sprintf("\"bk_app_code\":\"%s\"", r.Ctx.AppCode))

	data.WriteRune(',')
	data.WriteString(fmt.Sprintf("\"bk_app_secret\":\"%s\"", r.Ctx.AppSecret))

	data.WriteRune(',')
	data.WriteString(fmt.Sprintf("\"bk_username\":\"%s\"", r.Ctx.User))

	// add content.
	if r.Content != nil {
		content, err := json.Marshal(r.Content)
		if err != nil {
			return nil, fmt.Errorf("marshal content failed, %+v", err)
		}
		content = bytes.TrimPrefix(content, []byte{'{'})
		content = bytes.TrimSuffix(content, []byte{'}'})

		data.WriteRune(',')
		data.Write(content)
	}

	// add extra.
	if r.Extra != nil {
		extra, err := json.Marshal(r.Extra)
		if err != nil {
			return nil, fmt.Errorf("marshal extra failed, %+v", err)
		}
		extra = bytes.TrimPrefix(extra, []byte{'{'})
		extra = bytes.TrimSuffix(extra, []byte{'}'})

		data.WriteRune(',')
		data.Write(extra)
	}
	data.WriteRune('}')

	return data.Bytes(), nil
}
