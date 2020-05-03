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

package structs

import ()

// SignallingType is type for signalling.
type SignallingType int

const (
	// SignallingTypePublish is type of publishing.
	SignallingTypePublish SignallingType = iota

	// SignallingTypeRollback is type of rollback publishing.
	SignallingTypeRollback

	// TODO other signalling type...

	// SignallingTypeEnd is end of the types.
	SignallingTypeEnd
)

// Signalling struct.
type Signalling struct {
	// Type is signalling type.
	Type SignallingType

	// Publishing is publishing signalling.
	Publishing Publishing
}

// Publishing notification content.
type Publishing struct {
	// Bid is business id.
	Bid string

	// Appid is app id.
	Appid string

	// Cfgsetid is configset id.
	Cfgsetid string

	// CfgsetName is configset name.
	CfgsetName string

	// CfgsetFpath is configset fpath.
	CfgsetFpath string

	// Serialno is release serial num.
	Serialno uint64

	// Releaseid is release id.
	Releaseid string

	// Strategies is release strategies.
	Strategies string
}

// Metadata is load information metadata for resource report.
type Metadata struct {
	// Metadata is load information metadata.
	Metadata string `json:"Metadata"`
}

// ConnServer is connserver resource instance struct.
type ConnServer struct {
	// IP is connserver ip.
	IP string

	// Port is connserver port.
	Port int

	// ConnCount is count of connections in connserver.
	ConnCount int64
}

// RuleKeyType is type of template rule.
type RuleKeyType int

const (
	// RuleKeyTypeCluster is cluster rule type.
	RuleKeyTypeCluster RuleKeyType = iota

	// RuleKeyTypeZone is zone rule type.
	RuleKeyTypeZone
)

// Rule is bscp config template rule, template server would renders configs
// with the template base on GO inner template engine. When the cluster or zone
// name matched, the variables will be writed into the configs.
type Rule struct {
	// Type is template rule key type, 0 is cluster, 1 is zone.
	Type RuleKeyType `json:"type"`

	// Name is rule key name(cluster or zone name).
	Name string `json:"name"`

	// Variables is template rendering variables.
	Variables map[string]interface{} `json:"vars"`
}

/*
   Template Rule List Example:
   [
       {
           "type": 0,
           "name": "cluster1",
           "vars": {
               "k1": "v1a",
               "k2": 0,
               "k3": ["v3a", "v3b"]
           }
       },
       {
           "type": 1,
           "name": "zone1",
           "vars": {
               "k1": "v1b",
               "k2": 1,
               "k3": ["v3c", "v3d"]
           }
       }
   ]
*/

// RuleList is bscp configs template rule list.
type RuleList []Rule

const (
	// IntegrationMetadataKindBusiness is integration metadata kind for business.
	IntegrationMetadataKindBusiness = "business"

	// IntegrationMetadataKindConstruction is integration metadata kind for Construction.
	IntegrationMetadataKindConstruction = "construction"

	// IntegrationMetadataKindCommit is integration metadata kind for commit.
	IntegrationMetadataKindCommit = "commit"

	// IntegrationMetadataKindPublish is integration metadata kind for publishing.
	IntegrationMetadataKindPublish = "publish"

	// IntegrationMetadataKindEffect is integration metadata kind for effect.
	IntegrationMetadataKindEffect = "effect"
)

const (
	// IntegrationMetadataOpCreate is integration metadata op type for create.
	IntegrationMetadataOpCreate = "create"

	// IntegrationMetadataOpCommit is integration metadata op type for commit.
	IntegrationMetadataOpCommit = "commit"

	// IntegrationMetadataOpPub is integration metadata op type for publish.
	IntegrationMetadataOpPub = "publish"

	// IntegrationMetadataOpQuery is integration metadata op type for query.
	IntegrationMetadataOpQuery = "query"

	// IntegrationMetadataOpRollback is integration metadata op type for rollback.
	IntegrationMetadataOpRollback = "rollback"
)

