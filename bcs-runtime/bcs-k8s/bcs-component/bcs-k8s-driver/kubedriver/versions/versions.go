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

package versions

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

//ClientSetter client set for multiple version
type ClientSetter struct {
	ClientSet   string
	BodyContent *[]byte
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//IfWithClientSet check api prefix
func (cs *ClientSetter) IfWithClientSet(uri string) bool {
	if strings.HasPrefix(uri, "api") || strings.HasPrefix(uri, "apis") {
		return true
	}
	return false
}

//GetClientSetUrl get uri relative version + group
func (cs *ClientSetter) GetClientSetUrl(uri string, version string, apiPrefer map[string]string) error {
	formattedUri := FormatURI(uri)
	allUrl := apiVersionMap[version]

	for _, url := range allUrl {
		formerUrl, laterUrl, err := SplitUrlByVersion(url)

		if err != nil {
			continue
		}
		if laterUrl == formattedUri {
			formerUrlParts := strings.Split(formerUrl, "/")

			if formerUrlParts[1] == "api" {
				cs.ClientSet = formerUrl
				return nil
			}

			group := formerUrlParts[len(formerUrlParts)-3]
			preferVersion := apiPrefer[group]
			formerUrlParts[len(formerUrlParts)-2] = preferVersion
			cs.ClientSet = strings.Join(formerUrlParts, "/")
			return nil
		}
	}

	return errors.New("url not found, please check k8s api docs")
}

//AddVersionIntoBody AddVersionIntoBody
func (cs *ClientSetter) AddVersionIntoBody() error {
	bodyJson := json.Get(*cs.BodyContent).GetInterface().(map[string]interface{})
	bodyJson["apiVersion"] = cs.ClientSet
	bodyByte, err := json.Marshal(bodyJson)
	if err != nil {
		return err
	}
	cs.BodyContent = &bodyByte
	return nil
}

//FormatURI FormatURI
func FormatURI(uri string) string {
	uriSplitParts := strings.Split(uri, "/")
	for index, uriSplitPart := range uriSplitParts {
		if index == 0 && len(uriSplitParts) >= 2 {
			if uriSplitPart == "namespaces" {
				uriSplitParts[1] = "{namespace}"
			} else {
				uriSplitParts[1] = "{name}"
			}
		}
		if index != 1 && index%2 != 0 {
			if uriSplitParts[index-1] == "proxy" {
				uriSplitParts[index] = "{path}"
			} else {
				uriSplitParts[index] = "{name}"
			}
		}
	}
	return strings.Join(uriSplitParts, "/")
}

//SplitUrlByVersion SplitUrlByVersion
func SplitUrlByVersion(url string) (string, string, error) {
	r, err := regexp.Compile("/v[0-9][a-z0-9]*/")
	if err != nil {
		return "", "", err
	}
	indexs := r.FindStringIndex(url)
	if len(indexs) < 1 {
		return "", "", errors.New("doesn't exist version part")
	}
	lastPosition := indexs[1]
	return url[0:lastPosition], url[lastPosition:], nil
}

//FetchAllUrl fetch all k8s api list from
func FetchAllUrl(version string) ([]string, error) {
	jsonFiles, err := FetchAllJsonFiles("metadata/*.json")

	if err != nil {
		return nil, err
	}
	for _, jsonFile := range jsonFiles {
		if jsonFile != "metadata/"+version+".json" {
			continue
		}
		urlsByVersion, err := ReadJsonFile(jsonFile, "paths")
		if err != nil {
			return nil, err
		}
		return urlsByVersion.Keys(), nil
	}
	return nil, fmt.Errorf("%s version is not supported yet", version)
}

//FetchAllJsonFiles FetchAllJsonFiles
func FetchAllJsonFiles(jsonPath string) ([]string, error) {
	files, err := filepath.Glob(jsonPath)
	if err != nil {
		return nil, err
	}
	return files, nil
}

//ReadJsonFile ReadJsonFile
func ReadJsonFile(jsonFile string, path string) (jsoniter.Any, error) {
	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	// convert raw json content into []apiInfo
	apiList := json.Get(bytes, path)
	return apiList, nil
}
