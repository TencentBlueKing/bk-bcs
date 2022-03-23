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

package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// TerminalSize web终端发来的 resize 包
type TerminalSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

// 自定义 Manager 函数
type ManagerFunc func(podCtx *types.PodContext) error

// ConsoleManager websocket 流式处理器
type ConsoleManager struct {
	ctx           context.Context
	ConnTime      time.Time // 连接时间
	LastInputTime time.Time // 更新ws时间
	PodCtx        *types.PodContext
	redisClient   *redis.Client
	managerFuncs  []ManagerFunc
}

// NewConsoleManager :
func NewConsoleManager(ctx context.Context, podCtx *types.PodContext) *ConsoleManager {
	redisClient := storage.GetDefaultRedisSession().Client
	mgr := &ConsoleManager{
		ctx:           ctx,
		ConnTime:      time.Now(),
		LastInputTime: time.Now(),
		PodCtx:        podCtx,
		redisClient:   redisClient,
		managerFuncs:  []ManagerFunc{},
	}

	return mgr
}

// AddMgrFunc 添加自定义函数
func (c *ConsoleManager) AddMgrFunc(mgrFunc ManagerFunc) {
	c.managerFuncs = append(c.managerFuncs, mgrFunc)
}

// HandleInputMsg : 处理输入数据流
func (c *ConsoleManager) HandleInputMsg(msg []byte) ([]byte, error) {
	// 更新ws时间
	c.LastInputTime = time.Now()
	return msg, nil
}

// HandleInputMsg : 处理 Resize 数据流
func (c *ConsoleManager) HandleResizeMsg(msg []byte) (*TerminalSize, error) {
	resizeMsg := TerminalSize{}

	// 解析Json数据
	err := json.Unmarshal(msg, &resizeMsg)
	if err != nil {
		return nil, err
	}

	return &resizeMsg, nil
}

// HandleOutputMsg: 处理输出数据流
func (c *ConsoleManager) HandleOutputMsg(msg []byte) ([]byte, error) {
	return msg, nil
}

// Run: Manager 后台任务等
func (c *ConsoleManager) Run() error {
	interval := time.NewTicker(10 * time.Second)
	defer interval.Stop()

	for {
		select {
		case <-c.ctx.Done():
			logger.Infof("close %s ConsoleManager done", c.PodCtx.PodName)
			return c.ctx.Err()
		case <-interval.C:
			if err := c.tickTimeout(); err != nil {
				return err
			}
			// 自定义函数
			for _, managerFunc := range c.managerFuncs {
				if err := managerFunc(c.PodCtx); err != nil {
					return err
				}
			}
		}
	}
}

func (c *ConsoleManager) tickTimeout() error {
	nowTime := time.Now()
	idleTime := nowTime.Sub(c.LastInputTime).Seconds()
	if idleTime > TickTimeout {
		// BCS Console 已经分钟无操作
		msg := fmt.Sprintf("BCS Console 已经 %d 分钟无操作", TickTimeout/60)
		logger.Infof("tick timeout, close session %s, idle time, %.2f", c.PodCtx.PodName, idleTime)
		return errors.New(msg)
	}

	loginTime := nowTime.Sub(c.ConnTime).Seconds()
	if loginTime > LoginTimeout {
		// BCS Console 使用已经超过{}小时，请重新登录
		msg := fmt.Sprintf("BCS Console 使用已经超过 %d 小时，请重新登录", LoginTimeout/60)
		logger.Infof("tick timeout, close session %s, login time, %.2f", c.PodCtx.PodName, loginTime)
		return errors.New(msg)
	}
	return nil
}

// 提交数据
func (c *ConsoleManager) emit(data *types.AuditRecord) {
	dataByte, _ := json.Marshal(data)
	c.redisClient.RPush(context.Background(), queueName, dataByte)
}

// 审计
func (c *ConsoleManager) startRecord() {
	c.emit(nil)
}
