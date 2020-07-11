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

package task

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"

	proto "github.com/golang/protobuf/proto"
)

var BcsContainerExecutorPath string
var BcsProcessExecutorPath string
var BcsCniDir string
var NetImage string
var Passwd = static.BcsDefaultPasswd
var User = static.BcsDefaultUser

//init mesos executor info
func InitExecutorInfo(CExec, PExec, CniDir, netImage string) {
	if CExec != "" {
		BcsContainerExecutorPath = CExec
	} else {
		BcsContainerExecutorPath = "file:///data/bcs/mesos/libexec/mesos/bcs-container-executor"
	}

	if PExec != "" {
		BcsProcessExecutorPath = PExec
	} else {
		BcsProcessExecutorPath = "file:///data/bcs/mesos/libexec/mesos/bcs-process-executor"
	}

	if CniDir != "" {
		BcsCniDir = CniDir
	}

	if netImage != "" {
		NetImage = netImage
	}
}

//CreateBcsExecutorInfo special the bcs executor
func CreateBcsExecutorInfo(offer *mesos.Offer /*cmdOrUri string,*/, taskGroupID *string,
	version *types.Version, store store.Store) *mesos.ExecutorInfo {

	var cmdOrUri string
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		cmdOrUri = BcsProcessExecutorPath
	case commtypes.BcsDataType_APP, "", commtypes.BcsDataType_Daemonset:
		cmdOrUri = BcsContainerExecutorPath
	}

	pathSplit := strings.Split(cmdOrUri, "/")

	var base string

	if len(pathSplit) > 0 {
		base = pathSplit[len(pathSplit)-1]
	} else {
		base = cmdOrUri
	}

	// added  20180814, if it is process, then call the Bcs_Process_Executor and pass the user-pwd configuration
	var execCommand string
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		execCommand = fmt.Sprintf("./%s", base)
	case commtypes.BcsDataType_APP, "", commtypes.BcsDataType_Daemonset:
		var user string
		var passwd string
		if version.Container[0].Docker.ImagePullUser != "" {
			userConf := version.Container[0].Docker.ImagePullUser
			//user in secret
			if strings.HasPrefix(userConf, "secret::") {
				userConfSplit := strings.Split(userConf, "::")
				if len(userConfSplit) != 2 {
					blog.Error("ImagePullUser sercret config format(%s) error, version(%s.%s.%s)",
						userConf, version.RunAs, version.ID, version.Name)
					return nil
				}

				userStr := userConfSplit[1]
				userStrSplit := strings.Split(userStr, "||")
				if len(userStrSplit) != 2 {
					blog.Error("ImagePullUser sercret config format(%s) error, version(%s.%s.%s)",
						userStr, version.RunAs, version.ID, version.Name)
					return nil
				}

				secretName := strings.TrimSpace(userStrSplit[0])
				secretKey := strings.TrimSpace(userStrSplit[1])
				secretNs := version.RunAs
				blog.Infof("to get user from secret(%s.%s::%s)", secretNs, secretName, secretKey)
				bcsSecret, err := store.FetchSecret(secretNs, secretName)
				if err != nil {
					blog.Error("get bcssecret(%s.%s) err: %s", secretNs, secretName, err.Error())
					return nil
				}
				if bcsSecret == nil {
					blog.Error("bcssecret(%s.%s) not exist", secretNs, secretName)
					return nil
				}
				bcsSecretItem, ok := bcsSecret.Data[secretKey]
				if ok == false {
					blog.Error("bcssecret item(key:%s) not exist in bcssecret(%s.%s)",
						secretKey, secretNs, secretName)
					return nil
				}

				userBase := strings.TrimSpace(bcsSecretItem.Content)
				if userBase != "" {
					userScrt, err := base64.StdEncoding.DecodeString(userBase)
					if err != nil {
						blog.Error("Decode base64(%s) err: %s", userBase, err.Error())
						return nil
					}

					user = string(userScrt)

				} else {
					blog.Infof("user in secret(%s.%s::%s) is empty", secretNs, secretName, secretKey)
					user = ""
				}
				//raw user
			} else {
				user = userConf
			}
			//default user
		} else {
			user = User
		}

		if version.Container[0].Docker.ImagePullPasswd != "" {
			passwdConf := version.Container[0].Docker.ImagePullPasswd
			//passwd in secret
			if strings.HasPrefix(passwdConf, "secret::") {
				passwdConfSplit := strings.Split(passwdConf, "::")
				if len(passwdConfSplit) != 2 {
					blog.Error("ImagePullPasswd sercret config format(%s) error, version(%s.%s.%s)",
						passwdConf, version.RunAs, version.ID, version.Name)
					return nil
				}

				passwdStr := passwdConfSplit[1]
				passwdStrSplit := strings.Split(passwdStr, "||")
				if len(passwdStrSplit) != 2 {
					blog.Error("ImagePullPasswd sercret config format(%s) error, version(%s.%s.%s)",
						passwdStr, version.RunAs, version.ID, version.Name)
					return nil
				}

				secretName := strings.TrimSpace(passwdStrSplit[0])
				secretKey := strings.TrimSpace(passwdStrSplit[1])
				secretNs := version.RunAs
				blog.Infof("to get passwd from secret(%s.%s::%s)", secretNs, secretName, secretKey)
				bcsSecret, err := store.FetchSecret(secretNs, secretName)
				if err != nil {
					blog.Error("get bcssecret(%s.%s) err: %s", secretNs, secretName, err.Error())
					return nil
				}
				if bcsSecret == nil {
					blog.Error("bcssecret(%s.%s) not exist", secretNs, secretName)
					return nil
				}
				bcsSecretItem, ok := bcsSecret.Data[secretKey]
				if ok == false {
					blog.Error("bcssecret item(key:%s) not exist in bcssecret(%s.%s)",
						secretKey, secretNs, secretName)
					return nil
				}

				passwdBase := strings.TrimSpace(bcsSecretItem.Content)
				if passwdBase != "" {
					passwdScrt, err := base64.StdEncoding.DecodeString(passwdBase)
					if err != nil {
						blog.Error("Decode base64(%s) err: %s", passwdBase, err.Error())
						return nil
					}
					passwd = string(passwdScrt)
				} else {
					blog.Infof("passwd in secret(%s.%s::%s) is empty", secretNs, secretName, secretKey)
					passwd = ""
				}
				//raw passwd
			} else {
				passwd = passwdConf
			}
			//default passwd
		} else {
			passwd = Passwd
		}

		var networkType string
		if version.Container[0].Docker.NetworkType != "" {
			networkType = fmt.Sprintf(" --network-mode %s", version.Container[0].Docker.NetworkType)
		}

		var cniDir string
		if BcsCniDir != "" {
			cniDir = fmt.Sprintf(" --cni-plugin %s", BcsCniDir)
		}

		var netImage string
		if NetImage != "" {
			netImage = fmt.Sprintf(" --network-image %s", NetImage)
		}

		uuid, _ := encrypt.DesEncryptToBase([]byte(passwd))
		execCommand = fmt.Sprintf("./%s --user %s --uuid %s %s %s %s", base, user, string(uuid), networkType, cniDir, netImage)
	}

	//blog.Infof("exec:%s", execCommand)

	var resources []*mesos.Resource
	resources = append(resources, createScalarResource("cpus", float64(types.CPUS_PER_EXECUTOR)))
	resources = append(resources, createScalarResource("mem", float64(types.MEM_PER_EXECUTOR)))
	resources = append(resources, createScalarResource("disk", float64(types.DISK_PER_EXECUTOR)))

	return &mesos.ExecutorInfo{
		Type:        mesos.ExecutorInfo_CUSTOM.Enum(),
		FrameworkId: offer.FrameworkId,
		ExecutorId:  &mesos.ExecutorID{Value: taskGroupID},
		Command: &mesos.CommandInfo{
			Uris: []*mesos.CommandInfo_URI{
				{
					Value:      proto.String(cmdOrUri),
					Executable: proto.Bool(true),
				},
			},
			Value: proto.String(execCommand),
		},
		Name:      proto.String("BcsExec"),
		Source:    proto.String("bcs"),
		Resources: resources,
	}
}
