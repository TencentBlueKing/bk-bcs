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
 */

package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

	clusterclient "github.com/argoproj/argo-cd/v2/pkg/apiclient/cluster"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4/clustermanager"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// ClusterControl interface definition
type ClusterControl interface {
	Controller
	SingleStart(ctx context.Context)
	ForceSync(projectCode, clusterID string)
	SyncProject(ctx context.Context, projectCode string) error
}

// NewClusterController create project controller instance
func NewClusterController(opt *Options) ClusterControl {
	return &cluster{
		option: opt,
	}
}

// ClusterController for bk-bcs cluster information
// syncing to gitops system. depend on cluster-manager interface
type cluster struct {
	sync.Mutex

	option *Options
	client cm.ClusterManagerClient
	conn   *grpc.ClientConn
}

// Init controller
func (control *cluster) Init() error {
	if control.option == nil {
		return fmt.Errorf("cluster controller lost options")
	}
	//if control.option.Mode == common.ModeService {
	//	return fmt.Errorf("service mode is not implenmented")
	//}
	// init with raw grpc connection
	if err := control.initClient(); err != nil {
		return err
	}
	return nil
}

// Start controller
func (control *cluster) Start() error {
	return nil
}

// Stop controller
func (control *cluster) Stop() {
	control.conn.Close() // nolint
}

// SingleStart will start inner loop for cluster sync, and it
// will work until context cancel.
func (control *cluster) SingleStart(ctx context.Context) {
	blog.Infof("cluster controller single start....")
	if err := control.innerLoop(ctx); err != nil {
		blog.Errorf("inner loop first failed: %s", err.Error())
	}
	tick := time.NewTicker(time.Second * time.Duration(control.option.Interval))
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			blog.Infof("cluster controller ask to stop in SingleStart")
			return
		case <-tick.C:
			if err := control.innerLoop(ctx); err != nil {
				blog.Errorf("inner loop failed: %s", err.Error())
			}
		}
	}
}

// ForceSync specified cluster information
func (control *cluster) ForceSync(projectCode, clusterID string) {
	control.Lock()
	defer control.Unlock()

	// reading data from cluster-manager
	header := metadata.New(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", control.option.APIToken)})
	outCxt := metadata.NewOutgoingContext(context.Background(), header)
	response, err := control.client.GetCluster(outCxt, &cm.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		blog.Errorf("cluster controller get cluster %s from cluster-manager failure, %s",
			clusterID, err.Error())
		return
	}
	if response.Code != 0 {
		blog.Errorf("cluster-manager response for %s logic err, %s", clusterID, response.Message)
		return
	}
	if response.Data == nil {
		blog.Warnf("cluster-manager found no cluster %s", clusterID)
		return
	}
	if response.Data.IsShared {
		return
	}

	cls := response.Data
	argoCluster, err := control.option.Storage.GetCluster(context.Background(), &clusterclient.ClusterQuery{
		Name: cls.ClusterID,
	})
	if err != nil {
		blog.Errorf("query cluster '%s' from storage failed: %s", cls.ClusterID, err.Error())
		return
	}
	if argoCluster != nil {
		if err = control.updateToStorage(context.Background(), cls, argoCluster); err != nil {
			blog.Errorf("update cluster '%s' to storage failed: %s", cls.ClusterID, err.Error())
		}
		return
	}

	appPro, err := control.option.Storage.GetProject(context.Background(), projectCode)
	if err != nil {
		blog.Errorf("cluster controller get project %s for cluster %s from storage failure, %s",
			projectCode, clusterID, err.Error())
		return
	}
	// write data down to gitops system
	if err := control.saveToStorage(context.Background(), response.Data, appPro); err != nil {
		blog.Errorf("cluster controller save cluster %s to storage failure, %s", clusterID, err.Error())
		return
	}
}

func (control *cluster) initClient() error {
	//create grpc connection
	header := map[string]string{
		"x-content-type": "application/grpc+proto",
		"Content-Type":   "application/grpc",
	}
	if len(control.option.APIToken) != 0 {
		header["Authorization"] = fmt.Sprintf("Bearer %s", control.option.APIToken)
	}
	md := metadata.New(header)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.Header(&md)))
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(control.option.ClientTLS)))
	conn, err := grpc.Dial(control.option.APIGateway, opts...)
	if err != nil {
		blog.Errorf("cluster controller dial bcs-api-gateway %s failure, %s",
			control.option.APIGateway, err.Error())
		return err
	}
	control.client = cm.NewClusterManagerClient(conn)
	control.conn = conn
	blog.Infof("cluster controller init cluster-manager with %s successfully", control.option.APIGateway)
	return nil
}

