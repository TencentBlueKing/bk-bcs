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

package util

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

func CheckKind(kind types.BcsDataType, by []byte) error {
	var meta *types.TypeMeta

	err := json.Unmarshal(by, &meta)
	if err != nil {
		return fmt.Errorf("Unmarshal TypeMeta failed: %s", err.Error())
	}

	if meta.Kind != kind {
		return fmt.Errorf("Kind %s is invalid", meta.Kind)
	}

	return nil
	/*switch kind {
	case types.BcsDataType_APP:
		var obj types.ReplicaController

		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_APP {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	case types.BcsDataType_PROCESS:
		var obj types.ReplicaController
		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_PROCESS {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	case types.BcsDataType_DEPLOYMENT:
		var obj types.BcsDeployment

		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_DEPLOYMENT {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	case types.BcsDataType_SERVICE:
		var obj types.BcsService

		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_SERVICE {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	case types.BcsDataType_CONFIGMAP:
		var obj types.BcsConfigMap

		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_CONFIGMAP {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	case types.BcsDataType_SECRET:
		var obj types.BcsSecret

		err = json.Unmarshal(by, &obj)
		if err != nil {
			break
		}

		if obj.Kind != types.BcsDataType_SECRET {
			err = fmt.Errorf("Kind %s is invalid", obj.Kind)
		}

	default:
		return fmt.Errorf("Kind %s is invalid", kind)
	}*/

	//return err
}
