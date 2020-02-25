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

package batch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	mesostype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-client/cmd/utils"
	"bk-bcs/bcs-services/bcs-client/pkg/metastream"
	"bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"bk-bcs/bcs-services/bcs-client/pkg/storage/v1"

	"github.com/urfave/cli"
)

//NewCleanCommand sub command clean all  registration
func NewCleanCommand() cli.Command {
	return cli.Command{
		Name:  "clean",
		Usage: "delete multiple Mesos resources, like application/deployment/service/configmap/secret/customresourcedefinition",
		UsageText: `
		example:
			> helm template myname sometemplate -n bcs-system | grep -v "^#" | bcs-client clean 
			or reading resource from file
			> bcs-client clean -f myresource.json
			> bcs-client clean -f anyyaml.yaml --format yaml
		`,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "reading all resources reference from `FILE`",
			},
			cli.StringFlag{
				Name:  "clusterid",
				Usage: "Cluster ID",
			},
			cli.StringFlag{
				Name:  "format",
				Usage: "resource format, like json or yaml",
				Value: metastream.JSONFormat,
			},
		},
		Action: func(c *cli.Context) error {
			return clean(utils.NewClientContext(c))
		},
	}
}

type deleteFunc func(clusterID, namespace, name string, enforce bool) error

//clean multiple mesos json resources to bcs-scheduler
func clean(cxt *utils.ClientContext) error {
	//step: check parameter from command line
	if err := cxt.MustSpecified(utils.OptionClusterID); err != nil {
		return err
	}
	var data []byte
	var err error
	if !cxt.IsSet(utils.OptionFile) {
		//reading all data from stdin
		data, err = ioutil.ReadAll(os.Stdin)
	} else {
		data, err = cxt.FileData()
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("failed to clean: no available resource datas")
	}
	//step: reading json object from input(file or stdin)
	metaList := metastream.NewMetaStream(bytes.NewReader(data), cxt.String("format"))
	if metaList.Length() == 0 {
		return fmt.Errorf("failed to clean: No correct format resource")
	}
	//step: initialize storage client & scheduler client
	storage := v1.NewBcsStorage(utils.GetClientOption())
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())

	//step: delete all resource according json list
	for metaList.HasNext() {
		info := metaInfo{}
		//step: check json object from parsing
		info.apiVersion, info.kind, err = metaList.GetResourceKind()
		if err != nil {
			fmt.Printf("apply partial failed, %s, continue...\n", err.Error())
			continue
		}
		info.namespace, info.name, err = metaList.GetResourceKey()
		if err != nil {
			fmt.Printf("apply partial failed, %s, continue...\n", err.Error())
			continue
		}
		//info.rawJson = metaList.GetRawJSON()
		info.clusterID = cxt.ClusterID()
		//step: inspect resource object from storage
		//	if resource exist, delete resource in bcs-scheduler, print object status from response
		//	otherwise, do nothing~
		var inspectStatus error
		var delFunc deleteFunc
		switch mesostype.BcsDataType(strings.ToLower(info.kind)) {
		case mesostype.BcsDataType_APP:
			_, inspectStatus = storage.InspectApplication(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteApplication
		case mesostype.BcsDataType_PROCESS:
			_, inspectStatus = storage.InspectProcess(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteProcess
		case mesostype.BcsDataType_SECRET:
			_, inspectStatus = storage.InspectSecret(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteSecret
		case mesostype.BcsDataType_CONFIGMAP:
			_, inspectStatus = storage.InspectConfigMap(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteConfigMap
		case mesostype.BcsDataType_SERVICE:
			_, inspectStatus = storage.InspectService(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteService
		case mesostype.BcsDataType_DEPLOYMENT:
			_, inspectStatus = storage.InspectDeployment(cxt.ClusterID(), info.namespace, info.name)
			delFunc = scheduler.DeleteDeployment
		case mesostype.BcsDataType_CRD:
			_, inspectStatus = scheduler.GetCustomResourceDefinition(cxt.ClusterID(), info.name)
			delFunc = func(cluster, namespace, name string, enforce bool) error {
				return scheduler.DeleteCustomResourceDefinition(cluster, name)
			}
		default:
			//unkown type, try custom resource
			crdapiVersion, plural, crdErr := utils.GetCustomResourceTypeByKind(scheduler, cxt.ClusterID(), info.kind)
			if err != nil {
				fmt.Printf("resource %s/%s %s clean failed, %s\n", info.apiVersion, info.kind, info.name, crdErr.Error())
				continue
			}
			_, inspectStatus = scheduler.GetCustomResource(cxt.ClusterID(), crdapiVersion, plural, info.namespace, info.name)
			delFunc = func(cluster, namespace, name string, enforce bool) error {
				return scheduler.DeleteCustomResource(cluster, crdapiVersion, plural, namespace, name)
			}
		}
		cleanSpecifiedResource(inspectStatus, delFunc, &info)
	}
	return nil
}

func cleanSpecifiedResource(inspectStatus error, delfunc deleteFunc, info *metaInfo) {
	if inspectStatus == nil {
		//no error when inspect, it means data exist
		if err := delfunc(info.clusterID, info.namespace, info.name, false); err != nil {
			fmt.Printf("resource %s/%s %s clean failed, %s\n", info.apiVersion, info.kind, info.name, err.Error())
		} else {
			fmt.Printf("resource %s/%s %s clean successfully\n", info.apiVersion, info.kind, info.name)
		}
		return
	}
	if isObjectNotExist(inspectStatus) {
		fmt.Printf("resource %s/%s %s clean nothing\n", info.apiVersion, info.kind, info.name)
	} else {
		fmt.Printf("resource %s/%s %s clean failed, %s\n", info.apiVersion, info.kind, info.name, inspectStatus.Error())
	}
}
