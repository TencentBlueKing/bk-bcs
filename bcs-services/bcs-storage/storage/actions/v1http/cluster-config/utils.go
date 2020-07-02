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

package clusterConfig

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

type reqConfig struct {
	req  *restful.Request
	tank operator.Tank

	tableSvc string
	tableCls string
	tableVer string
	tableTpl string

	svcCondition  *operator.Condition
	clsCondition  *operator.Condition
	tplCondition  *operator.Condition
	data          operator.M
	features      operator.M
	service       string
	clusterId     string
	template      string
	stableVersion string
	stableVerData string
}

func newReqConfig(req *restful.Request) *reqConfig {
	return &reqConfig{
		req:      req,
		tank:     getNewTank(),
		tableSvc: tableSvc,
		tableCls: tableCls,
		tableVer: tableVer,
		tableTpl: tableTpl,
	}
}

func (rc *reqConfig) reset() {
	rc.svcCondition = nil
	rc.clsCondition = nil
	rc.tplCondition = nil
	rc.data = nil
	rc.features = nil
	rc.clusterId = ""
	rc.template = ""
	rc.stableVersion = ""
	rc.stableVerData = ""
}

func (rc *reqConfig) getServiceTable() string {
	return rc.tableSvc
}

func (rc *reqConfig) getClusterTable() string {
	return rc.tableCls
}

func (rc *reqConfig) getStableVersionTable() string {
	return rc.tableVer
}

func (rc *reqConfig) getTemplateTable() string {
	return rc.tableTpl
}

func (rc *reqConfig) getService() string {
	if rc.service == "" {
		service := rc.req.PathParameter(serviceTag)
		if service == "" {
			service = rc.req.QueryParameter(serviceTag)
			if service == "" {
				service = "test"
			}
		}
		rc.service = service
	}
	return rc.service
}

func (rc *reqConfig) getClsFeat() *operator.Condition {
	if rc.clsCondition == nil {
		clusterId := rc.req.PathParameter(clusterIdTag)
		features := operator.M{clusterIdTag: clusterId}

		rc.clusterId = clusterId
		rc.features = features
		rc.clsCondition = operator.NewCondition(operator.Eq, features)
	}
	return rc.clsCondition
}

func (rc *reqConfig) getMultiClsFeat() *operator.Condition {
	if rc.clsCondition == nil {
		condition := operator.BaseCondition
		if clusterIdNot := rc.req.QueryParameter(clusterIdNotTag); clusterIdNot != "" {
			condition = condition.AddOp(operator.Nin, clusterIdTag, strings.Split(clusterIdNot, ","))
		} else if clusterId := rc.req.QueryParameter(clusterIdTag); clusterId != "" {
			condition = condition.AddOp(operator.In, clusterIdTag, strings.Split(clusterId, ","))
		}
		rc.clsCondition = condition
	}
	return rc.clsCondition
}

func (rc *reqConfig) getSvcFeat() *operator.Condition {
	if rc.svcCondition == nil {
		rc.svcCondition = operator.BaseCondition.AddOp(operator.Eq, serviceTag, rc.getService())
	}
	return rc.svcCondition
}

func (rc *reqConfig) getTemplateFeat() *operator.Condition {
	if rc.tplCondition == nil {
		rc.tplCondition = operator.BaseCondition.AddOp(operator.Eq, serviceTag, rc.getService())
	}
	return rc.tplCondition
}

func (rc *reqConfig) getTemplate() (string, error) {
	if rc.template == "" {
		tank := rc.tank.From(rc.getTemplateTable()).Filter(rc.getTemplateFeat()).OrderBy("-" + versionTag).Query()
		if err := tank.GetError(); err != nil {
			blog.Errorf("Failed to get config template. err: %v", err)
			return "", err
		}

		r := tank.GetValue()
		if len(r) == 0 {
			err := storageErr.ConfigTemplateNoFound
			blog.Errorf("%v", err)
			return "", err
		}

		vm, _ := r[0].(map[string]interface{})
		vs, _ := vm[dataTag]
		s, ok := vs.(string)
		if !ok {
			err := storageErr.ConfigTemplateInvalid
			blog.Errorf("%v", err)
			return "", err
		}
		rc.template = s
	}
	return rc.template, nil
}

