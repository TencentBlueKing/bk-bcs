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

package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/container/v1"
	"k8s.io/client-go/rest"
)

const (
	// GoogleAuthPlugin is different with "gcp" which is already in client-go tree. It gets GCP service account secret
	// key from config and generate a google token source which is used to authenticate with google cloud.
	GoogleAuthPlugin = "google"
)

func init() {
	if err := rest.RegisterAuthProviderPlugin(GoogleAuthPlugin, newGoogleAuthProvider); err != nil {
		blog.Errorf("RegisterAuthProviderPlugin failed to register %s auth plugin: %v", GoogleAuthPlugin, err)
	}
}

var _ rest.AuthProvider = &googleAuthProvider{}

type googleAuthProvider struct {
	tokenSource oauth2.TokenSource
	persister   rest.AuthProviderConfigPersister
}

func (g *googleAuthProvider) WrapTransport(rt http.RoundTripper) http.RoundTripper {
	return &oauth2.Transport{
		Base:   rt,
		Source: g.tokenSource,
	}
}
func (g *googleAuthProvider) Login() error { return nil }

func newGoogleAuthProvider(_ string, config map[string]string, persister rest.AuthProviderConfigPersister) (rest.AuthProvider, error) {
	ts, err := google.CredentialsFromJSON(context.TODO(), []byte(config["credentials"]), container.CloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("newGoogleAuthProvider failed to create google token source: %+v", err)
	}
	cts, err := newCachedTokenSource(config["access-token"], config["expiry"], persister, ts.TokenSource, config)
	if err != nil {
		return nil, err
	}
	return &googleAuthProvider{tokenSource: cts, persister: persister}, nil
}

type cachedTokenSource struct {
	lk          sync.Mutex
	source      oauth2.TokenSource
	accessToken string
	expiry      time.Time
	persister   rest.AuthProviderConfigPersister
	cache       map[string]string
}

func newCachedTokenSource(accessToken, expiry string, persister rest.AuthProviderConfigPersister, ts oauth2.TokenSource, cache map[string]string) (*cachedTokenSource, error) {
	var expiryTime time.Time
	if parsedTime, err := time.Parse(time.RFC3339Nano, expiry); err == nil {
		expiryTime = parsedTime
	}
	if cache == nil {
		cache = make(map[string]string)
	}
	return &cachedTokenSource{
		source:      ts,
		accessToken: accessToken,
		expiry:      expiryTime,
		persister:   persister,
		cache:       cache,
	}, nil
}

func (t *cachedTokenSource) Token() (*oauth2.Token, error) {
	tok := t.cachedToken()
	if tok.Valid() && !tok.Expiry.IsZero() {
		return tok, nil
	}
	tok, err := t.source.Token()
	if err != nil {
		return nil, err
	}
	cache := t.update(tok)
	if t.persister != nil {
		if err := t.persister.Persist(cache); err != nil {
			blog.Errorf("Persist failed to persist token: %v", err)
		}
	}
	return tok, nil
}

func (t *cachedTokenSource) cachedToken() *oauth2.Token {
	t.lk.Lock()
	defer t.lk.Unlock()
	return &oauth2.Token{
		AccessToken: t.accessToken,
		TokenType:   "Bearer",
		Expiry:      t.expiry,
	}
}

func (t *cachedTokenSource) update(tok *oauth2.Token) map[string]string {
	t.lk.Lock()
	defer t.lk.Unlock()
	t.accessToken = tok.AccessToken
	t.expiry = tok.Expiry
	ret := map[string]string{}
	for k, v := range t.cache {
		ret[k] = v
	}
	ret["access-token"] = t.accessToken
	ret["expiry"] = t.expiry.Format(time.RFC3339Nano)
	return ret
}

// baseCache is the base configuration value for this TokenSource, without any cached ephemeral tokens.
func (t *cachedTokenSource) baseCache() map[string]string {
	t.lk.Lock()
	defer t.lk.Unlock()
	ret := map[string]string{}
	for k, v := range t.cache {
		ret[k] = v
	}
	delete(ret, "access-token")
	delete(ret, "expiry")
	return ret
}
