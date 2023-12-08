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

package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"

	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/kit"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/rest"
)

type kvService struct {
	authorizer auth.Authorizer
	cfgClient  pbcs.ConfigClient
}

func newKvService(authorizer auth.Authorizer,
	cfgClient pbcs.ConfigClient) *kvService {
	s := &kvService{
		authorizer: authorizer,
		cfgClient:  cfgClient,
	}
	return s
}

// Import is used to handle file import requests.
func (m *kvService) Import(w http.ResponseWriter, r *http.Request) {

	kt := kit.MustGetKit(r.Context())

	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}

	var kvs []*pbcs.BatchUpsertKvsReq_Kv
	reader := bufio.NewReader(r.Body)

	switch {
	case isJSON(reader):
		b, err := io.ReadAll(reader)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		if err = json.Unmarshal(b, &kvs); err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
	case isXML(reader):
		b, err := io.ReadAll(reader)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		if err := xml.Unmarshal(b, &kvs); err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
	case isYAML(reader):
		b, err := io.ReadAll(reader)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		if e := yaml.Unmarshal(b, &kvs); e != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
	case isXLSX(reader):
		kvsTmp, err := parseExcelFile(reader)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		kvs = kvsTmp

	default:
		_ = render.Render(w, r, rest.BadRequest(errors.New("unsupported file type")))
		return
	}

	req := &pbcs.BatchUpsertKvsReq{
		BizId: kt.BizID,
		AppId: uint32(appId),
		Kvs:   kvs,
	}
	resp, err := m.cfgClient.BatchUpsertKvs(kt.RpcCtx(), req)
	if err != nil {
		_ = render.Render(w, r, rest.BadRequest(err))
		return
	}

	_ = render.Render(w, r, rest.OKRender(resp))
}

func parseExcelFile(file io.Reader) ([]*pbcs.BatchUpsertKvsReq_Kv, error) {

	r := bufio.NewReader(file)
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var kvs []*pbcs.BatchUpsertKvsReq_Kv
	for i := 1; i < len(rows); i++ {
		kv := &pbcs.BatchUpsertKvsReq_Kv{
			Key:    rows[i][0],
			Value:  rows[i][1],
			KvType: rows[i][2],
		}
		kvs = append(kvs, kv)
	}

	return kvs, nil
}

func peek(r *bufio.Reader, size int) ([]byte, error) {

	peekedData, err := r.Peek(size)
	if err != nil {
		return nil, err
	}
	return peekedData, nil
}

func startsWith(r *bufio.Reader, prefix string) bool {
	peekedData, err := peek(r, len(prefix))
	if err != nil {
		return false
	}
	return bytes.Equal(peekedData, []byte(prefix))
}

func isJSON(r *bufio.Reader) bool {
	return startsWith(r, "[")
}

func isYAML(r *bufio.Reader) bool {
	return startsWith(r, "kvs :")
}

func isXML(r *bufio.Reader) bool {
	return startsWith(r, "<kvs>")
}

var (
	magicXlsx = []byte{0x50, 0x4B, 0x03, 0x04} // Xlsx文件的魔法数字
)

func isXLSX(r *bufio.Reader) bool {
	headerBytes, err := r.Peek(4)
	if err != nil {
		return false
	}
	magic := headerBytes[0:4]
	return bytes.Equal(magicXlsx, magic[0:4])
}
