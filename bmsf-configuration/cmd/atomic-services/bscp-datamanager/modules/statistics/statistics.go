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

package statistics

import (
	"time"

	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-datamanager/modules/metrics"
	"bk-bscp/internal/database"
	"bk-bscp/internal/dbsharding"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

const (
	// statInterval24h is 24h stat interval.
	statInterval24h = "24h"
)

// Collector is internal business statistics collector.
type Collector struct {
	viper            *viper.Viper
	metricsCollector *metrics.Collector

	// db sharding manager.
	smgr *dbsharding.ShardingManager

	isStop bool
}

// NewCollector creates a new internal statistics collector.
func NewCollector(viper *viper.Viper, metricsCollector *metrics.Collector,
	smgr *dbsharding.ShardingManager) *Collector {

	return &Collector{viper: viper, metricsCollector: metricsCollector, smgr: smgr}
}

// statBusiness stats business num.
func (c *Collector) statBusiness() (int64, error) {
	return c.smgr.QueryShardingCount()
}

// statBusinessApp stats business app count.
func (c *Collector) statBusinessApp() (map[string]int64, error) {
	// query all business shardings.
	shardings, err := c.smgr.QueryShardingList()
	if err != nil {
		return nil, err
	}

	stat := make(map[string]int64, 0)

	// stat each business.
	for _, sharding := range shardings {
		// sharding database.
		sd, err := c.smgr.ShardingDB(sharding.Key)
		if err != nil {
			return nil, err
		}

		// query app count.
		var totalCount int64

		if err := sd.DB().
			Model(&database.App{}).
			Where(&database.App{BizID: sharding.Key}).
			Where("Fstate = ?", pbcommon.CommonState_CS_VALID).
			Count(&totalCount).Error; err != nil {
			return nil, err
		}
		stat[sharding.Key] = totalCount
	}

	return stat, nil
}

func (c *Collector) statBusinessAppConfg() (map[string]map[string]int64, error) {
	// query all business shardings.
	shardings, err := c.smgr.QueryShardingList()
	if err != nil {
		return nil, err
	}

	stat := make(map[string]map[string]int64, 0)

	// stat each business.
	for _, sharding := range shardings {
		// sharding database.
		sd, err := c.smgr.ShardingDB(sharding.Key)
		if err != nil {
			return nil, err
		}

		// query apps.
		apps := []database.App{}

		if err := sd.DB().
			Order("Fupdate_time DESC, Fid DESC").
			Where(&database.App{BizID: sharding.Key}).
			Find(&apps).Error; err != nil {
			return nil, err
		}

		// query app config count.
		subStat := make(map[string]int64, 0)

		for _, app := range apps {
			var totalCount int64

			if err := sd.DB().
				Model(&database.Config{}).
				Where(&database.Config{BizID: sharding.Key, AppID: app.AppID}).
				Count(&totalCount).Error; err != nil {
				return nil, err
			}
			subStat[app.AppID] = totalCount
		}
		stat[sharding.Key] = subStat
	}

	return stat, nil
}

func (c *Collector) statBusinessAppRelease() (map[string]map[string]map[string]int64, error) {
	// query all business shardings.
	shardings, err := c.smgr.QueryShardingList()
	if err != nil {
		return nil, err
	}

	stat := make(map[string]map[string]map[string]int64, 0)

	// stat each business.
	for _, sharding := range shardings {
		// sharding database.
		sd, err := c.smgr.ShardingDB(sharding.Key)
		if err != nil {
			return nil, err
		}

		// query apps.
		apps := []database.App{}

		if err := sd.DB().
			Order("Fupdate_time DESC, Fid DESC").
			Where(&database.App{BizID: sharding.Key}).
			Find(&apps).Error; err != nil {
			return nil, err
		}

		// stat release count.
		subStat := make(map[string]map[string]int64, 0)

		// stat release count in last 24h.
		subStat24h := make(map[string]int64, 0)

		for _, app := range apps {
			var totalCount int64

			// last 24h.
			hours, _ := time.ParseDuration("-24h")
			last24h := time.Now().Add(hours).Format("2006-01-02 15:04:05")

			if err := sd.DB().
				Model(&database.Release{}).
				Where(&database.Release{BizID: sharding.Key, AppID: app.AppID}).
				Where("Fstate IN (?, ?)", pbcommon.ReleaseState_RS_PUBLISHED, pbcommon.ReleaseState_RS_ROLLBACKED).
				Where("Fupdate_time >= ?", last24h).
				Count(&totalCount).Error; err != nil {
				return nil, err
			}
			subStat24h[app.AppID] = totalCount
		}

		subStat[statInterval24h] = subStat24h
		stat[sharding.Key] = subStat
	}

	return stat, nil
}

func (c *Collector) collectBusiness() error {
	// stat business.
	bizNum, err := c.statBusiness()
	if err != nil {
		return err
	}

	// stat business app.
	appStat, err := c.statBusinessApp()
	if err != nil {
		return err
	}

	// stat business app config.
	configStat, err := c.statBusinessAppConfg()
	if err != nil {
		return err
	}

	// stat business app release.
	releaseStat, err := c.statBusinessAppRelease()
	if err != nil {
		return err
	}

	// prometheus metrics.
	c.metricsCollector.StatBusiness(bizNum, appStat, configStat, releaseStat)

	return nil
}

func (c *Collector) collectDBStatus() error {
	// stat database questions.
	questions, err := c.smgr.ShowQuestionsStatus()
	if err != nil {
		return err
	}

	// stat database threads connected.
	threadsConnected, err := c.smgr.ShowThreadsConnectedStatus()
	if err != nil {
		return err
	}

	// prometheus metrics.
	c.metricsCollector.StatDBStatus(questions.VarValue, threadsConnected.VarValue)

	return nil
}

func (c *Collector) start() {
	ticker := time.NewTicker(c.viper.GetDuration("metrics.internalStatInterval"))
	defer ticker.Stop()

	for {
		<-ticker.C

		if c.isStop {
			break
		}

		// stat business.
		if err := c.collectBusiness(); err != nil {
			logger.Warnf("business internal statistics, %+v", err)
		}

		// stat database status.
		if err := c.collectDBStatus(); err != nil {
			logger.Warnf("database status statistics, %+v", err)
		}
	}
}

// Start starts the internal statistics collector.
func (c *Collector) Start() {
	go c.start()
}

// Stop stops the internal statistics collector.
func (c *Collector) Stop() {
	c.isStop = true
}
