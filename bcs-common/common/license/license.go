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

package license

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
)

const (
	uri           = "/certificate"
	reqTimeLayout = "2006-01-02 15:04:05"
	reqGapSecond  = 10
)

type status struct {
	// If the check pass or not
	OK bool

	// Success message
	Message string

	// Status error
	Err error

	// Should program quit or not
	Quit bool
}

type config struct {
	c    *tls.Config
	cert []byte
}

type reqData struct {
	Cert     string `json:"certificate"`
	Platform string `json:"platform"`
	Time     string `json:"requesttime"`
}

type respData struct {
	Status         bool   `json:"status"`
	Result         int    `json:"result"`
	Message        string `json:"message"`
	MessageCn      string `json:"message_cn"`
	Time           string `json:"time"`
	ValidStartTime string `json:"validstarttime"`
	ValidEndTime   string `json:"validendtime"`
}

func loadLicense(lsConfig conf.LicenseServerConfig) (*config, error) {
	c, err := ssl.ClientTslConfVerity(lsConfig.LSCAFile, lsConfig.LSClientCertFile, lsConfig.LSClientKeyFile, static.LicenseServerClientCertPwd)
	if err != nil {
		return nil, err
	}

	cert, err := ioutil.ReadFile(lsConfig.LSClientCertFile)
	if err != nil {
		return nil, err
	}

	return &config{c, cert}, nil
}

func checkLicense(lsConfig conf.LicenseServerConfig) (s chan *status) {
	s = make(chan *status)

	var quit bool
	switch edition := version.GetEdition(); edition {
	case version.InnerEdition:
		go func() { s <- &status{OK: true, Quit: false, Err: nil} }()
		return
	case version.CommunicationEdition, version.EnterpriseEdition:
		quit = false
	}

	config, err := loadLicense(lsConfig)
	if err != nil {
		go func() { s <- &status{OK: false, Quit: true, Err: err} }()
		return
	}

	c := httpclient.NewHttpClient()
	c.SetTlsVerityConfig(config.c)

	data := &reqData{
		Cert:     string(config.cert),
		Platform: version.PlatformName,
	}
	head := http.Header{}
	head.Set("Content-Type", "application/json")

	target := url.URL{Scheme: "https", Host: lsConfig.LSAddress, Path: uri}

	go func() {
		for ; ; time.Sleep(reqGapSecond * time.Second) {
			data.Time = time.Now().Format(reqTimeLayout)
			var d []byte
			codec.EncJson(data, &d)
			resp, err := c.Post(target.String(), head, d)
			if err != nil {
				s <- &status{OK: false, Quit: quit, Err: err}
				continue
			}

			var r *respData
			if err = codec.DecJson(resp.Reply, &r); err != nil {
				s <- &status{OK: false, Quit: quit, Err: err}
				continue
			}

			if !r.Status {
				s <- &status{OK: false, Quit: quit, Err: errors.New(r.Message)}
				continue
			}

			s <- &status{OK: true, Quit: false, Err: nil, Message: fmt.Sprintf("Valid Time: %s - %s", r.ValidStartTime, r.ValidEndTime)}
			return
		}
	}()
	return
}

func CheckLicense(lsConfig conf.LicenseServerConfig) {
	c := checkLicense(lsConfig)
	for {
		status := <-c
		if status.OK {
			blog.Infof("License-Check PASS. %s", status.Message)
			return
		}

		blog.Errorf("License-Check FAILED. %v", status.Err)
		if status.Quit {
			os.Exit(1)
		}
	}
}
