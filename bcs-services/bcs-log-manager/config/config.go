package config

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
)

type ManagerConfig struct {
	CollectionConfigs []CollectionConfig
	BcsApiConfig      bcsapi.Config
	CAFile            string
	SystemDataID      string
}

// CollectionConfig defines some customed infomation of log collection.
// For example, customed dataid of some Cluster.
type CollectionConfig struct {
	//Config Spec.
	ConfigName      string                 `json:"config_name"`
	ConfigNamespace string                 `json:"config_namespace"`
	ConfigSpec      bcsv1.BcsLogConfigSpec `json:"config_spec"`
	// ClusterIDs comma split clusterid
	ClusterIDs string `json:"cluster_ids"`
}

type ControllerConfig struct {
	Credential *bcsapi.ClusterCredential
	CAFile     string
}
