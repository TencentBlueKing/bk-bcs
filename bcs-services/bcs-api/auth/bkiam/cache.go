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

package bkiam

import (
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/config"

	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

const (
	SubRoot   = "/bcs/services/auth"
	TokenPath = SubRoot + "/token"
	KeyPath   = SubRoot + "/key"
)

var (
	tokenKeyAlreadyExists = fmt.Errorf("token key already exists")
)

func NewTokenCache(conf *config.ApiServConfig) (*TokenCache, error) {
	tokenCache := &TokenCache{
		conf: conf,

		cacheReal:  make(map[string]*auth.Token),
		cacheQueue: make(map[string]*auth.Token),

		zk: zkclient.NewZkClient(strings.Split(conf.BKIamAuth.BKIamZookeeper, ",")),
	}

	if err := tokenCache.zk.Connect(); err != nil {
		blog.Errorf("new token cache init zk client failed: %v", err)
		return nil, err
	}

	if err := tokenCache.zk.CheckMulNode(TokenPath, nil); err != nil {
		blog.Errorf("new token cache init tokenPath(%s) failed: %v", TokenPath, err)
		return nil, err
	}
	if err := tokenCache.zk.CheckMulNode(KeyPath, nil); err != nil {
		blog.Errorf("new token cache init keyPath(%s) failed: %v", KeyPath, err)
		return nil, err
	}

	go tokenCache.start()
	return tokenCache, nil
}

type TokenCache struct {
	conf *config.ApiServConfig

	zk       *zkclient.ZkClient
	authPath string

	// cacheReal use token as key
	cacheReal map[string]*auth.Token
	realLock  sync.RWMutex

	// cacheQueue use username as key
	cacheQueue map[string]*auth.Token
	queueLock  sync.RWMutex
}

func (tc *TokenCache) Update(token *auth.Token) {
	tc.queueLock.Lock()
	defer tc.queueLock.Unlock()

	token.Sign(0)
	tc.cacheQueue[token.Username] = token
}

func (tc *TokenCache) Get(tokenKey string) (*auth.Token, error) {
	tc.realLock.RLock()
	defer tc.realLock.RUnlock()

	token, ok := tc.cacheReal[tokenKey]
	if !ok {
		return token, fmt.Errorf("token not exists")
	}

	if token.ExpireTime.Before(time.Now()) {
		return token, fmt.Errorf("token expired")
	}

	return token, nil
}

func (tc *TokenCache) start() {
	blog.Infof("TokenCache sync start")
	ticker := time.NewTicker(time.Duration(tc.conf.BKIamAuth.AuthTokenSyncTime) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go tc.syncFrom()
			go tc.syncTo()
		}
	}
}

func (tc *TokenCache) syncFrom() {
	usernameList, err := tc.fetchUsernameList()
	if err != nil {
		blog.Errorf("syncFrom fetch username list failed: %v", err)
		return
	}

	cacheReal := make(map[string]*auth.Token)
	for _, username := range usernameList {
		token, err := tc.fetchToken(tc.getTokenPath(username))
		if err != nil {
			blog.Errorf("syncFrom fetch token(user: %s) failed: %v", username, err)
			continue
		}

		cacheReal[token.Token] = token
	}

	tc.realLock.Lock()
	tc.cacheReal = cacheReal
	tc.realLock.Unlock()
}

func (tc *TokenCache) syncTo() {
	tc.queueLock.RLock()

	var wg sync.WaitGroup
	for _, token := range tc.cacheQueue {
		wg.Add(1)
		go tc.singleSyncTo(token, &wg)
	}
	wg.Wait()

	tc.queueLock.RUnlock()
	tc.queueLock.Lock()
	tc.cacheQueue = make(map[string]*auth.Token)
	tc.queueLock.Unlock()
}

