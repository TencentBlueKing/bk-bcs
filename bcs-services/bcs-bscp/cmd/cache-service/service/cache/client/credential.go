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

package client

import (
	"errors"
	"fmt"
	"time"

	prm "github.com/prometheus/client_golang/prometheus"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/keys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func (c *client) GetCredential(kt *kit.Kit, bizID uint32, credential string) (string, error) {
	start := time.Now()
	defer func() {
		c.mc.refreshLagMS.With(prm.Labels{"rsc": credentialRes, "biz": tools.Itoa(bizID)}).Observe(tools.SinceMS(start))
	}()

	cred, hit, err := c.getCredentialFromCache(kt, bizID, credential)
	if err != nil {
		return "", err
	}

	if !hit {
		cred, err = c.refreshCredentialFromCache(kt, bizID, credential)
		if err != nil {
			return "", err
		}
	}

	c.mc.hitCounter.With(prm.Labels{"rsc": credentialRes, "biz": tools.Itoa(bizID)}).Inc()

	return cred, nil
}

func (c *client) getCredentialFromCache(kt *kit.Kit, bizID uint32, credential string) (string, bool, error) {

	val, err := c.bds.Get(kt.Ctx, keys.Key.Credential(bizID, credential))
	if err != nil {
		return "", false, err
	}

	if len(val) == 0 {
		return "", false, nil
	}

	if val == keys.Key.NullValue() {
		return "", false, errf.New(errf.RecordNotFound, fmt.Sprintf("credential : %d-%s not found", bizID, credential))
	}

	return val, true, nil
}

// refreshCredentialFromCache get the credential from db and try to refresh to the cache.
func (c *client) refreshCredentialFromCache(kt *kit.Kit, bizID uint32, credential string) (string, error) {
	cancel := kt.CtxWithTimeoutMS(200)
	defer cancel()

	cred, size, err := c.queryCredentialFromCahce(kt, bizID, credential)
	if err != nil {
		return "", err
	}

	// refresh app credential cache.
	if err := c.bds.Set(kt.Ctx, keys.Key.Credential(bizID, credential),
		cred, keys.Key.CredentialTtlSec(false)); err != nil {
		return "", fmt.Errorf("set biz: %d, credential: %s, cache failed, err: %v", bizID, credential, err)
	}

	c.mc.cacheItemByteSize.With(prm.Labels{"rsc": credentialRes, "biz": tools.Itoa(bizID)}).Observe(float64(size))

	return cred, nil
}

func (c *client) queryCredentialFromCahce(kt *kit.Kit, bizID uint32, credential string) (string, int, error) {
	cred, err := c.op.Credential().GetByCredentialString(kt, bizID, credential)
	if err != nil {
		return "", 0, err
	}
	if errors.Is(err, errf.ErrCredentialInvalid) {
		return "", 0, errf.Newf(errf.InvalidParameter, "invalid credential: %s", credential)
	}
	details, _, err := c.op.CredentialScope().Get(kt, cred.ID, bizID)
	if err != nil {
		return "", 0, err
	}
	scope := make([]string, 0, len(details))
	for _, detail := range details {
		scope = append(scope, string(detail.Spec.CredentialScope))
	}
	credentialCache := &types.CredentialCache{
		Enabled: cred.Spec.Enable,
		Scope:   scope,
	}
	b, err := jsoni.Marshal(credentialCache)
	if err != nil {
		logs.Errorf("marshal credential: %d-%s,failed, err: %v", bizID, credential, err)
		return "", 0, err
	}
	return string(b), len(b), nil
}
