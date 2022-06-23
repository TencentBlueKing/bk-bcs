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

package storegw

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

// PortRange grpc端口范围
type PortRange struct {
	Start int64
	End   int64
}

// NewPortRange 解析portRange
func NewPortRange(matchString string) (*PortRange, error) {
	r := regexp.MustCompile(`^(?P<start>\d+)-(?P<end>\d+)$`)
	submatch := r.FindStringSubmatch(matchString)
	if len(submatch) == 0 {
		return nil, errors.New("port-range not valid")
	}

	start, err := strconv.ParseInt(submatch[1], 10, 64)
	if err != nil {
		return nil, err
	}

	end, err := strconv.ParseInt(submatch[2], 10, 64)
	if err != nil {
		return nil, err
	}

	return &PortRange{Start: start, End: end}, nil
}

// AllocatePort 动态选择一个合适的端口
func (p *PortRange) AllocatePort(idx int64) (int64, error) {
	if idx+p.Start > p.End {
		return 0, errors.New("Port resources have been exhausted")
	}

	// 按顺序选择一个合适的端口
	return idx + p.Start, nil
}
