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
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
)

// ClusterControl interface definition
type ClusterControl interface {
	Controller
	// only work until context cancel
	SingleStart(ctx context.Context)
	ForceSync(projectCode, clusterID string)
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
	syncing bool
	option  *Options
	client  cm.ClusterManagerClient
	conn    *grpc.ClientConn
}

// Init controller
func (control *cluster) Init() error {
	if control.option == nil {
		return fmt.Errorf("cluster controller lost options")
	}
	if control.option.Mode == common.ModeService {
		return fmt.Errorf("service mode is not implenmented")
	}
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

// Start controller
func (control *cluster) SingleStart(ctx context.Context) {
	blog.Infof("cluster controller single start....")
	tick := time.NewTicker(time.Second * time.Duration(control.option.Interval))
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			blog.Infof("cluster controller ask to stop in SingleStart")
			return
		case <-tick.C:
			control.innerLoop(ctx)
		}
	}
}

// ForceSync specified cluster information
func (control *cluster) ForceSync(projectCode, clusterID string) {
	if control.syncing {
		blog.Infof("cluster controller is under data synchronization")
		return
	}
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
	exist, err := control.isClusterExist(response.Data)
	if err != nil {
		blog.Errorf("cluster controller confirm cluster %s exsitencen failure, %s", clusterID, err.Error())
		return
	}
	if exist {
		blog.Infof("cluster controller found cluster %s already exist, skip", clusterID)
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

func (control *cluster) innerLoop(ctx context.Context) {
	control.syncing = true
	defer func() {
		control.syncing = false
	}()
	// list all project in local storage
	appProjects, err := control.option.Storage.ListProjects(ctx)
	if err != nil {
		blog.Errorf("cluster controller get all projects from gitops storage failure, %s", err.Error())
		return
	}
	controlledProjects := make(map[string]*v1alpha1.AppProject)
	for _, pro := range appProjects.Items {
		proID := common.GetBCSProjectID(pro.Annotations)
		if proID == "" {
			continue
		}
		controlledProjects[proID] = &pro
	}
	blog.Infof("cluster controller list raw projects %d, controlled projects %d",
		len(appProjects.Items), len(controlledProjects))
	// list all cluster for every project
	for proID, appPro := range controlledProjects {
		control.innerClusterLoop(ctx, proID, appPro)
	}
}

func (control *cluster) innerClusterLoop(ctx context.Context, projectID string, appPro *v1alpha1.AppProject) {
	header := metadata.New(map[string]string{"Authorization": fmt.Sprintf("Bearer %s", control.option.APIToken)})
	outCxt := metadata.NewOutgoingContext(ctx, header)
	clusters, err := control.client.ListCluster(outCxt, &cm.ListClusterReq{ProjectID: projectID})
	if err != nil {
		blog.Errorf("cluster controller list all clusters for project [%s]%s failure, %s. recovery from next tick",
			appPro.Name, projectID, err.Error())
		return
	}
	if clusters.Code != 0 {
		blog.Errorf("cluster controller list cluster for project [%s]%s failure, cluster-manager logic err: %s",
			appPro.Name, projectID, err.Error())
		return
	}
	if len(clusters.Data) == 0 {
		blog.Warnf("cluster controller get 0 cluster for project [%s]%s", appPro.Name, projectID)
		return
	}

	for _, cls := range clusters.Data {
		exist, err := control.isClusterExist(cls)
		if err != nil {
			blog.Errorf("cluster controller confirm cluster %s existence failure, %s. wait for next tick",
				cls.ClusterID, err.Error())
			continue
		}
		if exist {
			blog.Infof("cluster %s exist in gitops storage, skip", cls.ClusterID)
			continue
		}
		if err := control.saveToStorage(ctx, cls, appPro); err != nil {
			blog.Errorf("cluster controller save cluster %s to storage for project [%s]%s failure, %s",
				cls.ClusterID, appPro.Name, projectID, err.Error())
			continue
		}
		blog.Infof("cluster controller add new cluster %s for project [%s]%s",
			cls.ClusterID, appPro.Name, projectID)
	}
}

// isClusterExist check cluster information already in storage
func (control *cluster) isClusterExist(cls *cm.Cluster) (bool, error) {
	gitopsCluster, err := control.option.Storage.GetCluster(context.Background(), cls.ClusterID)
	if err != nil {
		blog.Errorf("query cluster %s from storage failure, %s", cls.ClusterID, err.Error())
		return false, err
	}
	if gitopsCluster == nil {
		blog.Warnf("no cluster %s in storage", cls.ClusterID)
		return false, nil
	}
	return true, nil
}

func (control *cluster) saveToStorage(ctx context.Context, cls *cm.Cluster, project *v1alpha1.AppProject) error {
	if cls.IsShared {
		return fmt.Errorf("shared cluster is not supported")
	}
	appCluster := &v1alpha1.Cluster{
		Name:    cls.ClusterID,
		Server:  fmt.Sprintf("https://%s/clusters/%s/", control.option.APIGateway, cls.ClusterID),
		Project: project.Name,
		Config: v1alpha1.ClusterConfig{
			BearerToken: control.option.APIToken,
		},
		Annotations: project.Annotations,
	}
	return control.option.Storage.CreateCluster(ctx, appCluster)
}
