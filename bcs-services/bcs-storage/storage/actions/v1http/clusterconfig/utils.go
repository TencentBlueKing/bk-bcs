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

package clusterconfig

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
)

func getService(req *restful.Request) string {
	for _, service := range []string{req.PathParameter(serviceTag), req.QueryParameter(serviceTag)} {
		if service != "" {
			return service
		}
	}
	return "test"
}

func getSvcCondition(req *restful.Request) *operator.Condition {
	return operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: getService(req)})
}

func getTemplateCondition(req *restful.Request) *operator.Condition {
	return operator.NewLeafCondition(operator.Eq, operator.M{serviceTag: getService(req)})
}

func getClsCondition(req *restful.Request) *operator.Condition {
	clusterID := lib.GetQueryParamString(req, clusterIdTag)
	features := operator.M{clusterIdTag: clusterID}
	return operator.NewLeafCondition(operator.Eq, features)
}

func getCls(req *restful.Request) ([]operator.M, error) {
	condition := getClsCondition(req)
	mList, err := GetClusterInfo(req.Request.Context(), condition)
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
	mList, err := GetClusterInfo(req.Request.Context(), condition)
	if err != nil {
		return nil, err
	}
	return mList, nil
}

func getSvcSet(ctx context.Context, resourceType string, opt *lib.StoreGetOption) (
	svcConfigSet *types.ConfigSet, err error) {
	var svcConfig []operator.M
	if svcConfig, err = GetData(ctx, resourceType, opt); err != nil || len(svcConfig) == 0 {
		if err == nil {
			err = storageErr.ServiceConfigNoFound
		}
		return
	}

	svcConfigRaw := svcConfig[0]
	svcConfigRawData := svcConfigRaw[dataTag]
	if svcConfigSet, err = types.ParseConfigSet(svcConfigRawData); err != nil {
		blog.Errorf("Failed to parse service configSet. err: %v", err)
		return
	}
	return svcConfigSet, nil
}

func getClsSet(clsConfig []operator.M) (clusterSet []types.ClusterSet, err error) {
	var clsConfigSet *types.ConfigSet
	clusterSet = make([]types.ClusterSet, 0, len(clsConfig))
	for _, clusterRaw := range clsConfig {
		clsConfigRawID := clusterRaw[clusterIdTag]
		clsConfigRawData := clusterRaw[dataTag]

		if clsConfigSet, err = types.ParseConfigSet(clsConfigRawData); err != nil {
			return nil, err
		}

		clusterID, _ := clsConfigRawID.(string)
		clusterSet = append(clusterSet, types.ClusterSet{ClusterId: clusterID, ClusterConfig: *clsConfigSet})
	}
	return clusterSet, err
}

func generateData(req *restful.Request, clsFunc func(req *restful.Request) ([]operator.M, error)) (
	config *types.DeployConfig, err error) {
	var clsConfig []operator.M
	// service name
	service := getService(req)
	// option
	opt := &lib.StoreGetOption{
		Cond: getSvcCondition(req),
	}
	// 获取 cls config
	if clsConfig, err = clsFunc(req); err != nil {
		return &types.DeployConfig{}, err
	}

	return GenerateData(req.Request.Context(), opt, clsConfig, service)
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

	template, err := GetTemplate(req.Request.Context(), getTemplateCondition(req))
	if err != nil {
		return nil, err
	}
	str := renderConfig.Render(template)

	r := lib.CopyMap(operator.M{clusterIdTag: clusterID})
	blog.Infof("renderConfig data: %s", str)

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
	// option
	opt := &lib.StorePutOption{
		Cond:          getClsCondition(req),
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}

	return SaveClusterInfoConfig(req.Request.Context(), data, opt)
}

func putStableVersion(req *restful.Request) error {
	version, err := getVerData(req)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to get version data. err %s", err.Error())
		blog.Errorf(errMsg)
		return fmt.Errorf(errMsg)
	}
	service := getService(req)
	if err = SaveStableVersion(req.Request.Context(), service, version); err != nil {
		blog.Errorf("Failed to set stable version of %s. err %s", service, err.Error())
		return err
	}
	return nil
}

func urlPath(oldURL string) string {
	return urlPrefix + oldURL
}

func wrapIP(s []string, df string) (r []string) {
	r = make([]string, 0, len(s))
	for _, v := range s {
		if !strings.Contains(v, ":") {
			v += ":" + df
		}
		r = append(r, v)
	}
	return r
}

func unwrapIP(s []string) (r []string) {
	r = make([]string, 0, len(s))
	for _, v := range s {
		if strings.Contains(v, ":") {
			v = strings.Split(v, ":")[0]
		}
		r = append(r, v)
	}
	return r
}
