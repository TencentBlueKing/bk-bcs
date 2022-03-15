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

package notify

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/lock"
	etcdlock "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/lock/etcd"
	"github.com/robfig/cron"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/esb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

type TokenNotify interface {
	Run()
	Stop()
}

type tokenNotify struct {
	// notify message
	emailTitle   string
	emailContent string
	rtxTitle     string
	rtxContent   string

	// bk app
	appCode       string
	appSecret     string
	apiHost       string
	sendEmailPath string
	sendRtxPath   string
	dryRun        bool

	// lock
	locker  *etcdlock.Client
	started int32

	// cron
	cron           *cron.Cron
	cronExpression string
}

const (
	TokenNotifyLockName = "tokenNotify"
)

func NewTokenNotify(op *options.UserManagerOptions) (TokenNotify, error) {
	var tlsCfg *tls.Config
	if !op.InsecureEtcd {
		var err error
		tlsCfg, err = op.Etcd.GetTLSConfig()
		if err != nil {
			return nil, err
		}
	}

	// init locker
	locker, err := etcdlock.New(
		lock.Endpoints(op.Etcd.Address),
		lock.Prefix("usermanager"),
		lock.TLS(tlsCfg),
	)
	if err != nil {
		blog.Errorf("init locker failed, err %s", err.Error())
		return nil, err
	}

	appSecret, err := encrypt.DesDecryptFromBase([]byte(op.TokenNotify.ESBConfig.AppSecret))
	if err != nil {
		return nil, fmt.Errorf("error decrypting app secret, %s", err.Error())
	}
	return &tokenNotify{
		emailTitle:     op.TokenNotify.EmailTitle,
		emailContent:   op.TokenNotify.EmailContent,
		rtxTitle:       op.TokenNotify.RtxTitle,
		rtxContent:     op.TokenNotify.RtxContent,
		appCode:        op.TokenNotify.ESBConfig.AppCode,
		appSecret:      string(appSecret),
		apiHost:        op.TokenNotify.ESBConfig.APIHost,
		sendEmailPath:  op.TokenNotify.ESBConfig.SendEmailPath,
		sendRtxPath:    op.TokenNotify.ESBConfig.SendRtxPath,
		dryRun:         op.TokenNotify.DryRun,
		locker:         locker,
		cron:           cron.New(),
		cronExpression: op.TokenNotify.NotifyCron,
	}, nil
}

func (t *tokenNotify) Run() {
	// acquire lock (or wait to have it)
	if err := t.locker.Lock(TokenNotifyLockName, lock.LockTTL(5)); err != nil {
		blog.Errorf("acquire lock err:", err.Error())
		return
	}
	atomic.StoreInt32(&t.started, 1)
	blog.Infof("acquire token notify lock success")

	// run job
	t.cron.AddFunc(t.cronExpression, func() { t.do() })
	t.cron.Start()
}

func (t *tokenNotify) Stop() {
	t.cron.Stop()
	if atomic.LoadInt32(&t.started) != 1 {
		// notify job not started, no need to release lock
		return
	}
	err := t.locker.Unlock(TokenNotifyLockName)
	if err != nil {
		blog.Errorf("unlock token notify lock failed, err %s", err.Error())
		return
	}
	blog.Infof("release token notify lock success")
}

func (t *tokenNotify) do() {
	blog.Infof("checking expired token")
	// get all tokens
	tokenNotifyStore := sqlstore.NewTokenNotifyStore(sqlstore.GCoreDB)
	tokens := sqlstore.GetAllTokens()
	// notify
	for _, token := range tokens {
		expiration := token.ExpiresAt.Sub(token.UpdatedAt)
		remain := time.Until(token.ExpiresAt)
		matched, phase := match(remain, expiration)
		if !matched {
			continue
		}
		tn := tokenNotifyStore.GetTokenNotifyByCondition(&models.BcsTokenNotify{Token: token.UserToken, Phase: phase})
		if len(tn) > 0 {
			// already notified
			continue
		}
		emailResp := t.sendEmail(token)
		t.insertRecord(emailResp, token, models.NotifyByEmail, phase)
		rtxResp := t.sendRtx(token)
		t.insertRecord(rtxResp, token, models.NotifyByRtx, phase)
	}
}

