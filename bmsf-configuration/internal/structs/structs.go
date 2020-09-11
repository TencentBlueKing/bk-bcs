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

	// SignallingTypeReload is type of reload publishing.
	SignallingTypeReload

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

// EffectInfo is effect info for reload action.
type EffectInfo struct {
	// Cfgsetid is configset id.
	Cfgsetid string

	// Releaseid is release id.
	Releaseid string
}

// ReloadSpec is reload specs.
type ReloadSpec struct {
	// MultiReleaseid is multi release id.
	MultiReleaseid string

	// Info is effect infos for reload action.
	Info []EffectInfo

	// Rollback is rollback reload flag.
	Rollback bool
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

	// ReloadSpec is spec of reload action.
	ReloadSpec ReloadSpec
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

/*
	Template Rule List Example:
	[
		{
			"cluster": "cluster1",
			"clusterLabels": {
				"environment": "test"
			},
			"vars": {
				"clusterVar": "cluster var 1"
			}
		},
		{
			"cluster": "cluster1",
			"clusterLabels": {
				"environment": "test"
			},
			"zones": [
				{
					"zone": "zone1",
					"vars": {
						"zoneVar": "zone var 1"
					}
				},
				{
					"zone": "zone2",
					"vars": {
						"zoneVar": "zone var 2"
					}
				}
			]
		},
		{
			"cluster": "cluster1",
			"clusterLabels": {
				"environment": "test"
			},
			"vars": {
				"clusterVar": "cluster var 2"
			}
			"zones": [
				{
					"zone": "zone1",
					"vars": {
						"zoneVar": "zone var 3"
					}
					"instances": [
						{
							"index": "127.0.0.1"
						},
						{
							"index": "127.0.0.2"
						}
					]
				},
				{
					"zone": "zone2",
					"instances": [
						{
							"index": "127.0.0.3"
						},
						{
							"index": "127.0.0.4"
						}
					]
				}
			]
		}
	]
*/

// RuleInstance is bscp config template rule instance, a rule instance generate a certain config
type RuleInstance struct {
	// Index is index of config instance of centain zone
	Index string `json:"index"`

	// Variables is template rendering variables.
	Variables map[string]interface{} `json:"vars"`
}

// RuleZone is bscp config template rule for certain rule
type RuleZone struct {
	// Zone zone name
	Zone string `json:"zone"`

	// Instances rule instances
	Instances []*RuleInstance `json:"instances"`

	// Variables is extra Zone variables
	Variables map[string]interface{} `json:"vars"`
}

// Rule is bscp config template rule, template server would renders configs
type Rule struct {
	// Cluster cluster name
	Cluster string `json:"cluster"`

	// ClusterLabels cluster labels
	ClusterLabels map[string]string `json:"clusterLabels"`

	// Zone zone name.
	Zones []*RuleZone `json:"zones"`

	// Variables is extra Cluster variables
	Variables map[string]interface{} `json:"vars"`
}

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

	// IntegrationMetadataKindReload is integration metadata kind for reload.
	IntegrationMetadataKindReload = "reload"
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

	// IntegrationMetadataOpReload is integration metadata op type for reload.
	IntegrationMetadataOpReload = "reload"
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

		// MultiReleaseid is inner id of target multi release, used for target release actions.
		MultiReleaseid string `yaml:"multiReleaseid"`

		// Rollback reload flag.
		Rollback bool `yaml:"rollback"`

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

			// Labels is OR label list used for release publishing strategy match.
			Labels map[string]string `yaml:"labels"`

			// LabelsAnd is AND label list used for release publishing strategy match.
			LabelsAnd map[string]string `yaml:"labelsAnd"`
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