func (rc *reqConfig) getStableVersion() (string, error) {
	if rc.stableVersion == "" {
		service := rc.getService()
		tank := rc.tank.From(rc.getStableVersionTable()).
			Filter(operator.BaseCondition.AddOp(operator.Eq, serviceTag, service)).Query()

		if err := tank.GetError(); err != nil {
			blog.Errorf("Failed to get stable version of %s. err: %v", service, err)
			return "", err
		}

		r := tank.GetValue()
		if len(r) == 0 {
			err := storageErr.StableVersionNoFound
			blog.Errorf("%v", err)
			return "", err
		}

		vm, _ := r[0].(map[string]interface{})
		vs, _ := vm[dataTag]
		s, ok := vs.(string)
		if !ok {
			err := storageErr.StableVersionInvalid
			blog.Errorf("%v", err)
			return "", err
		}
		rc.stableVersion = s
	}
	return rc.stableVersion, nil
}

func (rc *reqConfig) getCls() ([]interface{}, error) {
	return rc.get(rc.getClusterTable(), rc.getClsFeat())
}

func (rc *reqConfig) getMultiCls() ([]interface{}, error) {
	return rc.get(rc.getClusterTable(), rc.getMultiClsFeat())
}

func (rc *reqConfig) getSvc() ([]interface{}, error) {
	return rc.get(rc.getServiceTable(), rc.getSvcFeat())
}

func (rc *reqConfig) get(table string, condition *operator.Condition) (r []interface{}, err error) {
	tank := rc.tank.From(table).Filter(condition).Query()
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return
	}

	r = tank.GetValue()
	return
}

func (rc *reqConfig) getSvcSet() (svcConfigSet *types.ConfigSet, err error) {
	var svcConfig []interface{}
	if svcConfig, err = rc.getSvc(); err != nil || len(svcConfig) == 0 {
		if err == nil {
			err = storageErr.ServiceConfigNoFound
		}
		return
	}

	svcConfigRaw, _ := svcConfig[0].(map[string]interface{})
	svcConfigRawData, _ := svcConfigRaw[dataTag]
	if svcConfigSet, err = types.ParseConfigSet(svcConfigRawData); err != nil {
		blog.Errorf("Failed to parse service configSet. err: %v", err)
		return
	}
	return
}

func (rc *reqConfig) getClsSet(clsFunc func() ([]interface{}, error)) (clusterSet []types.ClusterSet, err error) {
	var clsConfig []interface{}
	if clsConfig, err = clsFunc(); err != nil {
		return
	}
	var clsConfigSet *types.ConfigSet
	clusterSet = make([]types.ClusterSet, 0, len(clsConfig))
	for _, clusterRaw := range clsConfig {
		clsConfigRaw, _ := clusterRaw.(map[string]interface{})
		clsConfigRawId, _ := clsConfigRaw[clusterIdTag]
		clsConfigRawData, _ := clsConfigRaw[dataTag]

		if clsConfigSet, err = types.ParseConfigSet(clsConfigRawData); err != nil {
			fmt.Println(err.Error())
			return
		}

		clusterId, _ := clsConfigRawId.(string)
		clusterSet = append(clusterSet, types.ClusterSet{ClusterId: clusterId, ClusterConfig: *clsConfigSet})
	}
	return
}

func (rc *reqConfig) generateData(clsFunc func() ([]interface{}, error)) (config types.DeployConfig, err error) {
	var svcConfigSet *types.ConfigSet
	var clsConfigSet []types.ClusterSet
	var stableVersion string

	if svcConfigSet, err = rc.getSvcSet(); err != nil {
		return
	}
	rc.reset()
	if clsConfigSet, err = rc.getClsSet(clsFunc); err != nil {
		return
	}
	rc.reset()
	if stableVersion, err = rc.getStableVersion(); err != nil {
		return
	}

	config = types.DeployConfig{
		Service:       rc.getService(),
		ServiceConfig: *svcConfigSet,
		Clusters:      clsConfigSet,
		StableVersion: stableVersion,
	}
	return
}

