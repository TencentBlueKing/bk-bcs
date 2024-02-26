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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/kv"
	pbrkv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/released-kv"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// Exporter The Exporter interface defines methods for exporting files.
type Exporter interface {
	Export() ([]byte, error)
}

// YAMLExporter implements the Exporter interface for exporting YAML files.
type YAMLExporter struct {
	OutData map[string]interface{}
}

// JSONExporter implements the Exporter interface for exporting JSON files.
type JSONExporter struct {
	OutData map[string]interface{}
}

// XLSXExporter implements the Exporter interface for exporting XLSX files.
type XLSXExporter struct {
	OutData []RkvOutData
}

// XMLExporter implements the Exporter interface for exporting XML files.
type XMLExporter struct {
	OutData []RkvOutData
}

// Export method implements the Exporter interface, exporting data as a byte slice in YAML format.
func (ye *YAMLExporter) Export() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buffer)
	// 设置缩进
	encoder.SetIndent(2)
	defer func() {
		_ = encoder.Close()
	}()
	err := encoder.Encode(ye.OutData)
	return buffer.Bytes(), err
}

// Export method implements the Exporter interface, exporting data as a byte slice in JSON format.
func (je *JSONExporter) Export() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(je.OutData)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Export method implements the Exporter interface, exporting data as a byte slice in XLSX format.
func (xe *XLSXExporter) Export() ([]byte, error) {
	f := excelize.NewFile()
	_ = f.SetCellValue("Sheet1", "A1", "key")
	_ = f.SetCellValue("Sheet1", "B1", "kv_type")
	_ = f.SetCellValue("Sheet1", "C1", "value")

	for i, data := range xe.OutData {
		row := i + 2
		_ = f.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), data.Key)
		_ = f.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), data.KvType)
		_ = f.SetCellValue("Sheet1", fmt.Sprintf("C%d", row), data.Value)
	}

	b, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// Export method implements the Exporter interface, exporting data as a byte slice in XML format.
func (xe *XMLExporter) Export() ([]byte, error) {
	return xml.Marshal(xe.OutData)
}

// Export is a method of kvService that handles the export functionality.
func (m *kvService) Export(w http.ResponseWriter, r *http.Request) {

	kt := kit.MustGetKit(r.Context())

	appIdStr := chi.URLParam(r, "app_id")
	appId, _ := strconv.Atoi(appIdStr)
	if appId == 0 {
		_ = render.Render(w, r, rest.BadRequest(errors.New("validation parameter fail")))
		return
	}
	format := r.URL.Query().Get("format")
	releaseIDStr := chi.URLParam(r, "release_id")
	releaseID, _ := strconv.Atoi(releaseIDStr)

	var outData map[string]interface{}
	// 如果releaseID等于0 获取未命名版本
	if releaseID == 0 {
		req := &pbcs.ListKvsReq{
			BizId:      kt.BizID,
			AppId:      uint32(appId),
			All:        true,
			WithStatus: true,
			Status: []string{constant.FileStateAdd, constant.FileStateRevise,
				constant.FileStateUnchange},
		}
		kvs, err := m.cfgClient.ListKvs(kt.RpcCtx(), req)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		outData = kvsToOutData(kvs.Details)
	} else {
		req := &pbcs.ListReleasedKvsReq{
			BizId:     kt.BizID,
			AppId:     uint32(appId),
			ReleaseId: uint32(releaseID),
			All:       true,
		}
		rkvs, err := m.cfgClient.ListReleasedKvs(kt.RpcCtx(), req)
		if err != nil {
			_ = render.Render(w, r, rest.BadRequest(err))
			return
		}
		outData = rkvsToOutData(rkvs.Details)
	}

	var exporter Exporter

	switch format {
	case "yaml":
		exporter = &YAMLExporter{OutData: outData}
		w.Header().Set("Content-Disposition", "attachment; filename=output.yaml")
		w.Header().Set("Content-Type", "application/x-yaml")
	case "json":
		exporter = &JSONExporter{OutData: outData}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=output.json")
	default:
		_ = render.Render(w, r, rest.BadRequest(errors.New("invalid format")))
		return
	}
	content, err := exporter.Export()
	if err != nil {
		logs.Errorf("export kv fail, err: %v", err)
		_ = render.Render(w, r, rest.BadRequest(err))
	}
	_, err = w.Write(content)
	if err != nil {
		logs.Errorf("Error writing response:%s", err)
		_ = render.Render(w, r, rest.BadRequest(err))
	}

}

// RkvOutData struct defines the format of exported data
type RkvOutData struct {
	Key    string `json:"key" yaml:"key" xml:"key"`
	KvType string `json:"kv_type" yaml:"kv_type" xml:"kv_type"`
	Value  string `json:"value" yaml:"value" xml:"value"`
}

func rkvsToOutData(details []*pbrkv.ReleasedKv) map[string]interface{} {
	d := map[string]interface{}{}
	for _, rkv := range details {
		var value interface{}
		value = rkv.Spec.Value
		switch rkv.Spec.KvType {
		case string(table.KvNumber):
			i, _ := strconv.Atoi(rkv.Spec.Value)
			value = i
		case string(table.KvJson):
			_ = json.Unmarshal([]byte(rkv.Spec.Value), &value)
			if reflect.TypeOf(value).Kind() != reflect.String {
				jm, _ := json.Marshal(value)
				value = string(jm)
			}
		}
		d[rkv.Spec.Key] = map[string]interface{}{
			"kv_type": rkv.Spec.KvType,
			"value":   value,
		}
	}

	return d
}

func kvsToOutData(details []*pbkv.Kv) map[string]interface{} {
	d := map[string]interface{}{}
	for _, rkv := range details {
		var value interface{}
		value = rkv.Spec.Value
		switch rkv.Spec.KvType {
		case string(table.KvNumber):
			i, _ := strconv.Atoi(rkv.Spec.Value)
			value = i
		case string(table.KvJson):
			_ = json.Unmarshal([]byte(rkv.Spec.Value), &value)
			if reflect.TypeOf(value).Kind() != reflect.String {
				jm, _ := json.Marshal(value)
				value = string(jm)
			}
		}
		d[rkv.Spec.Key] = map[string]interface{}{
			"kv_type": rkv.Spec.KvType,
			"value":   value,
		}
	}

	return d
}
