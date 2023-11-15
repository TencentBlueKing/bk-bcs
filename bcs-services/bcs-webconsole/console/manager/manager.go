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

// Package manager xxx
package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/pborman/ansi"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit/record"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// ManagerFunc 自定义 Manager 函数
type ManagerFunc func(podCtx *types.PodContext) error // nolint

var commandDelay = make(map[string]string, 0)

// ConsoleManager websocket 流式处理器
type ConsoleManager struct {
	ctx            context.Context
	ConnTime       time.Time // 连接时间
	LastInputTime  time.Time // 更新ws时间
	keyWaitingTime time.Time // 记录 webconsole key 响应时间
	podCtx         *types.PodContext
	managerFuncs   []ManagerFunc
	cmdParser      *audit.CmdParse
	recorder       *record.ReplyRecorder
}

// NewConsoleManager :
func NewConsoleManager(ctx context.Context, podCtx *types.PodContext,
	terminalSize *types.TerminalSize) (*ConsoleManager, error) {
	now := time.Now()
	mgr := &ConsoleManager{
		ctx:            ctx,
		ConnTime:       now,
		LastInputTime:  now,
		keyWaitingTime: now,
		podCtx:         podCtx,
		managerFuncs:   []ManagerFunc{},
		cmdParser:      audit.NewCmdParse(),
	}

	// 初始化 terminal record
	recorder, err := record.NewReplayRecord(ctx, mgr.podCtx, terminalSize)
	if err != nil {
		logger.Errorf("init ReplayRecord failed, err %s", err)
		return nil, err
	}
	mgr.recorder = recorder

	return mgr, nil
}

// AddMgrFunc 添加自定义函数
func (c *ConsoleManager) AddMgrFunc(mgrFunc ManagerFunc) {
	c.managerFuncs = append(c.managerFuncs, mgrFunc)
}

// HandleResizeMsg 处理 resize 数据
func (c *ConsoleManager) HandleResizeMsg(resizeMsg *types.TerminalSize) error {
	// replay 记录终端大小变化
	replaySize := fmt.Sprintf("%vx%v", resizeMsg.Cols, resizeMsg.Rows)
	record.RecordResizeEvent(c.recorder, []byte(replaySize))

	return nil
}

// HandleInputMsg : 处理输入数据流
func (c *ConsoleManager) HandleInputMsg(msg []byte) ([]byte, error) {
	now := time.Now()
	// 更新ws时间
	c.LastInputTime = now
	c.keyWaitingTime = now

	// 命令行解析与审计
	_, ss, err := ansi.Decode(msg)
	if err != nil {
		return msg, nil
	}

	c.cmdParser.Cmd = ss
	c.cmdParser.InputSlice = append(c.cmdParser.InputSlice, ss)

	return msg, nil
}

// HandleOutputMsg : 处理输出数据流
func (c *ConsoleManager) HandleOutputMsg(msg []byte) ([]byte, error) {
	return msg, nil
}

// HandlePostOutputMsg : 后置输出数据流处理，在HandleOutputMsg之后, 发送给websocket之前, 不能修改数据，没有错误返回
func (c *ConsoleManager) HandlePostOutputMsg(msg []byte) {
	// 命令行解析与审计
	c.auditCmd(msg)

	// replay 记录数据流
	record.RecordOutputEvent(c.recorder, msg)

	// 性能统计，按照用户设置的key统计
	if len(msg) > 0 && commandDelay[c.podCtx.Username] != "" {
		go userDelayCollect(string(msg[0]), c)
	}
}

// Run : Manager 后台任务等
func (c *ConsoleManager) Run(ctx *gin.Context) error {
	interval := time.NewTicker(10 * time.Second)
	defer interval.Stop()

	delayDataSync := time.NewTicker(1 * time.Second)
	defer delayDataSync.Stop()
	// 结束会话时,处理缓存/关闭文件
	defer c.recorder.End()

	for {
		select {
		case <-c.ctx.Done():
			logger.Infof("close %s ConsoleManager done", c.podCtx.PodName)
			return nil
		case <-interval.C:
			if err := c.handleIdleTimeout(ctx); err != nil {
				return err
			}
			// 自定义函数
			for _, managerFunc := range c.managerFuncs {
				if err := managerFunc(c.podCtx); err != nil {
					return err
				}
			}
			// 定时写入文件
			c.recorder.Flush()
		case <-delayDataSync.C:
			// 定时写入用户设置延时开关数据
			delayData, err := storage.GetDefaultRedisSession().Client.HGetAll(ctx, types.ConsoleKey).Result()
			if err != nil {
				logger.Warnf("failed to synchronize redis data, err: %s", err.Error())
				continue
			}
			commandDelay = delayData
		}
	}
}

