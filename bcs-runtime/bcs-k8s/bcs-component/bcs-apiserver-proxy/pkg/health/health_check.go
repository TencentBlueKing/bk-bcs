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

// Package health xxx
package health

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

var (
	// ErrSchemeInValid invalid scheme
	ErrSchemeInValid = errors.New("scheme invalid, should be http/https")
	// ErrHealthConfigNotInited show HealthConfig not inited
	ErrHealthConfigNotInited = errors.New("healthConfig not inited")
)

// HealthCheck is interface for check addr:port health
// nolint
type HealthCheck interface {
	IsHTTPAPIHealth(addr string, port uint32) bool
}

const (
	schemeHTTPS = "https"
	schemeHTTP  = "http"
)

func validateScheme(scheme string) error {
	if scheme != schemeHTTPS && scheme != schemeHTTP {
		return ErrSchemeInValid
	}

	return nil
}

// NewHealthConfig init HealthConfig
func NewHealthConfig(scheme string, path string) (HealthCheck, error) {
	err := validateScheme(scheme)
	if err != nil {
		return nil, err
	}

	return &HealthConfig{
		Shem: scheme,
		Path: path,
	}, nil
}

// HealthConfig conf immutable schem/path
// nolint
type HealthConfig struct {
	Shem string
	Path string
}

// IsHTTPAPIHealth for check schem://addr:port/Path health
func (hc *HealthConfig) IsHTTPAPIHealth(addr string, port uint32) bool {
	if hc == nil {
		blog.Errorf("IsHTTPAPIHealth empty: %v", ErrHealthConfigNotInited)
		return false
	}

	url := fmt.Sprintf("%s://%s:%d%s", hc.Shem, addr, port, hc.Path)
	client := &http.Client{
		Transport: &http.Transport{
			// nolint
			TLSClientConfig: &tls.Config{
				// NOCC:gas/tls(设计如此)
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		blog.Errorf("http health check failed, %v", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		blog.Errorf("IsHTTPAPIHealth[%s] error: %v", url, err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // nolint
		return false
	}

	return true
}