// IntegrationMetadataZone is struct for create zone in construction mode.
type IntegrationMetadataZone struct {
	// Name is zone name.
	Name string `yaml:"name"`

	// Memo is common backup information.
	Memo string `yaml:"memo"`
}

// IntegrationMetadataCluster is struct for create cluster in construction mode.
type IntegrationMetadataCluster struct {
	// Name is cluster name.
	Name string `yaml:"name"`

	// RClusterid is related rclusterid of cluster.
	RClusterid string `yaml:"rclusterid"`

	// Memo is common backup information.
	Memo string `yaml:"memo"`

	// Zones is struct for target cluster to create zone constructions.
	Zones []IntegrationMetadataZone `yaml:"zones"`
}

// IntegrationMetadata is integration metadata struct.
type IntegrationMetadata struct {
	// Kind is resource kind spec.
	Kind string `yaml:"kind"`

	// Version is version spec.
	Version string `yaml:"version"`

	// Op is resource operate type.
	Op string `yaml:"op"`

	// Spec is resource op metadata main block.
	Spec struct {
		// BusinessName is target business name.
		BusinessName string `yaml:"businessName"`

		// AppName is target application name.
		AppName string `yaml:"appName"`

		// ConfigSetName is target configset name.
		ConfigSetName string `yaml:"configSetName"`

		// ConfigSetFpath is sub path of target configset.
		ConfigSetFpath string `yaml:"configSetFpath"`

		// Depid is department id for business.
		Depid string `yaml:"depid"`

		// Dbid is database sharding instance id for business.
		Dbid string `yaml:"dbid"`

		// Dbname is mysql database name for business.
		Dbname string `yaml:"dbname"`

		// Memo is common backup information.
		Memo string `yaml:"memo"`

		// DeployType is deploy type of application.
		DeployType int32 `yaml:"deployType"`
	} `yaml:"spec"`

	// Construction is used for build app constructions.
	Construction struct {
		// Clusters is used for create cluster constructions.
		Clusters []IntegrationMetadataCluster `yaml:"clusters"`
	} `yaml:"construction"`

	// Relase is publishing release stuff.
	Release struct {
		// Name is release name.
		Name string `yaml:"name"`

		// Commitid is inner id of target commit.
		Commitid string `yaml:"commitid"`

		// Releaseid is inner id of target release, used to rollback target release.
		Releaseid string `yaml:"releaseid"`

		// NewReleaseid is inner id of target new release wanted to rollback.
		NewReleaseid string `yaml:"newReleaseid"`

		// StrategyName is name of target app publishing strategy, create if not exist.
		StrategyName string `yaml:"strategyName"`

		// Strategy is release publishing strategies.
		Strategy struct {
			// ClusterNames is cluster list used for release publishing strategy match.
			ClusterNames []string `yaml:"clusterNames"`

			// ZoneNames is zone list used for release publishing strategy match.
			ZoneNames []string `yaml:"zoneNames"`

			// Dcs is datacenter list used for release publishing strategy match.
			Dcs []string `yaml:"dcs"`

			// IPs is ip list used for release publishing strategy match.
			IPs []string `yaml:"ips"`

			// Labels is label list used for release publishing strategy match.
			Labels map[string]string `yaml:"labels"`
		} `yaml:"strategy"`
	} `yaml:"release"`

	// Template is bscp configs template stuff.
	Template struct {
		// Templateid is template id(may be from 3rd system not in bscp).
		Templateid string `yaml:"templateid"`

		// Template is bscp configs template content.
		Template string `yaml:"template"`

		// TemplateRule is template rules.
		TemplateRule string `yaml:"templateRule"`
	} `yaml:"template"`

	// Configs is bscp configs content.
	Configs string `yaml:"configs"`

	// Changes is commit changes.
	Changes string `yaml:"changes"`
}