func (rc *reqConfig) getReqData() (operator.M, error) {
	if rc.data == nil {
		tmp := types.BcsStorageClusterIf{NeedNat: true}
		if err := codec.DecJsonReader(rc.req.Request.Body, &tmp); err != nil {
			return nil, err
		}

		var renderConfig types.RenderConfig
		zk := wrapIp(tmp.ZkIp, "2181")
		dns := wrapIp(tmp.DnsIp, "53")
		clusterId := rc.clusterId
		lin := strings.Split(clusterId, "-")

		rc.service = tmp.Service
		renderConfig.MesosZk = strings.Join(zk, ",")
		renderConfig.MesosZkSpace = strings.Join(zk, " ")
		renderConfig.MesosZkSemicolon = strings.Join(zk, ";")
		renderConfig.MesosZkRaw = strings.Join(unwrapIp(zk), ",")
		renderConfig.MesosMaster = strings.Join(tmp.MasterIp, ",")
		renderConfig.MesosQuorum = strconv.Itoa((len(tmp.MasterIp) + 1) / 2)
		renderConfig.Dns = strings.Join(dns, " ")
		renderConfig.ClusterId = clusterId
		renderConfig.ClusterIdNum = lin[len(lin)-1]
		renderConfig.City = tmp.City
		renderConfig.JfrogUrl = tmp.JfrogUrl
		renderConfig.NeedNat = func() string {
			if tmp.NeedNat {
				return "true"
			}
			return "false"
		}()

		template, err := rc.getTemplate()
		if err != nil {
			return nil, err
		}
		str := renderConfig.Render(template)

		r := lib.CopyMap(rc.features)
		blog.Infof(str)
		var data map[string]interface{}
		err = codec.DecJson([]byte(str), &data)
		if err != nil {
			return nil, err
		}
		r[dataTag] = data
		rc.data = r
	}
	return rc.data, nil
}

func (rc *reqConfig) getVerData() (string, error) {
	if rc.stableVerData == "" {
		var tmp types.BcsStorageStableVersionIf
		if err := codec.DecJsonReader(rc.req.Request.Body, &tmp); err != nil {
			return "", err
		}

		rc.stableVerData = tmp.Version
	}
	return rc.stableVerData, nil
}

func (rc *reqConfig) putClsConfig() (err error) {
	tank := rc.tank.From(rc.getClusterTable()).Filter(rc.getClsFeat())

	data, err := rc.getReqData()
	if err != nil {
		return
	}

	// Update or insert
	timeNow := time.Now()

	queryTank := tank.Query()
	if err = queryTank.GetError(); err != nil {
		blog.Errorf("Failed to check if resource exist. err: %v", err)
		return
	}
	if queryTank.GetLen() == 0 {
		data.Update(createTimeTag, timeNow)
	}

	tank = tank.Index(indexKeys...).Upsert(data.Update(updateTimeTag, timeNow))
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to update. err: %v", err)
		return
	}
	return
}

func (rc *reqConfig) putStableVersion() (err error) {
	version, err := rc.getVerData()
	if err != nil {
		blog.Errorf("Failed to get version data. err: %v", err)
		return
	}

	service := rc.getService()
	tank := rc.tank.From(rc.getStableVersionTable()).
		Filter(operator.BaseCondition.AddOp(operator.Eq, serviceTag, service)).
		Update(operator.M{dataTag: version})

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to set stable version of %s. err: %v", service, err)
	}
	return
}

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (rc *reqConfig) exit() {
	if rc.tank != nil {
		rc.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return urlPrefix + oldUrl
}

func wrapIp(s []string, df string) (r []string) {
	r = []string{}
	for _, v := range s {
		if !strings.Contains(v, ":") {
			v += ":" + df
		}
		r = append(r, v)
	}
	return
}

func unwrapIp(s []string) (r []string) {
	r = []string{}
	for _, v := range s {
		if strings.Contains(v, ":") {
			v = strings.Split(v, ":")[0]
		}
		r = append(r, v)
	}
	return
}
