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

package client

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-services/bcs-health/util"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type ClientInterface interface {
	WatchJobs(slave *util.Slave, jobChan chan *util.Job, errChan chan error) error
	ListJobs(slave *util.Slave) ([]*util.Job, error)
	ReportJobs(job *util.Job) error
}

func NewClient(zkAddr string, t util.TLS) (ClientInterface, error) {
	var tlsCfg *tls.Config
	var err error
	c := &Client{cli: &http.Client{}}
	if len(t.CaFile) != 0 && len(t.CertFile) != 0 && len(t.KeyFile) != 0 {
		tlsCfg, err = ssl.ClientTslConfVerity(t.CaFile, t.CertFile, t.KeyFile, t.PassWord)
		if err != nil {
			return nil, err
		}
	}

	c.cli.Transport = &http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSClientConfig: tlsCfg,
	}

	master, err := NewMasterTracker(zkAddr)
	if err != nil {
		return nil, fmt.Errorf("new master tracker failed. err: %v", err)
	}
	c.master = master

	return c, nil
}

type Client struct {
	master *MasterTracker
	cli    *http.Client
	//trans  *http.Transport
}

const (
	watch_job_url  string = "/bcshealth/v1/watchjobs"
	list_job_url   string = "/bcshealth/v1/listjobs"
	report_job_url string = "/bcshealth/v1/reportjobs"
	// healthz_url    string = "/bcshealth/v1/healthz"
)

func (c *Client) WatchJobs(slave *util.Slave, jobChan chan *util.Job, errChan chan error) error {
	js, err := json.Marshal(slave)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s%s", c.master.GetServers(), watch_job_url)
	req, err := http.NewRequest("GET", url, bytes.NewReader(js))
	if err != nil {
		return err
	}
	blog.Info("--->>> initialize watching, url: %s", url)
	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	blog.Info("--->>> initialized watch.")

	go func() {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		for {
			r := new(util.SvrResponse)
			if err := decoder.Decode(r); err == io.EOF || err != nil {
				blog.Errorf("watch jobs, decode server response failed. err: %v", err)
				errChan <- err
				return
			}

			js, _ := json.Marshal(r)
			blog.V(5).Infof("watch decode a new response: %s", string(js))

			if r.Error != nil {
				blog.Errorf("watch jobs, got response, but is a err: %v", r.Error)
				continue
			}

			for _, j := range r.Jobs {
				if len(j.Zone) == 0 || len(j.Url) == 0 || len(j.Protocol) == 0 {
					blog.Errorf("watch jobs, but got a invalid job: [%s]", j.Name())
					continue
				}
				jobChan <- j
			}
		}
	}()
	return nil
}

func (c *Client) ListJobs(slave *util.Slave) ([]*util.Job, error) {
	js, err := json.Marshal(slave)
	if err != nil {
		return []*util.Job{}, err
	}

	url := fmt.Sprintf("%s%s", c.master.GetServers(), list_job_url)
	req, err := http.NewRequest("GET", url, bytes.NewReader(js))
	if err != nil {
		return []*util.Job{}, fmt.Errorf("new request failed, err: %v", err)
	}
	resp, err := c.cli.Do(req)
	if err != nil {
		return []*util.Job{}, fmt.Errorf("do request failed, err: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []*util.Job{}, fmt.Errorf("read body failed. err: %v", err)
	}

	r := new(util.SvrResponse)
	if err = json.Unmarshal(body, r); err != nil {
		return []*util.Job{}, fmt.Errorf("unmarshal body failed. err: %v", err)
	}

	if r.Error != nil {
		return []*util.Job{}, fmt.Errorf("request response with err: %v", err)
	}
	return r.Jobs, nil
}

func (c *Client) ReportJobs(job *util.Job) error {
	js, err := json.Marshal(job)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s%s", c.master.GetServers(), report_job_url)
	req, err := http.NewRequest("POST", url, bytes.NewReader(js))
	if err != nil {
		return err
	}
	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
