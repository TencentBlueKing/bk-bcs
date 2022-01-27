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

package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// SignatureMethodHMacSha256 method
	SignatureMethodHMacSha256 = "HmacSHA256"
)

// Client tke client
type Client struct {
	*http.Client

	credential CredentialInterface
	opts       Opts
}

// Opts for client config
type Opts struct {
	Method          string
	Region          string
	Host            string
	Path            string
	SignatureMethod string
	Schema          string

	Logger *logrus.Logger
}

// CredentialInterface credential interface
type CredentialInterface interface {
	GetSecretId() (string, error)
	GetSecretKey() (string, error)

	Values() (CredentialValues, error)
}

// CredentialValues xxx
type CredentialValues map[string]string

// Credential xxx
type Credential struct {
	// SecretId xxx
	SecretID  string
	// SecretKey xxx
	SecretKey string
}

// GetSecretId for secretID
func (cred Credential) GetSecretId() (string, error) {
	return cred.SecretID, nil
}

// GetSecretKey for secretKey
func (cred Credential) GetSecretKey() (string, error) {
	return cred.SecretKey, nil
}

// Values for credential values
func (cred Credential) Values() (CredentialValues, error) {
	return CredentialValues{}, nil
}

// NewClient init tke client
func NewClient(credential CredentialInterface, opts Opts) (*Client, error) {
	if opts.Method == "" {
		opts.Method = http.MethodGet
	}
	if opts.SignatureMethod == "" {
		opts.SignatureMethod = SignatureMethodHMacSha256
	}
	if opts.Schema == "" {
		opts.Schema = "https"
	}
	if opts.Logger == nil {
		opts.Logger = logrus.New()
	}
	return &Client{
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: time.Second * 60,
		},
		credential,
		opts,
	}, nil
}

// Invoke call different interface
func (client *Client) Invoke(action string, args interface{}, response interface{}) error {
	switch client.opts.Method {
	case "GET":
		return client.InvokeWithGET(action, args, response)
	default:
		return client.InvokeWithPOST(action, args, response)
	}
}

func (client *Client) initCommonArgs(args *url.Values) {
	args.Set("Region", client.opts.Region)
	args.Set("Timestamp", fmt.Sprint(uint(time.Now().Unix())))
	args.Set("Nonce", fmt.Sprint(uint(rand.Int())))
	args.Set("SignatureMethod", client.opts.SignatureMethod)
}

func (client *Client) signGetRequest(secretId, secretKey string, values *url.Values) string {

	values.Set("SecretId", secretId)

	keys := make([]string, 0, len(*values))
	for k := range *values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	kvs := make([]string, 0, len(keys))
	for _, k := range keys {
		kvs = append(kvs, fmt.Sprintf("%s=%s", k, values.Get(k)))
	}
	queryStr := strings.Join(kvs, "&")
	reqStr := fmt.Sprintf("GET%s%s?%s", strings.Split(client.opts.Host, ":")[0], client.opts.Path, queryStr)

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(reqStr))
	signature := mac.Sum(nil)

	b64Encoded := base64.StdEncoding.EncodeToString(signature)

	return b64Encoded
}

// InvokeWithGET get interface
func (client *Client) InvokeWithGET(action string, args interface{}, response interface{}) error {
	reqValues := url.Values{}

	credValues, err := client.credential.Values()
	if err != nil {
		return makeClientError(err)
	}

	for k, v := range credValues {
		reqValues.Set(k, v)
	}

	err = EncodeStruct(args, &reqValues)
	if err != nil {
		return makeClientError(err)
	}
	reqValues.Set("Action", action)
	client.initCommonArgs(&reqValues)

	secretIdD, err := client.credential.GetSecretId()
	if err != nil {
		return makeClientError(err)
	}
	secretKey, err := client.credential.GetSecretKey()
	if err != nil {
		return makeClientError(err)
	}

	signature := client.signGetRequest(secretIdD, secretKey, &reqValues)
	reqValues.Set("Signature", signature)

	reqQuery := reqValues.Encode()

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s://%s%s?%s", client.opts.Schema, client.opts.Host, client.opts.Path, reqQuery),
		nil,
	)

	if err != nil {
		return makeClientError(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return makeClientError(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return makeClientError(err)
	}

	client.opts.Logger.WithField("Action", action).Infof(
		"%s %s %d %s", "GET", req.URL, resp.StatusCode, body,
	)

	legacyErrorResponse := LegacyAPIError{}

	if err = json.Unmarshal(body, &legacyErrorResponse); err != nil {
		return makeClientError(err)
	}

	versionErrorResponse := VersionAPIError{}

	if err = json.Unmarshal(body, &versionErrorResponse); err != nil {
		return makeClientError(err)
	}

	if legacyErrorResponse.Code != NoErr || (legacyErrorResponse.CodeDesc != "" && legacyErrorResponse.CodeDesc != NoErrCodeDesc) {
		client.opts.Logger.WithField("Action", action).Errorf(
			"%s %s %d %s %v", "GET", req.URL, resp.StatusCode, body, legacyErrorResponse,
		)
		return legacyErrorResponse
	}

	if versionErrorResponse.Response.Error.Code != "" {
		client.opts.Logger.WithField("Action", action).Errorf(
			"%s %s %d %s %v", "GET", req.URL, resp.StatusCode, body, versionErrorResponse,
		)
		return versionErrorResponse
	}

	if err = json.Unmarshal(body, response); err != nil {
		return makeClientError(err)
	}

	return nil
}

// InvokeWithPOST post interface
func (client *Client) InvokeWithPOST(action string, args interface{}, response interface{}) error {
	return nil
}
