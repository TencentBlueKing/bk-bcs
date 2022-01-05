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
	"time"

	"github.com/bluele/gcache"

	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/internal/safeviper"
	"bk-bscp/pkg/logger"
)

// SidecarInstance is sidecar struct in connection manager module.
type SidecarInstance struct {
	BizID   string
	AppID   string
	CloudID string
	IP      string
	Path    string
	Labels  string
}

// Session is sidecar connection session struct.
type Session struct {
	// Sidecar instance content.
	Sidecar SidecarInstance

	// PubCh is publishing channel for connserver.
	PubCh chan interface{}
}

// Manager is session manager.
type Manager struct {
	// config viper as context here.
	viper *safeviper.SafeViper

	// sessions cache, appid -> sidecar instance session cache.
	sessions gcache.Cache
}

// NewManager creates new Manager.
func NewManager(viper *safeviper.SafeViper) *Manager {
	return &Manager{
		viper:    viper,
		sessions: gcache.New(0).EvictType(gcache.TYPE_SIMPLE).Build(),
	}
}

// ConnCount returns connection count.
func (mgr *Manager) ConnCount() (int64, error) {
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
	}

	return count, nil
}

func (mgr *Manager) sessionKey(cloudID, ip, path string) string {
	return cloudID + ":" + ip + ":" + filepath.Clean(path)
}

// FlushSession flushs sidecar connection session.
func (mgr *Manager) FlushSession(ping *pb.SCCMDPing, ch chan interface{}) error {
	if ping == nil {
		return errors.New("invalid ping struct: nil")
	}
	if ch == nil {
		return errors.New("invalid notification channel: nil")
	}

	cache := gcache.New(0).EvictType(gcache.TYPE_SIMPLE).Build()

	m, err := mgr.sessions.Get(ping.AppId)
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
			BizID:   ping.BizId,
			AppID:   ping.AppId,
			CloudID: ping.CloudId,
			IP:      ping.Ip,
			Path:    ping.Path,
			Labels:  ping.Labels,
		},
		PubCh: ch,
	}

	if err := cache.SetWithExpire(mgr.sessionKey(ping.CloudId, ping.Ip, ping.Path), session,
		time.Duration(ping.Timeout)*time.Second); err != nil {
		return err
	}

	if err := mgr.sessions.Set(ping.AppId, cache); err != nil {
		return err
	}

	return nil
}

// GetSessions returns sidecar instance sessions of target app.
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

// GetAllSessions returns all sidecar instance
func (mgr *Manager) GetAllSessions() ([]*Session, error) {
	sessions := []*Session{}
	objMap := mgr.sessions.GetALL(false)

	for _, obj := range objMap {
		cache, ok := obj.(gcache.Cache)
		if !ok {
			logger.Warnf("invalid session cache content %+v, skip", obj)
			continue
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
	}

	return sessions, nil
}

// DeleteSession deletes session of target sidecar connection.
func (mgr *Manager) DeleteSession(ins *SidecarInstance) error {
	if ins == nil {
		return errors.New("invalid sidecar instance: nil")
	}

	sCache, err := mgr.sessions.Get(ins.AppID)
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

	cache.Remove(mgr.sessionKey(ins.CloudID, ins.IP, ins.Path))

	return nil
}
