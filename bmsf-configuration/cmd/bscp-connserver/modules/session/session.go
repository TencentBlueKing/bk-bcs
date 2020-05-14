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
	"time"

	"github.com/bluele/gcache"
	"github.com/spf13/viper"

	pb "bk-bscp/internal/protocol/connserver"
	"bk-bscp/pkg/logger"
)

// SidecarInstance is sidecar struct in connection manager module.
type SidecarInstance struct {
	Bid       string
	Appid     string
	Clusterid string
	Zoneid    string
	Dc        string
	IP        string
	Labels    string
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
	viper *viper.Viper

	// sessions cache, appid -> sidecar instance session cache.
	sessions gcache.Cache
}

// NewManager creates new Manager.
func NewManager(viper *viper.Viper) *Manager {
	return &Manager{
		viper:    viper,
		sessions: gcache.New(0).EvictType(gcache.TYPE_SIMPLE).Build(),
	}
}

// ConnCount returns connection count.
func (mgr *Manager) ConnCount() (int64, error) {
	var count int64

	for _, appid := range mgr.sessions.Keys(false) {
		sessions, err := mgr.sessions.Get(appid)
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

func (mgr *Manager) sessionKey(dc, ip string) string {
	return dc + ":" + ip
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

	m, err := mgr.sessions.Get(ping.Appid)
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
			Bid:       ping.Bid,
			Appid:     ping.Appid,
			Clusterid: ping.Clusterid,
			Zoneid:    ping.Zoneid,
			Dc:        ping.Dc,
			IP:        ping.IP,
			Labels:    ping.Labels,
		},
		PubCh: ch,
	}

	if err := cache.SetWithExpire(mgr.sessionKey(ping.Dc, ping.IP), session, time.Duration(ping.Timeout)*time.Second); err != nil {
		return err
	}

	if err := mgr.sessions.Set(ping.Appid, cache); err != nil {
		return err
	}

	return nil
}

// GetSessions returns sidecar instance sessions of target app.
func (mgr *Manager) GetSessions(appid string) ([]*Session, error) {
	sessions := []*Session{}

	sCache, err := mgr.sessions.Get(appid)
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
func (mgr *Manager) DeleteSession(ins *SidecarInstance) error {
	if ins == nil {
		return errors.New("invalid sidecar instance: nil")
	}

	sCache, err := mgr.sessions.Get(ins.Appid)
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

	cache.Remove(mgr.sessionKey(ins.Dc, ins.IP))
	return nil
}
