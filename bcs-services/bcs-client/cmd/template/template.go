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

package template

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"

	"github.com/urfave/cli"
)

func NewTemplateCommand() cli.Command {
	return cli.Command{
		Name:  "template",
		Usage: "get json templates of application, service and so on",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "type, t",
				Usage: "Template type, app/service/configmap/secret/deployment/agentsettings",
			},
		},
		Action: func(c *cli.Context) error {
			if err := template(utils.NewClientContext(c)); err != nil {
				return err
			}
			return nil
		},
	}
}

func template(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionType); err != nil {
		return err
	}

	resourceType := c.String(utils.OptionType)

	switch resourceType {
	case "app", "application":
		return getTemplate(applicationTemplate)
	case "configmap":
		return getTemplate(configMapTemplate)
	case "secret":
		return getTemplate(secretTemplate)
	case "service":
		return getTemplate(serviceTemplate)
	case "deploy", "deployment":
		return getTemplate(deploymentTemplate)
	case "as", "agentsettings":
		return getTemplate(agentSettingsTemplate)
	default:
		return fmt.Errorf("invalid type: %s", resourceType)
	}
}

const (
	applicationTemplate   = "{\"apiVersion\":\"v4\",\"kind\":\"application\",\"restartPolicy\":{\"policy\":\"Never\",\"interval\":5,\"backoff\":10},\"killPolicy\":{\"gracePeriod\":5},\"constraint\":{\"IntersectionItem\":[]},\"metadata\":{\"annotations\":{},\"labels\":{\"podname\":\"app-test\"},\"name\":\"app-test\",\"namespace\":\"defaultGroup\"},\"spec\":{\"instance\":1,\"template\":{\"spec\":{\"containers\":[{\"command\":\"/test/start.sh\",\"args\":[\"8899\"],\"parameters\":[],\"type\":\"MESOS\",\"env\":[],\"image\":\"docker_image:latest\",\"imagePullPolicy\":\"Always\",\"privileged\":false,\"ports\":[{\"containerPort\":8899,\"name\":\"test-port\",\"protocol\":\"HTTP\"}],\"healthChecks\":[],\"resources\":{\"limits\":{\"cpu\":\"0.1\",\"memory\":\"50\"}},\"volumes\":[],\"secrets\":[],\"configmaps\":[]}],\"networkMode\":\"BRIDGE\",\"networkType\":\"BRIDGE\"}}}}"
	configMapTemplate     = "{\"apiVersion\":\"v4\",\"kind\":\"configmap\",\"metadata\":{\"name\":\"configmap-test\",\"namespace\":\"defaultGroup\",\"labels\":{}},\"datas\":{\"item-one\":{\"type\":\"file\",\"content\":\"Y29uZmlnIGNvbnRleHQ=\"},\"item-two\":{\"type\":\"file\",\"content\":\"Y29uZmlnIGNvbnRleHQ=\"}}}"
	secretTemplate        = "{\"apiVersion\":\"v4\",\"kind\":\"secret\",\"metadata\":{\"name\":\"secret-test\",\"namespace\":\"defaultGroup\",\"labels\":{}},\"type\":\"\",\"datas\":{\"secret-subkey\":{\"path\":\"ECRET_ENV_TEST\",\"content\":\"Y29uZmlnIGNvbnRleHQ=\"}}}"
	serviceTemplate       = "{\"apiVersion\":\"v4\",\"kind\":\"service\",\"metadata\":{\"name\":\"service-test\",\"namespace\":\"defaultGroup\",\"labels\":{\"BCSGROUP\":\"external\"}},\"spec\":{\"selector\":{\"podname\":\"app-test\"},\"ports\":[{\"name\":\"test-port\",\"protocol\":\"tcp\",\"servicePort\":8889}]}}"
	deploymentTemplate    = "{\"apiVersion\":\"v4\",\"kind\":\"deployment\",\"metadata\":{\"labels\":{\"podname\":\"deployment-test\"},\"name\":\"deployment-test\",\"namespace\":\"defaultGroup\"},\"restartPolicy\":{\"policy\":\"Always\",\"interval\":5,\"backoff\":10},\"constraint\":{\"IntersectionItem\":[]},\"spec\":{\"instance\":2,\"selector\":{\"podname\":\"app-test\"},\"template\":{\"metadata\":{\"labels\":{},\"name\":\"deployment-test\",\"namespace\":\"defaultGroup\"},\"spec\":{\"containers\":[{\"command\":\"/test/start.sh\",\"args\":[\"8899\"],\"parameters\":[],\"type\":\"MESOS\",\"env\":[],\"image\":\"docker_image:latest\",\"imagePullPolicy\":\"Always\",\"privileged\":false,\"ports\":[{\"containerPort\":8899,\"name\":\"test-port\",\"protocol\":\"HTTP\"}],\"healthChecks\":[],\"resources\":{\"limits\":{\"cpu\":\"0.1\",\"memory\":\"50\"}},\"volumes\":[],\"secrets\":[],\"configmaps\":[]}],\"networkMode\":\"BRIDGE\",\"networkType\":\"BRIDGE\"}},\"strategy\":{\"type\":\"RollingUpdate\",\"rollingupdate\":{\"maxUnavilable\":1,\"maxSurge\":1,\"upgradeDuration\":60,\"autoUpgrade\":false,\"rollingOrder\":\"CreateFirst\",\"pause\":false}}}}"
	agentSettingsTemplate = "[{\"innerIP\":\"127.0.0.1\",\"disabled\":true,\"strings\":{\"attr1\":{\"value\":\"hahaha\"}},\"scalars\":{\"foo\":{\"value\":0.01}}}]"
)

func getTemplate(template string) error {
	var data interface{}
	if err := codec.DecJson([]byte(template), &data); err != nil {
		return fmt.Errorf("decode template error: %v", err)
	}

	fmt.Printf("%s\n", utils.TryIndent(data))
	return nil
}
