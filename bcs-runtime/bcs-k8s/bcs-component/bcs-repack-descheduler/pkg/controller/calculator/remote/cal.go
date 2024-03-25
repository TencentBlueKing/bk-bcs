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

// Package remote 定义远程计算的实现
package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
)

// CalculatorRemote defines the remote calculator
type CalculatorRemote struct {
	op *options.DeSchedulerOption
}

// NewCalculatorRemote create the instance of remote calculator
func NewCalculatorRemote(op *options.DeSchedulerOption) calculator.CalculateInterface {
	return &CalculatorRemote{
		op: op,
	}
}

// Calculate will calculate converge response from remote.
func (c *CalculatorRemote) Calculate(ctx context.Context, calculatorRequest *calculator.CalculateConvergeRequest) (
	plan calculator.ResultPlan, err error) {
	bs, err := json.Marshal(calculatorRequest)
	if err != nil {
		return plan, errors.Wrapf(err, "marshal failed")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.op.BKDataUrl, bytes.NewBuffer(bs))
	if err != nil {
		return plan, errors.Wrapf(err, "create request failed")
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	blog.V(4).Infof("Calculator request created and will do request.")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return plan, errors.Wrapf(err, "do request failed")
	}
	blog.V(4).Infof("Calculator do request completed.")
	defer resp.Body.Close()
	respBS, err := io.ReadAll(resp.Body)
	if err != nil {
		return plan, errors.Wrapf(err, "read response body failed")
	}
	if resp.StatusCode != http.StatusOK {
		return plan, errors.Errorf("response status not 200 but %d, respBody: %s", resp.StatusCode, string(respBS))
	}
	blog.V(4).Infof("Calculator read response body success.")

	calculatorResp := new(calculator.CalculateConvergeResponse)
	if err := json.Unmarshal(respBS, calculatorResp); err != nil {
		return plan, errors.Wrapf(err, "unmarshal failed, body: %s", string(respBS))
	}
	if err := c.checkResponse(calculatorResp); err != nil {
		return plan, err
	}
	return calculatorResp.Data.Data.Data[0].Output[0].Plan, nil
}

const (
	defaultInfraPlatformSuccessStatus = "success"
)

func (c *CalculatorRemote) checkResponse(resp *calculator.CalculateConvergeResponse) error {
	blog.V(4).Infof("Calculator response: %s", resp.String())
	if !resp.Result && !resp.Data.Result {
		return errors.Errorf("calculator response result false, resp: %s", resp.String())
	}
	if resp.Data.Data.Status != defaultInfraPlatformSuccessStatus {
		return errors.Errorf("calculator response status not success, resp: %s", resp.String())
	}
	resultData := resp.Data.Data.Data
	if len(resultData) == 0 {
		return errors.Errorf("calculator response data length 0, resp: %s", resp.String())
	}
	if len(resultData[0].Output) == 0 {
		return errors.Errorf("calculator response output length 0, resp: %s", resp.String())
	}
	if len(resultData[0].Output[0].Plan.Plans) == 0 {
		return errors.Errorf("calculator response output plans 0, resp: %s", resp.String())
	}
	return nil
}