func (tc *TokenCache) singleSyncTo(token *auth.Token, wg *sync.WaitGroup) {
	defer wg.Done()

	oldToken, err := tc.fetchToken(tc.getTokenPath(token.Username))
	if err != nil && err != zk.ErrNoNode {
		blog.Errorf("singleSyncTo fetch token(user: %s) failed: %v", token.Username, err)
		return
	}

	// user already exist, then update the old token
	if err == nil {
		oldToken.ExpireTime = token.ExpireTime
		oldToken.Message = token.Message
		if err := tc.updateToken(oldToken); err != nil {
			blog.Errorf("singleSyncTo update token(token: %s, user: %s) failed: %v", oldToken.Token, oldToken.Username, err)
			return
		}

		blog.V(3).Infof("singleSyncTo success to update token(token: %s, user: %s)", oldToken.Token, oldToken.Username)
		return
	}

	// user not exist, then create a new token
	for i := 0; i < 3; i++ {
		token.Generate()
		if err := tc.createToken(token); err == tokenKeyAlreadyExists {
			continue
		} else if err != nil {
			blog.Errorf("singleSyncTo create token(token: %s, user: %s) failed: %v", oldToken.Token, oldToken.Username, err)
			return
		}

		blog.V(3).Infof("singleSyncTo success to create token(token: %s, user: %s)", oldToken.Token, oldToken.Username)
		return
	}

	blog.Errorf("singleSyncTo create tokenKey duplicated for 3 times, will not try again")
}

func (tc *TokenCache) fetchUsernameList() ([]string, error) {
	return tc.zk.GetChildren(tc.getTokenPath(""))
}

func (tc *TokenCache) fetchToken(tokenPath string) (*auth.Token, error) {
	raw, _, _, err := tc.zk.GetW(tokenPath)
	if err != nil {
		return nil, err
	}

	var token auth.Token
	err = codec.DecJson(raw, &token)
	return &token, err
}

func (tc *TokenCache) updateToken(token *auth.Token) error {
	token.UpdateTime = time.Now()

	var data []byte
	if err := codec.EncJson(token, &data); err != nil {
		blog.Errorf("updateToken encode token(token: %s, user: %s) failed: %v", token.Token, token.Username, err)
		return err
	}

	if err := tc.zk.Update(tc.getTokenPath(token.Username), string(data)); err != nil {
		blog.Errorf("updateToken update token(token: %s, user: %s) failed: %v", token.Token, token.Username, err)
		return err
	}

	return nil
}

func (tc *TokenCache) createToken(token *auth.Token) error {
	token.CreateTime = time.Now()
	token.UpdateTime = token.CreateTime

	if ok, err := tc.occupyKey(token); err != nil {
		blog.Errorf("createToken occupy key(token: %s, user: %s) failed: %v", token.Token, token.Username, err)
		return err
	} else if !ok {
		blog.Errorf("createToken occupy key(token: %s, user: %s) failed: %v", token.Token, token.Username, tokenKeyAlreadyExists)
		return tokenKeyAlreadyExists
	}

	var data []byte
	if err := codec.EncJson(token, &data); err != nil {
		blog.Errorf("createToken encode token(token: %s, user: %s) failed: %v", token.Token, token.Username, err)
		return err
	}

	if err := tc.zk.Create(tc.getTokenPath(token.Username), data); err != nil {
		blog.Errorf("createToken create token(token: %s, user: %s) failed: %v", token.Token, token.Username, err)
		return err
	}

	return nil
}

func (tc *TokenCache) occupyKey(token *auth.Token) (bool, error) {
	err := tc.zk.Create(tc.getKeyPath(token.Token), []byte(token.Token))
	if err == nil || err != zk.ErrNodeExists {
		return true, nil
	}

	return false, err
}

func (tc *TokenCache) releaseKey(token *auth.Token) error {
	err := tc.zk.Del(tc.getKeyPath(token.Token), -1)
	if err == zk.ErrNoNode {
		err = nil
	}

	return err
}

func (tc *TokenCache) getTokenPath(tokenUsername string) string {
	return path.Join(tc.authPath, TokenPath, tokenUsername)
}

func (tc *TokenCache) getKeyPath(tokenKey string) string {
	return path.Join(tc.authPath, KeyPath, tokenKey)
}