// auditCmd 命令行审计, 不能影响主流程
func (c *ConsoleManager) auditCmd(outputMsg []byte) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("[audit cmd panic], err: %v, debug strace: %s", r, debug.Stack())
		}
	}()

	// 输入输出映射,用于查找历史命令
	out, ss, err := ansi.Decode(outputMsg)
	if err != nil {
		logger.Error("decode output error: %s", err)
		return
	}

	// DOTO:历史命令问题,可能解析问题导致
	if strings.ReplaceAll(string(ss.Code), "\b", "") == "" {
		rex := regexp.MustCompile("\\x1b\\[\\d+P") // nolint
		l := rex.Split(string(out), -1)
		ss.Code = ansi.Name(l[len(l)-1])
	}
	// 时序性问题不可避免
	c.cmdParser.CmdResult[c.cmdParser.Cmd] = ss

	if c.cmdParser.Cmd != nil && c.cmdParser.Cmd.Code == "\r" {
		cmd := audit.ResolveInOut(c.cmdParser)
		if cmd != "" {
			logger.Infof("UserName=%s  SessionID=%s  Command=%s",
				c.podCtx.Username, c.podCtx.SessionId, cmd)
		}
	}
}

func (c *ConsoleManager) handleIdleTimeout(ctx *gin.Context) error {
	nowTime := time.Now()
	idleTime := nowTime.Sub(c.LastInputTime)
	if idleTime > c.podCtx.GetConnIdleTimeout() {
		// BCS Console 已经分钟无操作
		msg := i18n.GetMessage(ctx, "BCS Console 已经{}分钟无操作", map[string]int64{"time": int64(idleTime.Minutes())})
		logger.Infof("conn idle timeout, close session %s, idle time, %s", c.podCtx.PodName, idleTime)
		return errors.New(msg)
	}

	loginTime := nowTime.Sub(c.ConnTime).Seconds()
	if loginTime > LoginTimeout {
		// BCS Console 使用已经超过{}小时，请重新登录
		msg := i18n.GetMessage(ctx, "BCS Console 使用已经超过{}小时，请重新登录", map[string]int{"time": LoginTimeout / 60})
		logger.Infof("tick timeout, close session %s, login time, %.2f", c.podCtx.PodName, loginTime)
		return errors.New(msg)
	}
	return nil
}

// 用户延时命令统计数据
func userDelayCollect(msg string, c *ConsoleManager) {
	// 取出用户设置的延时key
	// 匹配子字符串，如果包含则表示开启了命令延时统计
	msgPart := "\"cluster_id\":\"" + c.podCtx.ClusterId + "\",\"enabled\":true,\"console_key\":\"" + msg
	if strings.Contains(commandDelay[c.podCtx.Username], msgPart) {
		delayData := types.DelayData{
			ClusterId:    c.podCtx.ClusterId,
			TimeDuration: time.Since(c.keyWaitingTime).String(),
			CreateTime:   time.Now().Format(time.DateTime),
			SessionId:    c.podCtx.SessionId,
			PodName:      c.podCtx.PodName,
			CommandKey:   msg,
		}
		delayDataByte, err := json.Marshal(delayData)
		if err != nil {
			logger.Errorf("json Marshal failed, err: %s", err.Error())
			return
		}
		// 查看用户是否已经有统计数据在Redis中
		listLen := storage.GetDefaultRedisSession().Client.LLen(
			c.ctx, types.DelayUser+c.podCtx.Username).Val()
		// 往Redis数据
		err = storage.GetDefaultRedisSession().Client.RPush(
			c.ctx, types.DelayUser+c.podCtx.Username, string(delayDataByte)).Err()
		if err != nil {
			logger.Errorf("redis list push failed, err: %s", err.Error())
			return
		}
		// 没有数据的情况下设置列表过期时间，暂定一天
		if listLen == 0 {
			// 列表设置过期时间
			storage.GetDefaultRedisSession().Client.Expire(
				c.ctx, types.DelayUser+c.podCtx.Username, types.DelayUserExpire)
		}
	}
}