func (control *cluster) innerLoop(ctx context.Context) error {
	// list all project in local storage
	appProjects, err := control.option.Storage.ListProjects(ctx)
	if err != nil {
		return errors.Wrapf(err, "innerLoop get all projects fro gitops storage failed")
	}
	controlledProjects := make(map[string]*v1alpha1.AppProject)
	for i, pro := range appProjects.Items {
		proID := common.GetBCSProjectID(pro.Annotations)
		if proID == "" {
			continue
		}
		controlledProjects[proID] = &appProjects.Items[i]
	}
	blog.Infof("cluster controller list raw projects %d, controlled projects %d",
		len(appProjects.Items), len(controlledProjects))
	// list all cluster for every project
	for proID, appPro := range controlledProjects {
		blog.Infof("syncing clusters for project [%s]%s", appPro.Name, proID)
		if err := control.syncClustersByProject(ctx, proID, appPro); err != nil {
			blog.Errorf("sync clusters for project [%s]%s failed: %s", appPro.Name, proID, err.Error())
			//continue
		}
		blog.Infof("sync clusters for project [%s]%s complete, next...", appPro.Name, proID)

		// sync secret init
		if err = control.option.Secret.InitProjectSecret(ctx, appPro.Name); err != nil {
			if utils.IsSecretAlreadyExist(err) {
				continue
			}
			blog.Errorf("sync secrets for project [%s]%s failed: %s", appPro.Name, proID, err.Error())
			continue
		}
		blog.Infof("init project [%s]%s secrets complete", appPro.Name, proID)

		// sync secret info to pro annotations
		//secretVal := vaultcommon.GetVaultSecForProAnno(appPro.Name)
		secretVal, err := control.option.Secret.GetProjectSecret(ctx, appPro.Name)
		if err != nil {
			blog.Errorf("[getErr]sync secret info to pro annotations [%s]%s failed: %s", appPro.Name, proID, err.Error())
			continue
		}
		actualVal, ok := appPro.Annotations[common.SecretKey]
		if !ok {
			appPro.Annotations[common.SecretKey] = secretVal
			if err := control.option.Storage.UpdateProject(ctx, appPro); err != nil {
				blog.Errorf("[existErr]sync secret info to pro annotations [%s]%s failed: %s", appPro.Name, proID, err.Error())
			}
		} else {
			if secretVal != actualVal {
				appPro.Annotations[common.SecretKey] = secretVal
				if err := control.option.Storage.UpdateProject(ctx, appPro); err != nil {
					blog.Errorf("[valErr]sync secret info to pro annotations [%s]%s failed: %s", appPro.Name, proID, err.Error())
				}
			}
		}
		blog.Infof("sync secret info to pro annotations[val:%s] [%s]%s complete. next...", secretVal, appPro.Name, proID)
	}
	return nil
}

// SyncProject sync all clusters by project code
func (control *cluster) SyncProject(ctx context.Context, projectCode string) error {
	argoProject, err := control.option.Storage.GetProject(ctx, projectCode)
	if err != nil {
		return errors.Wrapf(err, "get project '%s' failed", projectCode)
	}
	if argoProject == nil {
		return errors.Errorf("project '%s' not exist", projectCode)
	}
	proID := common.GetBCSProjectID(argoProject.Annotations)
	if proID == "" {
		return errors.Errorf("project '%s' is not under control", projectCode)
	}
	return control.syncClustersByProject(ctx, proID, argoProject)
}

func (control *cluster) syncClustersByProject(ctx context.Context, projectID string,
	appPro *v1alpha1.AppProject) error {
	control.Lock()
	defer control.Unlock()

	clusterMap, err := control.buildClustersByProject(ctx, projectID)
	if err != nil {
		return errors.Wrapf(err, "list clusters from project managerfor project [%s]%s failed",
			appPro.Name, projectID)
	}
	argoClusterMap, err := control.buildArgoClusters(ctx, projectID)
	if err != nil {
		return errors.Wrapf(err, "list clusters from argo stage for project '%s' failed", projectID)
	}

	needUpdate, needCreate, needDelete := control.compareClusters(clusterMap, argoClusterMap)
	for _, clsID := range needUpdate {
		cls := clusterMap[clsID]
		argoCls := argoClusterMap[clsID]
		if err = control.updateToStorage(ctx, cls, argoCls); err != nil {
			blog.Errorf("cluster '%s' update to argo storage failed: %s", clsID, err.Error())
			continue
		}
		blog.Infof("update cluster '%s' to argo storage success", clsID)
	}
	for _, clsID := range needCreate {
		cls := clusterMap[clsID]
		if err = control.saveToStorage(ctx, cls, appPro); err != nil {
			blog.Errorf("cluster '%s' save to argo storage failed: %s", clsID, err.Error())
			continue
		}
		blog.Infof("save cluster '%s' to argo storage success", clsID)
	}
	for _, clsID := range needDelete {
		if err = control.option.Storage.DeleteCluster(ctx, clsID); err != nil {
			blog.Errorf("delete cluster '%s' from argo storage failed: %s", clsID, err.Error())
			continue
		}
		blog.Infof("delete cluster '%s' argo success", clsID)
	}
	return nil
}

