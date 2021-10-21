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

package clusterconfig

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"

	"github.com/emicklei/go-restful"
)

func getService(req *restful.Request) string {
	service := req.PathParameter(serviceTag)
	if service == "" {
		service = req.QueryParameter(serviceTag)
		if service == "" {
			service = "test"
		}
	}
	return service
}

func getSvcCondition(req *restful.Request) *operator.Condition {
	return operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: getService(req)})
}

func getTemplateCondition(req *restful.Request) *operator.Condition {
	return operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: getService(req)})
}

func getTemplate(req *restful.Request) (string, error) {
	condition := getTemplateCondition(req)
	getOption := &lib.StoreGetOption{
		Cond: condition,
		Sort: map[string]int{
			versionTag: -1,
		},
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableTpl, getOption)
	if err != nil {
		return "", err
	}
	if len(mList) == 0 {
		err := storageErr.ConfigTemplateNoFound
		blog.Errorf("%s", err.Error())
		return "", err
	}
	vs, _ := mList[0][dataTag]
	s, ok := vs.(string)
	if !ok {
		err := storageErr.ConfigTemplateInvalid
		blog.Errorf("%s", err.Error())
		return "", err
	}
	return s, nil
}

func getStableVersion(req *restful.Request) (string, error) {
	condition := operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: getService(req)})
	getOption := &lib.StoreGetOption{
		Cond: condition,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableVer, getOption)
	if err != nil {
		return "", err
	}
	vs, _ := mList[0][dataTag]
	s, ok := vs.(string)
	if !ok {
		err := storageErr.StableVersionInvalid
		blog.Errorf("%v", err)
		return "", err
	}
	return s, nil
}

func getClsCondition(req *restful.Request) *operator.Condition {
	clusterID := lib.GetQueryParamString(req, clusterIdTag)
	features := operator.M{clusterIdTag: clusterID}
	return operator.NewLeafCondition(operator.Eq, features)
}

func getCls(req *restful.Request) ([]operator.M, error) {
	condition := getClsCondition(req)
	getOption := &lib.StoreGetOption{
		Cond: condition,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableCls, getOption)
	if err != nil {
		return nil, err
	}
	return mList, nil
}

func getMultiClsCondition(req *restful.Request) *operator.Condition {
	var condList []*operator.Condition
	if clusterIDNot := lib.GetQueryParamString(req, clusterIdNotTag); clusterIDNot != "" {
		condList = append(condList, operator.NewLeafCondition(
			operator.Nin, operator.M{clusterIdTag: strings.Split(clusterIDNot, ",")}))
	} else if clusterID := lib.GetQueryParamString(req, clusterIdTag); clusterID != "" {
		condList = append(condList, operator.NewLeafCondition(
			operator.In, operator.M{clusterIdTag: strings.Split(clusterID, ",")}))
	}
	clsCondition := operator.NewBranchCondition(operator.And, condList...)
	return clsCondition
}

func getMultiCls(req *restful.Request) ([]operator.M, error) {
	condition := getMultiClsCondition(req)
	getOption := &lib.StoreGetOption{
		Cond: condition,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableCls, getOption)
	if err != nil {
		return nil, err
	}
	return mList, nil
}

func getSvc(req *restful.Request) ([]operator.M, error) {
	condition := getSvcCondition(req)
	getOption := &lib.StoreGetOption{
		Cond: condition,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableSvc, getOption)
	if err != nil {
		return nil, err
	}
	return mList, nil
}

func getSvcSet(req *restful.Request) (svcConfigSet *types.ConfigSet, err error) {
	var svcConfig []operator.M
	if svcConfig, err = getSvc(req); err != nil || len(svcConfig) == 0 {
		if err == nil {
			err = storageErr.ServiceConfigNoFound
		}
		return
	}

	svcConfigRaw := svcConfig[0]
	svcConfigRawData, _ := svcConfigRaw[dataTag]
	if svcConfigSet, err = types.ParseConfigSet(svcConfigRawData); err != nil {
		blog.Errorf("Failed to parse service configSet. err: %v", err)
		return
	}
	return
}

func getClsSet(req *restful.Request, clsFunc func(req *restful.Request) ([]operator.M, error)) (
	[]types.ClusterSet, error) {
	clsConfig := make([]operator.M, 0)
	var err error
	if clsConfig, err = clsFunc(req); err != nil {
		return nil, err
	}
	var clsConfigSet *types.ConfigSet
	clusterSet := make([]types.ClusterSet, 0, len(clsConfig))
	for _, clusterRaw := range clsConfig {
		clsConfigRawID, _ := clusterRaw[clusterIdTag]
		clsConfigRawData, _ := clusterRaw[dataTag]

		if clsConfigSet, err = types.ParseConfigSet(clsConfigRawData); err != nil {
			return nil, err
		}

		clusterID, _ := clsConfigRawID.(string)
		clusterSet = append(clusterSet, types.ClusterSet{ClusterId: clusterID, ClusterConfig: *clsConfigSet})
	}
	return clusterSet, err
}

