package k8s

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/config"
)

type LogManager struct {
	userManagerCli *bcsapi.UserManagerCli
	configs        *config.ManagerConfig
	controllers    map[string]*ClusterLogController
	caFile         string
}

func NewManager(conf *config.ManagerConfig) *LogManager {
	manager := &LogManager{
		configs: conf,
		caFile:  conf.CAFile,
	}
	cli := bcsapi.NewClient(&conf.BcsApiConfig)
	manager.userManagerCli = cli.UserManager()
	return manager
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
			controller, err := NewClusterLogController(&config.ControllerConfig{Credential: cc, CAFile: m.caFile})
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
