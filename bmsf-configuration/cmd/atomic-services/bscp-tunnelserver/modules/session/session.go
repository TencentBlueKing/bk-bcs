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

package session

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	"bk-bscp/cmd/atomic-services/bscp-tunnelserver/modules"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/internal/types"
	"bk-bscp/pkg/logger"
)

// SidecarInstance is sidecar app instance struct in session manager module.
type SidecarInstance struct {
	BizID   string
	AppID   string
	CloudID string
	IP      string
	Path    string
	Labels  string
}

// Session is sidecar app instance session struct.
type Session struct {
	// Sidecar instance content.
	Sidecar SidecarInstance

	// PluginID is gse plugin id.
	PluginID string

	// CloudID is gse plugin cloudid.
	CloudID int32

	// PubFunc is gse publish handle func.
	PubFunc func(sendProcesserMessage *modules.GSESendProcesserMessage) error
}

// Manager is session manager.
type Manager struct {
	// config viper as context here.
	viper *viper.Viper

	// sessions cache, appid -> app instance session cache.
	sessions gcache.Cache

	// flush records used to control session flush interval.
	flushRecord   map[string]int64
	flushRecordMu sync.RWMutex
}

// NewManager creates new Manager.
func NewManager(viper *viper.Viper) *Manager {
	return &Manager{
		viper:       viper,
		sessions:    gcache.New(0).EvictType(gcache.TYPE_SIMPLE).Build(),
		flushRecord: make(map[string]int64, 0),
	}
}

// SessionCount returns session count.
func (mgr *Manager) SessionCount() (int64, error) {
	var count int64

	for _, appID := range mgr.sessions.Keys(false) {
		sessions, err := mgr.sessions.Get(appID)
		if err != nil || sessions == nil {
			return 0, fmt.Errorf("invalid session cache, %+v, %+v", sessions, err)
		}
		cache, ok := sessions.(gcache.Cache)
		if !ok {
			return 0, errors.New("invalid session cache content, can't get the session this moment")
		}
		count += int64(cache.Len(true))

		// purge.
		for _, key := range cache.Keys(false) {
			cache.Get(key)
		}
	}
	return count, nil
}

func (mgr *Manager) sessionKey(cloudID, ip, path string) string {
	return cloudID + ":" + ip + ":" + filepath.Clean(path)
}

// FlushSession flushs instance session.
func (mgr *Manager) FlushSession(instance *pbcommon.AppInstance, pluginID string, cloudID int32,
	pubFunc func(sendProcesserMessage *modules.GSESendProcesserMessage) error,
	addedFunc func(interface{}, interface{}) error,
	evictedFunc func(interface{}, interface{}),
	timeout time.Duration) error {

	if instance == nil {
		return errors.New("invalid sidecar instance struct: nil")
	}
	cache := gcache.New(0).EvictType(gcache.TYPE_SIMPLE).EvictedFunc(evictedFunc).Build()

	m, err := mgr.sessions.Get(instance.AppId)
	if err != nil && err != gcache.KeyNotFoundError {
		return err
	}
	if err == nil {
		v, ok := m.(gcache.Cache)
		if !ok {
			return errors.New("can't flush session, invalid app session cache struct")
		}
		cache = v
	}

	session := &Session{
		Sidecar: SidecarInstance{
			BizID:   instance.BizId,
			AppID:   instance.AppId,
			CloudID: instance.CloudId,
			IP:      instance.Ip,
			Path:    instance.Path,
			Labels:  instance.Labels,
		},
		PluginID: pluginID,
		CloudID:  cloudID,
		PubFunc:  pubFunc,
	}

	sessionKey := mgr.sessionKey(instance.CloudId, instance.Ip, instance.Path)

	// flush session in database or cache now.
	mgr.flushRecordMu.RLock()
	lastFlushTimestamp := mgr.flushRecord[sessionKey]
	mgr.flushRecordMu.RUnlock()

	// last flush timestamp.
	interval := time.Now().Unix() - lastFlushTimestamp

	// flush session when there is no session cache or reach the interval limit.
	if _, err := cache.Get(sessionKey); err != nil {
		// NOTE: record cache missing. session missing means the instance flush
		// action is block or the instance back from offline state.
		logger.Warnf("flush session get local cache, instance: %+v, %+v", instance, err)
	}

	// flush database session.
	if interval > int64(types.AppInstanceFlushDBSessionInterval/time.Second) {
		if err := addedFunc(sessionKey, session); err == nil {
			mgr.flushRecordMu.Lock()
			mgr.flushRecord[sessionKey] = time.Now().Unix()
			mgr.flushRecordMu.Unlock()
		}
	}

	if err := cache.SetWithExpire(sessionKey, session, timeout); err != nil {
		return err
	}

	if err := mgr.sessions.Set(instance.AppId, cache); err != nil {
		return err
	}

	return nil
}

// GetSessions returns sidecar app instance sessions of target app.
func (mgr *Manager) GetSessions(appID string) ([]*Session, error) {
	sessions := []*Session{}

	sCache, err := mgr.sessions.Get(appID)
	if err == gcache.KeyNotFoundError {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if sCache == nil {
		return nil, nil
	}

	cache, ok := sCache.(gcache.Cache)
	if !ok {
		return nil, errors.New("invalid session cache content, can't get the session this moment")
	}

	for _, key := range cache.Keys(false) {
		s, err := cache.Get(key)
		if err != nil || s == nil {
			logger.Warn("get sessions, invalid session cache, key[%+v], %+v, %+v", key, s, err)
			continue
		}
		session, ok := s.(*Session)
		if !ok {
			return nil, errors.New("invalid session cache content, can't get the session this moment")
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

// DeleteSession deletes session of target sidecar connection.
func (mgr *Manager) DeleteSession(instance *pbcommon.AppInstance) error {
	if instance == nil {
		return errors.New("invalid sidecar instance: nil")
	}

	sCache, err := mgr.sessions.Get(instance.AppId)
	if err == gcache.KeyNotFoundError {
		return nil
	} else if err != nil {
		return err
	}

	if sCache == nil {
		return nil
	}
	cache, ok := sCache.(gcache.Cache)
	if !ok {
		return errors.New("invalid session cache content, can't delete the session this moment")
	}

	cache.Remove(mgr.sessionKey(instance.CloudId, instance.Ip, instance.Path))
	return nil
}
