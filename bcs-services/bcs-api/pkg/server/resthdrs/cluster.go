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

package resthdrs

import (
	"fmt"
	"reflect"
	"strings"

	"bk-bcs/bcs-common/common"
	"bk-bcs/bcs-common/common/blog"
	m "bk-bcs/bcs-services/bcs-api/pkg/models"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/utils"
	"bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/dchest/uniuri"
	"github.com/emicklei/go-restful"
	"github.com/iancoleman/strcase"
)

func initCluster(cluster *m.Cluster) error {
	return nil
}

func createCluster(cluster *m.Cluster) (*m.Cluster, error) {
	clusterId := cluster.ID
	if sqlstore.GetCluster(clusterId) != nil {
		return nil, utils.NewClusterAreadyExistError("create failed, cluster with this id already exists")
	}

	err := sqlstore.CreateCluster(cluster)
	if err != nil {
		return nil, utils.NewCannotCreateClusterError(fmt.Sprintf("can not create cluster, error: %s", err))
	}

	err = initCluster(cluster)
	if err != nil {
		return nil, utils.NewClusterInitFailedError(fmt.Sprintf("cluster init failed, error: %s", err))
	}
	return sqlstore.GetCluster(clusterId), nil
}

func createClusterWithExternalInfo(cluster *m.Cluster, externalClusterInfo *m.BCSClusterInfo, response *restful.Response) {
	cluster, err := createCluster(cluster)
	// convert type name to screaming snake
	errorCode := strcase.ToScreamingSnake(fmt.Sprint(reflect.TypeOf(cluster)))
	if err != nil {
		message := fmt.Sprintf("errcode: %d, can not create cluster, error: %s", common.BcsErrApiInternalDbError, err)
		WriteClientError(response, errorCode, message)
		return
	}

	err = sqlstore.CreateBCSClusterInfo(externalClusterInfo)
	if err != nil {
		message := fmt.Sprintf("errcode: %d, can not create external cluster info, error: %s", common.BcsErrApiInternalDbError, err)
		WriteServerError(response, "CANNOT_CREATE_EXTERNAL_CLUSTER_INFO", message)
		return
	}
	response.WriteEntity(*sqlstore.GetCluster(cluster.ID))

}

// PlainCluster

type CreatePlainClusterForm struct {
	ID string `json:"id" validate:"required"`
}

// CreatePlainCluster creates a "plain" cluster for current user
func CreatePlainCluster(request *restful.Request, response *restful.Response) {
	form := CreatePlainClusterForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		response.WriteEntity(FormatValidationError(err))
		return
	}

	// Prepend a fixed prefix to avoid id conflict across providers
	clusterId := m.ClusterIdPrefixPlain + form.ID
	user := filters.GetUser(request)
	cluster := &m.Cluster{
		ID:          clusterId,
		Provider:    m.ClusterProviderPlain,
		CreatorId:   user.ID,
		TurnOnAdmin: false,
	}
	cluster, err = createCluster(cluster)
	// convert type name to screaming snake
	errorCode := strcase.ToScreamingSnake(fmt.Sprint(reflect.TypeOf(cluster)))
	if err != nil {
		WriteClientError(response, errorCode, fmt.Sprintf("can not create cluster, error: %s", err))
		return
	}
	// init plain cluster permissions
	for _, name := range []string{m.ClusterPermNameView, m.ClusterPermNameEdit} {
		err := sqlstore.SaveUserClusterPerm(m.PermBackendTypeSyncOnce, user, cluster, name, true)
		if err != nil {
			blog.Errorf("error save userCluster permission: %s", err.Error())
		}
	}
	response.WriteEntity(*cluster)
}

// BCSCluster

// CreateBCSClusterForm
type CreateBCSClusterForm struct {
	ID        string `json:"id" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}

// CreateBCSCluster creates a "BCS" cluster for current user
func CreateBCSCluster(request *restful.Request, response *restful.Response) {
	blog.Debug(fmt.Sprintf("CreateBCSCluster begin"))
	form := CreateBCSClusterForm{}
	request.ReadEntity(&form)

	err := validate.Struct(&form)
	if err != nil {
		blog.Debug(fmt.Sprintf("CreateBCSCluster form validate failed, %s", err))
		response.WriteEntity(FormatValidationError(err))
		return
	}

	// check the permission
	user := filters.GetUser(request)

	// check if cluster exists already
	externalClusterInfo := sqlstore.QueryBCSClusterInfo(&m.BCSClusterInfo{
		SourceProjectId: form.ProjectID,
		SourceClusterId: form.ID,
	})
	if externalClusterInfo != nil {
		message := fmt.Sprintf("errcode: %d, create failed, cluster with this id already exists", common.BcsErrApiBadRequest)
		WriteClientError(response, "CLUSTER_ALREADY_EXISTS", message)
		return
	}

	// Use the "{external_id}-{random-identifier}" as the real cluster id to ensure both uniqueness and readability
	// "BCS-K8S-15007" -> "bcs-bcs-k8s-15007-FvBewMk3"
	clusterId := fmt.Sprintf("%s%s-%s", m.ClusterIdPrefixBCS, strings.ToLower(form.ID), uniuri.NewLen(8))
	cluster := &m.Cluster{
		ID:          clusterId,
		Provider:    m.ClusterProviderBCS,
		CreatorId:   user.ID,
		TurnOnAdmin: false,
	}
	externalClusterInfo = &m.BCSClusterInfo{
		ClusterId:       clusterId,
		SourceProjectId: form.ProjectID,
		SourceClusterId: form.ID,
	}
	createClusterWithExternalInfo(cluster, externalClusterInfo, response)
}