func generateData(
	req *restful.Request,
	clsFunc func(req *restful.Request) ([]operator.M, error)) (types.DeployConfig, error) {

	var svcConfigSet *types.ConfigSet
	var clsConfigSet []types.ClusterSet
	var stableVersion string

	var err error
	if svcConfigSet, err = getSvcSet(req); err != nil {
		return types.DeployConfig{}, err
	}
	if clsConfigSet, err = getClsSet(req, clsFunc); err != nil {
		return types.DeployConfig{}, err
	}
	if stableVersion, err = getStableVersion(req); err != nil {
		return types.DeployConfig{}, err
	}

	config := types.DeployConfig{
		Service:       getService(req),
		ServiceConfig: *svcConfigSet,
		Clusters:      clsConfigSet,
		StableVersion: stableVersion,
	}
	return config, nil
}

func getReqData(req *restful.Request) (operator.M, error) {
	tmp := types.BcsStorageClusterIf{NeedNat: true}
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}

	var renderConfig types.RenderConfig
	zk := wrapIP(tmp.ZkIp, "2181")
	dns := wrapIP(tmp.DnsIp, "53")
	clusterID := lib.GetQueryParamString(req, clusterIdTag)
	lin := strings.Split(clusterID, "-")

	renderConfig.MesosZk = strings.Join(zk, ",")
	renderConfig.MesosZkSpace = strings.Join(zk, " ")
	renderConfig.MesosZkSemicolon = strings.Join(zk, ";")
	renderConfig.MesosZkRaw = strings.Join(unwrapIP(zk), ",")
	renderConfig.MesosMaster = strings.Join(tmp.MasterIp, ",")
	renderConfig.MesosQuorum = strconv.Itoa((len(tmp.MasterIp) + 1) / 2)
	renderConfig.Dns = strings.Join(dns, " ")
	renderConfig.ClusterId = clusterID
	renderConfig.ClusterIdNum = lin[len(lin)-1]
	renderConfig.City = tmp.City
	renderConfig.JfrogUrl = tmp.JfrogUrl
	renderConfig.NeedNat = func() string {
		if tmp.NeedNat {
			return "true"
		}
		return "false"
	}()

	template, err := getTemplate(req)
	if err != nil {
		return nil, err
	}
	str := renderConfig.Render(template)

	r := lib.CopyMap(operator.M{clusterIdTag: clusterID})
	blog.Infof(str)
	var data map[string]interface{}
	err = codec.DecJson([]byte(str), &data)
	if err != nil {
		return nil, err
	}
	r[dataTag] = data
	return r, nil
}

func getVerData(req *restful.Request) (string, error) {
	var tmp types.BcsStorageStableVersionIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return "", err
	}
	return tmp.Version, nil
}

func putClsConfig(req *restful.Request) error {

	data, err := getReqData(req)
	if err != nil {
		return err
	}

	condition := getClsCondition(req)
	putOption := &lib.StorePutOption{
		Cond:          condition,
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}

	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	err = store.Put(req.Request.Context(), tableCls, data, putOption)
	if err != nil {
		return err
	}
	return nil
}

func putStableVersion(req *restful.Request) error {
	version, err := getVerData(req)
	if err != nil {
		blog.Errorf("Failed to get version data. err %s", err.Error())
		return fmt.Errorf("Failed to get version data. err %s", err.Error())
	}

	service := getService(req)
	condition := operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: service})
	putOption := &lib.StorePutOption{
		Cond: condition,
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	if err := store.Put(req.Request.Context(), tableVer, operator.M{dataTag: version}, putOption); err != nil {
		blog.Errorf("Failed to set stable version of %s. err %s", service, err.Error())
		return err
	}
	return nil
}

func urlPath(oldURL string) string {
	return urlPrefix + oldURL
}

func wrapIP(s []string, df string) (r []string) {
	r = []string{}
	for _, v := range s {
		if !strings.Contains(v, ":") {
			v += ":" + df
		}
		r = append(r, v)
	}
	return
}

func unwrapIP(s []string) (r []string) {
	r = []string{}
	for _, v := range s {
		if strings.Contains(v, ":") {
			v = strings.Split(v, ":")[0]
		}
		r = append(r, v)
	}
	return
}
