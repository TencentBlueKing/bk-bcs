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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	proto "github.com/golang/protobuf/proto"
)

//BcsContainerExecutorPath default container binary path
var BcsContainerExecutorPath string

//BcsProcessExecutorPath default process binary path
var BcsProcessExecutorPath string

// BcsCniDir default bcs cni plugin directory
var BcsCniDir string

// NetImage default network image information
var NetImage string

// Passwd registry pass
var Passwd = static.BcsDefaultPasswd

// User registry user
var User = static.BcsDefaultUser

//InitExecutorInfo init mesos executor info
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
func CreateBcsExecutorInfo(offer *mesos.Offer, taskGroupID *string,
	version *types.Version, store store.Store) (*mesos.ExecutorInfo, error) {

	var cmdOrURI string
	switch version.Kind {
	case commtypes.BcsDataType_PROCESS:
		cmdOrURI = BcsProcessExecutorPath
	case commtypes.BcsDataType_APP, "", commtypes.BcsDataType_Daemonset:
		cmdOrURI = BcsContainerExecutorPath
	}

	pathSplit := strings.Split(cmdOrURI, "/")

	var base string

	if len(pathSplit) > 0 {
		base = pathSplit[len(pathSplit)-1]
	} else {
		base = cmdOrURI
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
					return nil, fmt.Errorf("image secret of user formation err")
				}

				userStr := userConfSplit[1]
				userStrSplit := strings.Split(userStr, "||")
				if len(userStrSplit) != 2 {
					blog.Error("ImagePullUser sercret config format(%s) error, version(%s.%s.%s)",
						userStr, version.RunAs, version.ID, version.Name)
					return nil, fmt.Errorf("image secret user information err")
				}

				secretName := strings.TrimSpace(userStrSplit[0])
				secretKey := strings.TrimSpace(userStrSplit[1])
				secretNs := version.RunAs
				blog.Infof("to get user from secret(%s.%s::%s)", secretNs, secretName, secretKey)
				bcsSecret, err := store.FetchSecret(secretNs, secretName)
				if err != nil {
					blog.Error("get bcssecret(%s.%s) err: %s", secretNs, secretName, err.Error())
					return nil, fmt.Errorf("user secret error: %s", err.Error())
				}
				if bcsSecret == nil {
					blog.Error("bcssecret(%s.%s) not exist", secretNs, secretName)
					return nil, fmt.Errorf("secret %s do not exist", secretName)
				}
				bcsSecretItem, ok := bcsSecret.Data[secretKey]
				if ok == false {
					blog.Error("bcssecret item(key:%s) not exist in bcssecret(%s.%s)",
						secretKey, secretNs, secretName)
					return nil, fmt.Errorf("secret key %s do not exist", secretKey)
				}

				userBase := strings.TrimSpace(bcsSecretItem.Content)
				if userBase != "" {
					userScrt, err := base64.StdEncoding.DecodeString(userBase)
					if err != nil {
						blog.Error("Decode base64(%s) err: %s", userBase, err.Error())
						return nil, fmt.Errorf("secret content decode err, %s", err.Error())
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
					return nil, fmt.Errorf("image secret of passwd formation err")
				}

				passwdStr := passwdConfSplit[1]
				passwdStrSplit := strings.Split(passwdStr, "||")
				if len(passwdStrSplit) != 2 {
					blog.Error("ImagePullPasswd sercret config format(%s) error, version(%s.%s.%s)",
						passwdStr, version.RunAs, version.ID, version.Name)
					return nil, fmt.Errorf("image passwd secret formation err")
				}

				secretName := strings.TrimSpace(passwdStrSplit[0])
				secretKey := strings.TrimSpace(passwdStrSplit[1])
				secretNs := version.RunAs
				blog.Infof("to get passwd from secret(%s.%s::%s)", secretNs, secretName, secretKey)
				bcsSecret, err := store.FetchSecret(secretNs, secretName)
				if err != nil {
					blog.Error("get bcssecret(%s.%s) err: %s", secretNs, secretName, err.Error())
					return nil, fmt.Errorf("passwd secret error: %s", err.Error())
				}
				if bcsSecret == nil {
					blog.Error("bcssecret(%s.%s) not exist", secretNs, secretName)
					return nil, fmt.Errorf("passwd secret do not exist")
				}
				bcsSecretItem, ok := bcsSecret.Data[secretKey]
				if ok == false {
					blog.Error("bcssecret item(key:%s) not exist in bcssecret(%s.%s)",
						secretKey, secretNs, secretName)
					return nil, fmt.Errorf("passwd secret key do not exist")
				}

				passwdBase := strings.TrimSpace(bcsSecretItem.Content)
				if passwdBase != "" {
					passwdScrt, err := base64.StdEncoding.DecodeString(passwdBase)
					if err != nil {
						blog.Error("Decode base64(%s) err: %s", passwdBase, err.Error())
						return nil, fmt.Errorf("passwd secret content decode err, %s", err.Error())
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
					Value:      proto.String(cmdOrURI),
					Executable: proto.Bool(true),
				},
			},
			Value: proto.String(execCommand),
		},
		Name:      proto.String("BcsExec"),
		Source:    proto.String("bcs"),
		Resources: resources,
	}, nil
}