func (control *cluster) buildArgoClusters(ctx context.Context,
	projectID string) (map[string]*v1alpha1.Cluster, error) {
	argoClusters, err := control.option.Storage.ListClustersByProject(ctx, projectID)
	if err != nil {
		return nil, errors.Wrapf(err, "list clusters by project '%s' failed", projectID)
	}
	argoClusterMap := make(map[string]*v1alpha1.Cluster)
	for i := range argoClusters.Items {
		item := argoClusters.Items[i]
		argoClusterMap[item.Name] = &item
	}
	return argoClusterMap, nil
}

func (control *cluster) buildClustersByProject(ctx context.Context,
	projectID string) (map[string]*clustermanager.Cluster, error) {
	bcsCtx := metadata.NewOutgoingContext(ctx,
		metadata.New(map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", control.option.APIToken),
		}),
	)
	clusters, err := control.client.ListCluster(bcsCtx, &cm.ListClusterReq{ProjectID: projectID})
	if err != nil {
		return nil, errors.Wrapf(err, "list clusters failed")
	}
	if clusters.Code != 0 {
		return nil, errors.Errorf("list clusters resp code not 0 build %d: %s",
			clusters.Code, clusters.Message)
	}
	clusterMap := make(map[string]*clustermanager.Cluster)
	for _, item := range clusters.Data {
		if !item.IsShared {
			clusterMap[item.ClusterID] = item
		}
	}
	return clusterMap, nil
}

func (control *cluster) compareClusters(clusterMap map[string]*clustermanager.Cluster,
	argoClusterMap map[string]*v1alpha1.Cluster) ([]string, []string, []string) {
	needUpdate := make([]string, 0)
	needCreate := make([]string, 0)
	needDelete := make([]string, 0)
	for clsID := range clusterMap {
		_, ok := argoClusterMap[clsID]
		if ok {
			needUpdate = append(needUpdate, clsID)
		} else {
			needCreate = append(needCreate, clsID)
		}
	}
	for clsID := range argoClusterMap {
		if _, ok := clusterMap[clsID]; !ok {
			needDelete = append(needDelete, clsID)
		}
	}
	return needUpdate, needCreate, needDelete
}

// updateToStorage check the cluster whether exist. If existed, we need
// check the cluster's name whether changed, and update the cluster object
// from argocd.
func (control *cluster) updateToStorage(ctx context.Context, cls *cm.Cluster,
	argoCluster *v1alpha1.Cluster) error {
	// we should check the cluster's attr whether there has been a change, if changed
	// need to update to argocd storage
	needUpdate := false
	if argoCluster.Annotations[common.ClusterAliaName] != cls.ClusterName {
		needUpdate = true
		argoCluster.Annotations[common.ClusterAliaName] = cls.ClusterName
	}
	if argoCluster.Annotations[common.ClusterEnv] != cls.Environment {
		needUpdate = true
		argoCluster.Annotations[common.ClusterEnv] = cls.Environment
	}
	if !needUpdate {
		return nil
	}
	// the cluster will be updated if cluster alias name is changed
	if err := control.option.Storage.UpdateCluster(ctx, argoCluster); err != nil {
		return errors.Wrapf(err, "update cluster '%s' from gitops storage failed", cls.ClusterID)
	}
	return nil
}

func (control *cluster) saveToStorage(ctx context.Context, cls *cm.Cluster, project *v1alpha1.AppProject) error {
	if cls.IsShared {
		return fmt.Errorf("shared cluster is not supported")
	}
	clusterAnnotation := utils.DeepCopyMap(project.Annotations)
	clusterAnnotation[common.ClusterAliaName] = cls.ClusterName
	clusterAnnotation[common.ClusterEnv] = cls.Environment
	appCluster := &v1alpha1.Cluster{
		Name:    cls.ClusterID,
		Server:  fmt.Sprintf("https://%s/clusters/%s/", control.option.APIGatewayForCluster, cls.ClusterID),
		Project: project.Name,
		Config: v1alpha1.ClusterConfig{
			BearerToken: control.option.APIToken,
		},
		Annotations: clusterAnnotation,
	}
	return control.option.Storage.CreateCluster(ctx, appCluster)
}
