/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package v0v0v0v202011201517

import (
	"context"

	"github.com/spf13/viper"

	"bk-bscp/cmd/middle-services/bscp-patcher/modules/hpm"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	"bk-bscp/pkg/logger"
)

// Patch is v0.0.0 upgrade patch sign at data 202011201517.
type Patch struct {
	// Name patch version name.
	Name string
}

// GetName returns patch version name.
func (p *Patch) GetName() string {
	return p.Name
}

// NeedToSkip is the func to decide if the patch should be skipped.
func (p *Patch) NeedToSkip(kind string) bool {
	return kind != hpm.ProductKindOA
}

// PatchFunc is the func which would puts target patch drived by hpm.
func (p *Patch) PatchFunc(ctx context.Context, viper *viper.Viper, smgr *dbsharding.ShardingManager) error {
	sd, err := smgr.ShardingDB(dbsharding.BSCPDBKEY)
	if err != nil {
		return err
	}

	st := &database.System{}
	err = sd.DB().Last(st).Error
	if err != nil {
		return err
	}
	logger.Infof("execute patch %s success, %+v", p.Name, st)

	return nil
}
