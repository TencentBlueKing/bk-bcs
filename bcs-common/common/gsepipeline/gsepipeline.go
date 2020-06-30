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

package gsepipeline

import (
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	gseclient "github.com/Tencent/bk-bcs/bcs-common/common/gsepipeline/client"
)

type gseStorage struct {
	client gseclient.AsyncProducer
}

//NewGseStorage create a new gseclient
func NewGseStorage(endpoint string) (Storage, error) {

	blog.Info("endpoint:%v", endpoint)

	client, err := gseclient.New(endpoint)
	if nil != err {
		return nil, err
	}

	if nil == err {
		gseStorageClient := &gseStorage{client: client}

		return gseStorageClient, nil
	}

	return nil, err

}

// AddStats pushing Stats info
func (gse *gseStorage) AddStats(dmsg LogMsg) error {

	b, err := json.Marshal(dmsg)
	if err != nil {
		return err
	}

	gse.client.Input(&gseclient.ProducerMessage{DataID: uint32(dmsg.DataID), Value: b})

	return nil
}

// Close close pipeline
func (gse *gseStorage) Close() error {

	gse.client.Close()
	return nil
}
