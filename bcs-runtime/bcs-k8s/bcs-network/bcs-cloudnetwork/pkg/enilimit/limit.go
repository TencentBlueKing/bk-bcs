/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package enilimit

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// EnvNameExtraForEniLimitation env name for eni limitation
	EnvNameForExtraEniLimitation = "EXTRA_ENI_LIMITATION"
)

// ExtraLimitation limitation for eni number and ip number for each eni
type ExtraLimitation struct {
	EniNum int `json:"maxEniNum"`
	IPNum  int `json:"maxIPNum"`
}

// Getter eni limitation getter
type Getter struct {
	values map[string]ExtraLimitation
}

// NewGetterFromEnv create limitation getter from environment
func NewGetterFromEnv() (*Getter, error) {
	envValue := os.Getenv(EnvNameForExtraEniLimitation)
	if len(envValue) == 0 {
		return nil, fmt.Errorf("env %s is empty", EnvNameForExtraEniLimitation)
	}
	limitMap := make(map[string]ExtraLimitation)
	if err := json.Unmarshal([]byte(envValue), &limitMap); err != nil {
		return nil, fmt.Errorf("decode value %s from env %s failed", string(envValue), EnvNameForExtraEniLimitation)
	}
	return &Getter{
		values: limitMap,
	}, nil
}

// GetLimit get eni num limit and ip num limit
func (g *Getter) GetLimit(vmType string) (int, int, bool) {
	li, ok := g.values[vmType]
	if !ok {
		return 0, 0, false
	}
	return li.EniNum, li.IPNum, true
}
