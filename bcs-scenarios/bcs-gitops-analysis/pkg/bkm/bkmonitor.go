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

// Package bkm xx
package bkm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-analysis/options"
)

// BKMonitorMessage defines the struct send to bkm
type BKMonitorMessage struct {
	DataID      int64                   `json:"data_id"`
	AccessToken string                  `json:"access_token"`
	Data        []*BKMonitorMessageData `json:"data"`
}

// BKMonitorMessageData defines the data that send to bkm
type BKMonitorMessageData struct {
	Metrics   map[string]interface{} `json:"metrics"`
	Target    string                 `json:"target"`
	Dimension map[string]string      `json:"dimension"`
	Timestamp int64                  `json:"timestamp"`
}

// BKMonitorClient defies the handler to bkmonitor
type BKMonitorClient struct {
	op *options.AnalysisOptions
}

// NewBKMonitorClient create the bkmonitor client
func NewBKMonitorClient() *BKMonitorClient {
	return &BKMonitorClient{
		op: options.GlobalOptions(),
	}
}

// IsPushTurnOn check push to bkm is turn-on
func (b *BKMonitorClient) IsPushTurnOn() bool {
	return b.op.BKMonitorPushUrl != ""
}

// Push the message to bkmonitor
func (b *BKMonitorClient) Push(message *BKMonitorMessage) {
	if !b.IsPushTurnOn() {
		return
	}
	bs, _ := json.Marshal(message)
	httpClient := http.DefaultClient
	httpClient.Timeout = 30 * time.Second
	// blog.Infof("push to bkmonitor: %s", string(bs))
	resp, err := httpClient.Post(b.op.BKMonitorPushUrl, "application/json", bytes.NewBuffer(bs))
	if err != nil {
		blog.Errorf("analysis push to bkmonitor failed: %s", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		blog.Infof("push to bkmonitor success")
		return
	}
	if bs, err = io.ReadAll(resp.Body); err != nil {
		blog.Errorf("push to bkmonitor failed: read resp body failed(%d): %s",
			resp.StatusCode, err.Error())
		return
	}
	blog.Errorf("push to bkmonitor failed, resp not 200 but %d: %s", resp.StatusCode, string(bs))
}

// BKMonitorGetMessage defines the message of query bkmonitor
type BKMonitorGetMessage struct {
	PromQL         string            `json:"promql"`
	PromQLs        map[string]string `json:"promqls"`
	Start          string            `json:"start"`
	End            string            `json:"end"`
	RecentDuration int64             `json:"recentDuration,omitempty"` // unit:second
	Step           string            `json:"step"`

	BKBizType string `json:"bkBizType"`
}

// Complete the start/end/step with default
func (m *BKMonitorGetMessage) Complete() {
	timeNow := time.Now()
	if m.RecentDuration != 0 {
		m.Start = fmt.Sprintf("%d", timeNow.Add(-time.Duration(m.RecentDuration)*time.Second).Unix())
		m.RecentDuration = 0
	} else if m.Start == "" {
		m.Start = fmt.Sprintf("%d", timeNow.Add(-24*time.Hour).Unix())
	}
	if m.End == "" {
		m.End = fmt.Sprintf("%d", timeNow.Unix())
	}
	if m.Step == "" {
		m.Step = "3600s"
	}
}

// BKMonitorGetResponse defines the message that bkmonitor response
type BKMonitorGetResponse struct {
	Series []BKMonitorSeries `json:"series"`
}

// BKMonitorSeries defines the bkmonitor response series
type BKMonitorSeries struct {
	GroupKeys   []string    `json:"group_keys"`
	GroupValues []string    `json:"group_values"`
	Values      [][]float64 `json:"values"`
}

// BKMonitorSeriesItem defines the series of bkmonitor search
type BKMonitorSeriesItem struct {
	Name      string  `json:"name,omitempty"`
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// Get the bkmonitor data with queryh
func (b *BKMonitorClient) Get(message *BKMonitorGetMessage) (*BKMonitorGetResponse, error) {
	message.Complete()
	paseJSON, _ := json.Marshal(message) // nolint
	blog.Infof("query bkmonitor: %s", string(paseJSON))
	req, err := http.NewRequest(http.MethodPost, b.op.BKMonitorGetUrl, bytes.NewBuffer(paseJSON))
	if err != nil {
		return nil, errors.Wrapf(err, "create http request failed")
	}
	req.Header.Set("X-Bkapi-Authorization",
		fmt.Sprintf(`{"bk_app_code":"%s","bk_app_secret":"%s", "bk_username": "%s"}`,
			b.op.Auth.AppCode, b.op.Auth.AppSecret, b.op.BKMonitorGetUser))
	if message.BKBizType != "" {
		req.Header.Set("X-Bk-Scope-Space-Uid", message.BKBizType)
	} else {
		req.Header.Set("X-Bk-Scope-Space-Uid", fmt.Sprintf("bkcc__%d", b.op.BKMonitorGetBizID))
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "do http request failed")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "read http response failed")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("bkmonitor get resp code not 200 but %d: %s", resp.StatusCode, string(body))
	}
	bkmResp := new(BKMonitorGetResponse)
	if err = json.Unmarshal(body, bkmResp); err != nil {
		return nil, errors.Wrapf(err, "bkmonitor get resp unmarshal '%s' failed", string(body))
	}
	return bkmResp, nil
}