func (t *tokenNotify) sendEmail(token models.BcsUser) *APIResponse {
	// get email content
	emailContent, err := generateNotifyContent(t.emailContent, notifyUser{Username: token.Name, ExpiredAt: token.ExpiresAt})
	if err != nil {
		blog.Errorf("generate email content err: %s", err.Error())
	}
	payload := map[string]interface{}{
		"receiver": token.Name,
		"title":    t.emailTitle,
		"content":  emailContent,
	}
	var resp *APIResponse

	// if dry run, just pass and insert record
	if t.dryRun {
		resp = &APIResponse{
			Result:    true,
			Code:      "0",
			Message:   "dry run",
			RequestID: "dry run",
		}
	} else {
		resp, err = t.requestEsb("POST", t.sendEmailPath, payload)
		if err != nil {
			resp = &APIResponse{
				Result:  false,
				Code:    "-1",
				Message: err.Error(),
			}
			blog.Errorf("send email err: %s", err.Error())
		}
	}
	return resp
}

func (t *tokenNotify) sendRtx(token models.BcsUser) *APIResponse {
	// get rtx content
	rtxContent, err := generateNotifyContent(t.emailContent, notifyUser{Username: token.Name, ExpiredAt: token.ExpiresAt})
	if err != nil {
		blog.Errorf("generate rtx content err: %s", err.Error())
	}
	payload := map[string]interface{}{
		"sender":   token.Name,
		"receiver": token.Name,
		"title":    t.rtxTitle,
		"message":  rtxContent,
	}
	var resp *APIResponse
	if t.dryRun {
		resp = &APIResponse{
			Result:    true,
			Code:      "0",
			Message:   "dry run",
			RequestID: "dry run",
		}
	} else {
		resp, err = t.requestEsb("POST", t.sendRtxPath, payload)
		if err != nil {
			resp = &APIResponse{
				Result:  false,
				Code:    "-1",
				Message: err.Error(),
			}
			blog.Errorf("send rtx err: %s", err.Error())
		}
	}
	return resp
}

func (t *tokenNotify) insertRecord(resp *APIResponse, token models.BcsUser, notifyType models.NotifyType,
	phase models.NotifyPhase) {
	tokenNotify := &models.BcsTokenNotify{
		Token:      token.UserToken,
		NotifyType: notifyType,
		Phase:      phase,
		Result:     resp.Result,
		Message:    resp.Message,
		RequestID:  resp.RequestID,
	}
	tokenNotifyStore := sqlstore.NewTokenNotifyStore(sqlstore.GCoreDB)
	err := tokenNotifyStore.CreateTokenNotify(tokenNotify)
	if err != nil {
		blog.Errorf("insert token notify err: %s", err.Error())
	}

	if resp.Result {
		blog.Infof("send %s to [%s] success", notifyType, token.Name)
	} else {
		blog.Errorf("send %s to [%s] failed: %s", notifyType, token.Name, resp.Message)
	}
}

type APIResponse struct {
	Result    bool        `json:"result"`
	Code      string      `json:"code"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
}

func (t *tokenNotify) requestEsb(method, url string, payload map[string]interface{}) (*APIResponse, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload can't be nil")
	}
	//set payload app parameter
	payload[esb.EsbRequestPayloadAppcode] = t.appCode
	payload[esb.EsbRequestPayloadAppsecret] = t.appSecret
	payloadBytes, _ := json.Marshal(payload)
	//new request body
	body := bytes.NewBuffer(payloadBytes)
	//request url
	url = fmt.Sprintf("%s%s", t.apiHost, url)

	//new request object
	req, _ := http.NewRequest(method, url, body)
	//set header application/json
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request esb %s failed: %s", url, err.Error())
	}
	defer resp.Body.Close()

	// Parse body as JSON
	var result APIResponse
	respBody, _ := ioutil.ReadAll(resp.Body)
	blog.V(3).Infof("request esb %s resp body(%s)", url, string(respBody))

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("non-Json body(%s) response: %s", string(respBody), err.Error())
	}

	//http response status code != 200
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("response code %d body %s", resp.StatusCode, respBody)
	}
	return &result, nil
}

type notifyUser struct {
	Username  string
	ExpiredAt time.Time
}

func generateNotifyContent(content string, users notifyUser) (string, error) {
	tmpl, err := template.New("test").Parse(content)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, users)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
