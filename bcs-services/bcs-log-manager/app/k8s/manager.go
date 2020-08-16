package k8s

import (
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
	bcsv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/pkg/apis/bk-bcs/v1"
)

const (
	KubeSystemNamespace = "kube-system"
	BCSSystemNamespace  = "bcs-system"
	KubePublicNamespace = "kube-public"
)

var SystemNamspaces = []string{"kube-system", "bcs-system", "kube-public"}

type LogManager struct {
	userManagerCli *bcsapi.UserManagerCli
	config         *config.ManagerConfig
	controllers    map[string]*ClusterLogController
}

func NewManager(conf *config.ManagerConfig) *LogManager {
	manager := &LogManager{
		config: conf,
	}
	cli := bcsapi.NewClient(&conf.BcsApiConfig)
	manager.userManagerCli = cli.UserManager()
	manager.addSystemCollectionConfig()
	return manager
}

func (m *LogManager) addSystemCollectionConfig() {
	collectConfig := config.CollectionConfig{
		ConfigNamespace: "default",
		ConfigName:      "",
		ConfigSpec: bcsv1.BcsLogConfigSpec{
			ConfigType:        "custom",
			Stdout:            true,
			StdDataId:         m.config.SystemDataID,
			WorkloadNamespace: "",
			PodLabels:         true,
			LogTags: map[string]string{
				"platform": "bcs",
			},
		},
		ClusterIDs: "",
	}
	for _, namespace := range SystemNamspaces {
		collectConfig.ConfigName = fmt.Sprintf("%s-%s", LogConfigKind, namespace)
		collectConfig.ConfigSpec.WorkloadNamespace = namespace
		m.config.CollectionConfigs = append(m.config.CollectionConfigs, collectConfig)
	}
}

func (m *LogManager) Start() {
	go m.run()
}

func (m *LogManager) run() {
	var cnt int64
	cnt = 0
	for {
		ccinfo, err := m.userManagerCli.ListAllClusters()
		if err != nil {
			blog.Errorf("ListAllClusters failed: %s", err.Error())
			goto WaitLabel
		}
		cnt++
		for _, cc := range ccinfo {
			if _, ok := m.controllers[cc.ClusterID]; ok {
				m.controllers[cc.ClusterID].SetTick(cnt)
				continue
			}
			controller, err := NewClusterLogController(&config.ControllerConfig{Credential: cc, CAFile: m.config.CAFile})
			if err != nil {
				blog.Errorf("Create Cluster Log Controller failed, Cluster Id: %s, Cluster Domain: %s, error info: %s", cc.ClusterID, cc.ClusterDomain, err.Error())
				continue
			}
			controller.SetTick(cnt)
			m.controllers[cc.ClusterID] = controller
			controller.Start()
		}
		for k, v := range m.controllers {
			if v.GetTick() == cnt {
				continue
			}
			m.controllers[k].Stop()
			delete(m.controllers, k)
		}
		m.distributeTasks()
	WaitLabel:
		<-time.After(time.Minute)
	}
}

func (m *LogManager) distributeTasks() {
	for _, logconf := range m.config.CollectionConfigs {
		clusters := strings.Split(strings.ToLower(logconf.ClusterIDs), ",")
		for _, clusterid := range clusters {
			if _, ok := m.controllers[clusterid]; !ok {
				blog.Errorf("Wrong cluster ID %s of collection config %+v", clusterid, logconf)
				continue
			}
			m.controllers[clusterid].AddCollectionTask <- &logconf
		}
		if logconf.ClusterIDs == "" {
			for _, ctrl := range m.controllers {
				ctrl.AddCollectionTask <- &logconf
			}
		}
	}
}
