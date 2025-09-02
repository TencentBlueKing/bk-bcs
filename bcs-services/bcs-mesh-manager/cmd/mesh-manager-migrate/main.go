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

// Package main 提供Istio迁移工具
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager-migrate/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager-migrate/internal/migrator"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// 解析命令行参数（包括配置文件和迁移参数）
	configPath := flag.String("f", "./migrate-config.json", "migrate configuration json file")
	migrateOpts := parseMigrateFlags()
	flag.Parse()

	// 验证迁移参数
	if err := validateMigrateParams(migrateOpts); err != nil {
		log.Fatalf("failed to validate migrate parameters: %v", err)
	}

	// 加载配置文件
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config failed for file %s: %v", *configPath, err)
	}

	// 覆盖配置参数
	if migrateOpts.MongodbUsername != "" {
		cfg.Mongo.Username = migrateOpts.MongodbUsername
	}
	if migrateOpts.MongodbPassword != "" {
		cfg.Mongo.Password = migrateOpts.MongodbPassword
	}
	if migrateOpts.MongoAddress != "" {
		cfg.Mongo.Address = migrateOpts.MongoAddress
	}

	// 初始化迁移器
	m := migrator.New(cfg)
	if err := m.Init(); err != nil {
		log.Fatalf("init migrator failed: %v", err)
	}

	// 执行迁移
	if err := m.Migrate(migrateOpts); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("migration completed successfully")
}

// parseMigrateFlags 解析迁移参数
func parseMigrateFlags() *migrator.MigrateOptions {
	migrateOpts := &migrator.MigrateOptions{}
	// required params
	flag.StringVar(&migrateOpts.PrimaryClusterID, "primaryClusterID",
		"", "target cluster ID for migration (required)")
	flag.StringVar(&migrateOpts.MeshName, "meshName", "", "mesh name for migration (required)")
	flag.StringVar(&migrateOpts.ProjectCode, "projectCode", "", "project code (required)")
	flag.StringVar(&migrateOpts.IstiodReleaseName, "istiodReleaseName",
		"bcs-istio-istiod", "istiod release name(required)")
	flag.StringVar(&migrateOpts.BaseReleaseName, "baseReleaseName",
		"bcs-istio-base", "istio base release name(required)")
	flag.StringVar(&migrateOpts.MongodbUsername, "mongodbUsername", "", "mongodb username(required)")
	flag.StringVar(&migrateOpts.MongodbPassword, "mongodbPassword", "", "mongodb password(required)")
	flag.StringVar(&migrateOpts.MongoAddress, "mongoAddress", "", "mongodb address(required)")
	flag.StringVar(&migrateOpts.KubeconfigPath, "kubeconfigPath", "", "kubeconfig path(required)")
	flag.StringVar(&migrateOpts.BcsUsername, "bcsUsername", "", "bcs username(required)")

	// 多集群配置(optional)
	flag.BoolVar(&migrateOpts.MultiClusterEnabled, "multiClusterEnabled", false, "multi cluster enabled")
	flag.StringVar(&migrateOpts.GatewaysReleaseName, "gatewaysReleaseName",
		"bcs-istio-eastwestgateway", "istio gateways release name")
	flag.StringVar(&migrateOpts.RemoteClusters, "remoteClusters", "", "remote clusters, comma separated")

	flag.StringVar(&migrateOpts.Description, "description", "", "mesh description")
	flag.BoolVar(&migrateOpts.MetricsEnabled, "metricsEnabled", false, "enable metrics collection")
	flag.BoolVar(&migrateOpts.ControlPlaneMetricsEnabled, "controlPlaneMetricsEnabled",
		false, "enable control plane metrics")
	flag.BoolVar(&migrateOpts.DataPlaneMetricsEnabled, "dataPlaneMetricsEnabled",
		false, "enable data plane metrics")
	flag.StringVar(&migrateOpts.TracingEndpoint, "tracingEndpoint", "", "tracing endpoint")
	flag.StringVar(&migrateOpts.BkToken, "bkToken", "", "tracing bk token")

	return migrateOpts
}

func validateMigrateParams(opts *migrator.MigrateOptions) error {
	if opts.PrimaryClusterID == "" {
		return fmt.Errorf("primaryClusterID is required")
	}
	if opts.ProjectCode == "" {
		return fmt.Errorf("projectCode is required")
	}
	if opts.MeshName == "" {
		return fmt.Errorf("meshName is required")
	}
	if opts.IstiodReleaseName == "" {
		return fmt.Errorf("istiodReleaseName is required")
	}
	if opts.BaseReleaseName == "" {
		return fmt.Errorf("baseReleaseName is required")
	}
	if opts.MongoAddress == "" {
		return fmt.Errorf("mongoAddress is required")
	}
	if opts.KubeconfigPath == "" {
		return fmt.Errorf("kubeconfigPath is required")
	}
	if opts.MultiClusterEnabled {
		if opts.GatewaysReleaseName == "" {
			return fmt.Errorf("gatewaysReleaseName is required")
		}
	}
	return nil
}
