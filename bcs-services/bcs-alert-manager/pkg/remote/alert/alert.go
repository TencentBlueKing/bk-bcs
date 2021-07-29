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

package alert

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/metrics"

	"github.com/parnurzeal/gorequest"
)

const (
	DefaultAlarmProjectID = "5805f1b824134fa39318fb0cf59f694b"
)

// AlarmReqData request alertServer body
type AlarmReqData struct {
	StartsTime   time.Time         `json:"startsAt"`
	EndsTime     time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL,omitempty"`
	Annotations  map[string]string `json:"annotations"` // 非判断信息
	Labels       map[string]string `json:"labels"`      // key/value键值对, 通过label可以判断是否是同一告警
}

// AlarmResData resp body
type AlarmResData struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	ErrType string `json:"errorType"`
	Error   string `json:"error"`
}

// BcsAlarmInterface for alarmInterface
type BcsAlarmInterface interface {
	SendAlarmInfoToAlertServer(data []AlarmReqData, timeOut time.Duration) error
}

type alarmServer struct {
	server      string
	appCode     string
	appSecret   string
	testDebug   bool
	serverDebug bool
}

// Option xxx
type Option func(alarm *alarmServer)

// WithTestDebug set alarmServer testDebug
func WithTestDebug(test bool) Option {
	return func(alarm *alarmServer) {
		alarm.testDebug = test
	}
}

// NewAlertServer init alert server object
func NewAlertServer(options *config.AlertServerOptions, opts ...Option) BcsAlarmInterface {
	err := validateServerOptions(options)
	if err != nil {
		glog.Errorf("init alertServer failed: %v", err)
		return nil
	}

	alarmSvr := &alarmServer{
		server:      options.Server,
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		serverDebug: options.ServerDebug,
	}

	for _, opt := range opts {
		opt(alarmSvr)
	}

	return alarmSvr
}

// SendAlarmInfoToAlertServer send alarmInfo to alertServer
func (s *alarmServer) SendAlarmInfoToAlertServer(data []AlarmReqData, timeOut time.Duration) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "SendAlarmInfoToAlertServer"
		path    = "/api/v1/bcs/alerts"
	)

	var (
		url      = s.server + path
		start    = time.Now()
		respData = &AlarmResData{}
	)

	resp, _, errs := gorequest.New().
		Timeout(timeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Param("app_code", s.appCode).
		Param("app_secret", s.appSecret).
		SetDebug(s.serverDebug).
		Send(data).
		EndStruct(respData)
	if len(errs) > 0 {
		glog.Errorf("call api SendAlarmInfoToAlertServer failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportAlertAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("alertServer error: code[%v], status[%v], errType[%s], err[%s]",
			resp.StatusCode, respData.Status, respData.ErrType, respData.Error)
		if !s.testDebug {
			metrics.ReportAlertAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", resp.StatusCode), start)
		}
		return errMsg
	}

	if !s.testDebug {
		metrics.ReportAlertAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", resp.StatusCode), start)
	}

	return nil
}

func validateServerOptions(options *config.AlertServerOptions) error {
	if options == nil {
		return ErrInvalidateOptions
	}

	if !strings.HasPrefix(options.Server, "http") && !strings.HasPrefix(options.Server, "https") {
		return ErrBadServer
	}

	if len(options.AppCode) == 0 || len(options.AppSecret) == 0 {
		return ErrInvalidateAuth
	}

	return nil
}

// PackURLPath will pack the url with a url template and args
func PackURLPath(tpl string, args map[string]string) string {
	if args == nil {
		return tpl
	}
	for k, v := range args {
		tpl = strings.Replace(tpl, "{"+k+"}", url.QueryEscape(v), 1)
	}
	return tpl
}
