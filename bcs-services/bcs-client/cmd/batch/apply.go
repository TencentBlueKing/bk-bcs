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
	"net/url"
	"os"
	"strings"

	mesostype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-client/cmd/utils"
	"bk-bcs/bcs-services/bcs-client/pkg/metastream"
	"bk-bcs/bcs-services/bcs-client/pkg/scheduler/v4"
	"bk-bcs/bcs-services/bcs-client/pkg/storage/v1"

	"github.com/urfave/cli"
)

//NewApplyCommand sub command apply registration
func NewApplyCommand() cli.Command {
	return cli.Command{
		Name:  "apply",
		Usage: "create multiple Mesos resources, like application/deployment/service/configmap/secret",
		UsageText: `
		example:
			> helm template myname sometemplate -n bcs-system | grep -v "^#" | bcs-client apply 
			or reading resource from file
			> bcs-client apply -f myresource.json
			> bcs-client apply -f anyyaml.yaml --format yaml
		`,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "from-file, f",
				Usage: "reading with configuration `FILE`",
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
			return apply(utils.NewClientContext(c))
		},
	}
}

type createFunc func(string, string, []byte) error
type updateFunc func(string, string, []byte, url.Values) error

type metaInfo struct {
	clusterID  string
	apiVersion string
	kind       string
	namespace  string
	name       string
	rawJson    []byte
}

//apply multiple mesos json resources to bcs-scheduler
func apply(cxt *utils.ClientContext) error {
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
		return fmt.Errorf("failed to apply: no available resource datas")
	}
	//step: reading json object from input(file or stdin)
	metaList := metastream.NewMetaStream(bytes.NewReader(data), cxt.String("format"))
	if metaList.Length() == 0 {
		return fmt.Errorf("failed to Apply: No correct format resource")
	}
	//step: initialize storage client & scheduler client
	storage := v1.NewBcsStorage(utils.GetClientOption())
	scheduler := v4.NewBcsScheduler(utils.GetClientOption())

	//step: create/update all resource according json list
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
		info.rawJson = metaList.GetRawJSON()
		utils.DebugPrintf("debugInfo: %s\n", string(info.rawJson))
		info.clusterID = cxt.ClusterID()
		//step: inspect resource object from storage
		//	if resource exist, update resource to bcs-scheduler, print object status from response
		//	otherwise, create resources to bcs-scheduler and print object status from response
		var inspectStatus error
		var create createFunc
		var update updateFunc
		switch mesostype.BcsDataType(strings.ToLower(info.kind)) {
		case mesostype.BcsDataType_APP:
			_, inspectStatus = storage.InspectApplication(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateApplication
			update = scheduler.UpdateApplication
		case mesostype.BcsDataType_PROCESS:
			_, inspectStatus = storage.InspectProcess(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateProcess
			update = scheduler.UpdateProcess
		case mesostype.BcsDataType_SECRET:
			_, inspectStatus = storage.InspectSecret(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateSecret
			update = scheduler.UpdateSecret
		case mesostype.BcsDataType_CONFIGMAP:
			_, inspectStatus = storage.InspectConfigMap(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateConfigMap
			update = scheduler.UpdateConfigMap
		case mesostype.BcsDataType_SERVICE:
			_, inspectStatus = storage.InspectService(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateService
			update = scheduler.UpdateService
		case mesostype.BcsDataType_DEPLOYMENT:
			_, inspectStatus = storage.InspectDeployment(cxt.ClusterID(), info.namespace, info.name)
			create = scheduler.CreateDeployment
			update = scheduler.UpdateDeployment
		case mesostype.BcsDataType_CRD:
			_, inspectStatus = scheduler.GetCustomResourceDefinition(cxt.ClusterID(), info.name)
			create = func(cluster string, ns string, data []byte) error {
				return scheduler.CreateCustomResourceDefinition(cluster, data)
			}
			update = func(cluster, ns string, data []byte, urlv url.Values) error {
				return scheduler.UpdateCustomResourceDefinition(cluster, info.name, data)
			}
		default:
			//unkown type, try custom resource
			crdapiVersion, plural, crdErr := utils.GetCustomResourceTypeByKind(scheduler, cxt.ClusterID(), info.kind)
			if err != nil {
				fmt.Printf("resource %s/%s %s apply failed, %s\n", info.apiVersion, info.kind, info.name, crdErr.Error())
				continue
			}
			_, inspectStatus = scheduler.GetCustomResource(cxt.ClusterID(), crdapiVersion, plural, info.namespace, info.name)
			create = func(cluster string, ns string, data []byte) error {
				return scheduler.CreateCustomResource(cluster, crdapiVersion, plural, ns, data)
			}
			update = func(cluster, ns string, data []byte, urlv url.Values) error {
				return scheduler.UpdateCustomResource(cluster, crdapiVersion, plural, ns, info.name, data)
			}
		}
		applySpecifiedResource(inspectStatus, create, update, &info)
	}
	return nil
}

func isObjectNotExist(err error) bool {
	str := err.Error()
	return strings.Contains(str, "resource does not exist") || strings.Contains(str, "not found")
}

func applySpecifiedResource(inspectStatus error, create createFunc, update updateFunc, info *metaInfo) {
	if inspectStatus == nil {
		//update object
		if err := update(info.clusterID, info.namespace, info.rawJson, nil); err != nil {
			fmt.Printf("resource %s/%s %s update successfully\n", info.apiVersion, info.kind, info.name)
		} else {
			fmt.Printf("resource %s/%s %s update failed, %s\n", info.apiVersion, info.kind, info.name, err.Error())
		}
		return
	}

	if !isObjectNotExist(inspectStatus) {
		fmt.Printf("resource %s/%s %s apply failed, %s\n", info.apiVersion, info.kind, info.name, inspectStatus.Error())
		return
	}
	//create
	if err := create(info.clusterID, info.namespace, info.rawJson); err != nil {
		fmt.Printf("resource %s/%s %s create failed, %s\n", info.apiVersion, info.kind, info.name, err.Error())
	} else {
		fmt.Printf("resource %s/%s %s create successfully\n", info.apiVersion, info.kind, info.name)
	}
}
