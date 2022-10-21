<<<<<<< HEAD
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
=======
/**
 * @Author: Ambition
 * @Description:
 * @File: client_test
 * @Version: 1.0.0
 * @Date: 2022/10/17 11:48
>>>>>>> cd831f67dcced2448d87af2258a3604299f448fc
 */

package pkg

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func Test_GeProjectList(t *testing.T) {
	client, _, err := NewBcsProjectCli(context.Background(), &Config{
		APIServer: "127.0.0.1:8091",
		AuthToken: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	rsp, err := client.ListProjects(context.Background(), &bcsproject.ListProjectsRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
