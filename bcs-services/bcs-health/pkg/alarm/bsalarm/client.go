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

package bsalarm

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/statistic"
)

func NewKafkaClient(dataid, path, confFile, logdir string) (*KafkaClient, error) {
	if len(path) == 0 || len(confFile) == 0 {
		return nil, errors.New("invalid configuration with kafka bin path or conf file")
	}

	cmd := exec.Command(path, "-c", confFile, "-path.logs", logdir)

	cli := &KafkaClient{
		dataid: dataid,
		cmd:    cmd,
		stop:   make(chan struct{}),
	}

	if err := cli.cmd.Start(); err != nil {
		return nil, fmt.Errorf("start kafka command failed, err: %v", err)
	}

	kafkaid := "kafka_id"
	statistic.Reset(kafkaid)

	go func() {
		// TODO: re-launch the kafka client again if wait failed.
		if err := cli.cmd.Wait(); err != nil {
			statistic.Set(kafkaid, errors.New("kafka client is exit, can not send alarm event now"))
			close(cli.stop)
			blog.Errorf("kafka client exit with err: %v", err)
			fmt.Fprintf(os.Stderr, "kafka client exit with err: %v", err)
			os.Exit(1)
			return
		}
		blog.Errorf("kafka client exit.")
	}()

	sockPath, err := filepath.Abs(DefaultSock)
	if err != nil {
		return nil, fmt.Errorf("abs sock path failed, err: %v", err)
	}

	cli.client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
	}

	return cli, nil
}

type KafkaClient struct {
	dataid string
	client *http.Client
	cmd    *exec.Cmd
	stderr *bufio.Reader
	stop   chan struct{}
}

func (k *KafkaClient) SendEvent(e *AlarmEvent) error {
	stdin := StdInput{
		DataID: k.dataid,
		Data:   *e,
		UUID:   e.Event.UUID,
	}

	js, err := json.Marshal(stdin)
	if err != nil {
		return fmt.Errorf("marshal alarm event failed, uuid:%s, err: %v", stdin.UUID, err)
	}

	blog.Infof("try to send blue shield event: %s", string(js))

	req, err := http.NewRequest("POST", "http://unix/v1/create", strings.NewReader(string(js)))
	if err != nil {
		return fmt.Errorf("new request failed, err: %v", err)
	}

	resp, err := k.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request failed, err: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		blog.Errorf("read response body failed, uuid: %s, err: %v", e.Event.UUID, err)
		return err
	}

	r := new(StdOutput)
	if err := json.Unmarshal(body, r); err != nil {
		blog.Errorf("unmarshal response body failed, uuid: %s, body: %s, err: %v", e.Event.UUID, string(body), err)
		return err
	}

	if r.Success {
		return nil
	}

	return fmt.Errorf("send event err: %s", r.Message)
}
