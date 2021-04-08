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

package local

import (
	"fmt"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/spf13/viper"

	"bk-bscp/internal/casbin/gorm-adapter"
	"bk-bscp/internal/database"
)

const (
	// defaultLoadPolicyInterval is default load policy file interval.
	defaultLoadPolicyInterval = 3 * time.Second
)

// Controller is local file auth controller.
type Controller struct {
	viper *viper.Viper

	// core auth controller base on casbin.
	enforcer *casbin.SyncedEnforcer
}

// NewController creates a new Controller instance.
func NewController(viper *viper.Viper) (*Controller, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
		viper.GetString("database.user"),
		viper.GetString("database.passwd"),
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		database.BSCPDB,
		viper.GetDuration("database.connTimeout"),
		viper.GetDuration("database.readTimeout"),
		viper.GetDuration("database.writeTimeout"),
		database.BSCPCHARSET,
	)

	// gorm database adapter.
	localAuth := database.LocalAuth{}
	adapter, err := gormadapter.NewAdapter("mysql", dsn, database.BSCPDB, localAuth.TableName())
	if err != nil {
		return nil, err
	}

	// NOTE: RBAC x ABAC x ACL multi auth mode.
	// Could add user or resource into 'g',
	// g "alice, admin", alice is admin now.
	// g "book-1, book-group", book-1 resource added to group now.
	// And could change model to support rbac_with_resource_roles(资源角色)、rbac_with_domains(域租户) and so on.
	rbac := model.NewModel()
	rbac.AddDef("r", "r", "sub, obj, act")
	rbac.AddDef("p", "p", "sub, obj, act")
	rbac.AddDef("g", "g", "_, _")
	rbac.AddDef("e", "e", "some(where (p.eft == allow))")
	rbac.AddDef("m", "m", "g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act")

	enforcer, err := casbin.NewSyncedEnforcer(rbac, adapter)
	if err != nil {
		return nil, err
	}
	enforcer.EnableAutoSave(true)
	enforcer.StartAutoLoadPolicy(defaultLoadPolicyInterval)

	controller := &Controller{viper: viper, enforcer: enforcer}

	return controller, nil
}

// Authorize authorizes base on the model/policy in local mode, decides whether a "subject" can access a "object"
// with the operation "action", input parameters are usually: (sub, obj, act).
// example:  "alice" "data1" "read"
func (c *Controller) Authorize(rvals ...interface{}) (bool, error) {
	return c.enforcer.Enforce(rvals...)
}

// AddPolicy adds an authorization rule to the current policy.
// If the rule already exists, the function returns false and the rule will not be added.
// Otherwise the function returns true by adding the new rule.
func (c *Controller) AddPolicy(rvals ...interface{}) (bool, error) {
	return c.enforcer.AddPolicy(rvals...)
}

// RemovePolicy removes an authorization rule from the current policy.
func (c *Controller) RemovePolicy(rvals ...interface{}) (bool, error) {
	return c.enforcer.RemovePolicy(rvals...)
}
