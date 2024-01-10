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

package repository

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tmplprocess"
)

// variableCacheTTLSeconds is ttl for variable cache, 7 day.
const (
	variableCacheTTLSeconds = 3600 * 24 * 7
)

var errContentSizeExceedsLimit = fmt.Errorf("content size exceeds maximum limit %d", constant.MaxRenderBytes)

// VariableCacher is used to set/get template variables with cache
type VariableCacher interface {
	// SetVariables sets template variables into cache, the variables extracted from repository file content
	SetVariables(kt *kit.Kit, sign string, checkSize bool) ([]string, error)
	// GetVariables gets template variables from the cache firstly; if not found, then get from the repository
	GetVariables(kt *kit.Kit, sign string, checkSize bool) ([]string, error)
}

type varCacher struct {
	p        BaseProvider
	bds      bedis.Client
	tmplProc tmplprocess.TmplProcessor
}

// newVariableCacher new a variable cacher
func newVariableCacher(redisConf cc.RedisCluster, p BaseProvider) (VariableCacher, error) {
	// init redis client
	bds, err := bedis.NewRedisCache(redisConf)
	if err != nil {
		return nil, fmt.Errorf("new redis cluster failed, err: %v", err)
	}

	return &varCacher{
		p:        p,
		bds:      bds,
		tmplProc: tmplprocess.NewTmplProcessor(),
	}, nil
}

// SetVariables sets template variables into cache, the variables extracted from repository file content
func (c *varCacher) SetVariables(kt *kit.Kit, sign string, checkSize bool) ([]string, error) {
	if checkSize {
		// check content byte size
		m, err := c.p.Metadata(kt, sign)
		if err != nil {
			return nil, err
		}
		if m.ByteSize > constant.MaxRenderBytes {
			return nil, errContentSizeExceedsLimit
		}
	}

	// get template variables from repository
	vars, err := c.getVarsFromRepo(kt, sign)
	if err != nil {
		return nil, err
	}

	var varsBytes []byte
	varsBytes, err = json.Marshal(vars)
	if err != nil {
		return nil, err
	}

	// set variables into cache
	err = c.bds.Set(kt.Ctx, variableCacheKey(sign), string(varsBytes), variableCacheTTLSeconds)
	if err != nil {
		return nil, err
	}

	return vars, nil
}

// GetVariables gets template variables from the cache firstly; if not found, then get from the repository
func (c *varCacher) GetVariables(kt *kit.Kit, sign string, checkSize bool) ([]string, error) {
	if checkSize {
		// check content byte size
		m, err := c.p.Metadata(kt, sign)
		if err != nil {
			return nil, err
		}
		if m.ByteSize > constant.MaxRenderBytes {
			return nil, errContentSizeExceedsLimit
		}
	}

	// get variables from cache
	val, err := c.bds.Get(kt.Ctx, variableCacheKey(sign))
	if err != nil {
		return nil, err
	}

	var vars []string
	if len(val) > 0 {
		if err = json.Unmarshal([]byte(val), &vars); err != nil {
			return nil, err
		}
		return vars, nil
	}

	// if no variables in cache, get them from the repository and set them into cache
	return c.SetVariables(kt, sign, false)
}

// getVarsFromRepo get template variables from repository
func (c *varCacher) getVarsFromRepo(kt *kit.Kit, sign string) ([]string, error) {
	// download content and extract variables from it
	body, _, err := c.p.Download(kt, sign)
	if err != nil {
		return nil, err
	}

	var content []byte
	content, err = io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	return c.tmplProc.ExtractVariables(content), nil
}

func variableCacheKey(sign string) string {
	return "vars_" + sign
}
