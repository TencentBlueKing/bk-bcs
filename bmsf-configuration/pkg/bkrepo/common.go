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

package bkrepo

import (
	"errors"
	"fmt"
	"strings"
)

// GenContentURL returns config content URL of target source.
func GenContentURL(host, project, contentID string) (string, error) {
	if len(host) == 0 {
		return "", errors.New("empty host")
	}
	if len(project) == 0 {
		return "", errors.New("empty project")
	}
	if len(contentID) == 0 {
		return "", errors.New("empty contentID")
	}
	return fmt.Sprintf("%s/%s/%s%s/%s/%s",
		host, GENERICAPIPATH, BSCPBIZIDPREFIX, project, CONFIGSREPONAME, strings.ToUpper(contentID)), nil
}
