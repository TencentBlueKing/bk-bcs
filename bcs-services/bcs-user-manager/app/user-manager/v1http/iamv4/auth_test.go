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

package iamv4

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
)

func serveAuth(t *testing.T, fn func(ctx context.Context) (string, error), setAuth func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	orig := FetchToken
	FetchToken = fn
	resetTokenCache()
	t.Cleanup(func() {
		resetTokenCache()
		FetchToken = orig
	})

	ws := new(restful.WebService)
	ws.Route(AuthFunc(ws.POST("/v1/iamv4-provider/resources")).To(func(*restful.Request, *restful.Response) {}))
	container := restful.NewContainer()
	container.Add(ws)

	req := httptest.NewRequest(http.MethodPost, "/v1/iamv4-provider/resources", nil)
	if setAuth != nil {
		setAuth(req)
	}
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)
	return rec
}

func TestAuthenticateMissingBasic(t *testing.T) {
	rec := serveAuth(t, func(context.Context) (string, error) { return "tok", nil }, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorEnvelope(t, rec)
}

func TestAuthenticateWrongUser(t *testing.T) {
	rec := serveAuth(t, func(context.Context) (string, error) { return "tok", nil }, func(r *http.Request) {
		r.SetBasicAuth("other", "tok")
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestAuthenticateWrongToken(t *testing.T) {
	rec := serveAuth(t, func(context.Context) (string, error) { return "tok", nil }, func(r *http.Request) {
		r.SetBasicAuth(basicUser, "bad")
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`"tok"`)) || bytes.Contains(rec.Body.Bytes(), []byte("bad")) {
		t.Fatalf("token leaked: %s", rec.Body.String())
	}
}

func TestAuthenticateNotConfigured(t *testing.T) {
	rec := serveAuth(t, func(context.Context) (string, error) { return "", ErrNotConfigured }, func(r *http.Request) {
		r.SetBasicAuth(basicUser, "tok")
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticateOK(t *testing.T) {
	rec := serveAuth(t, func(context.Context) (string, error) { return "tok", nil }, func(r *http.Request) {
		r.SetBasicAuth(basicUser, "tok")
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticateCachesToken(t *testing.T) {
	calls := 0
	fn := func(context.Context) (string, error) {
		calls++
		return "tok", nil
	}
	rec := serveAuth(t, fn, func(r *http.Request) { r.SetBasicAuth(basicUser, "tok") })
	if rec.Code != http.StatusOK {
		t.Fatalf("first status=%d", rec.Code)
	}
	rec = serveAuthReuse(t, fn, func(r *http.Request) { r.SetBasicAuth(basicUser, "tok") })
	if rec.Code != http.StatusOK {
		t.Fatalf("second status=%d", rec.Code)
	}
	if calls != 1 {
		t.Fatalf("FetchToken calls=%d, want 1", calls)
	}
}

func TestAuthenticateDoesNotCacheFailure(t *testing.T) {
	calls := 0
	fn := func(context.Context) (string, error) {
		calls++
		if calls == 1 {
			return "", errors.New("upstream down")
		}
		return "tok", nil
	}
	rec := serveAuth(t, fn, func(r *http.Request) { r.SetBasicAuth(basicUser, "tok") })
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("first status=%d", rec.Code)
	}
	rec = serveAuthReuse(t, fn, func(r *http.Request) { r.SetBasicAuth(basicUser, "tok") })
	if rec.Code != http.StatusOK {
		t.Fatalf("retry status=%d body=%s", rec.Code, rec.Body.String())
	}
	if calls != 2 {
		t.Fatalf("FetchToken calls=%d, want 2", calls)
	}
}

func serveAuthReuse(t *testing.T, fn func(ctx context.Context) (string, error), setAuth func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	ws := new(restful.WebService)
	ws.Route(AuthFunc(ws.POST("/v1/iamv4-provider/resources")).To(func(*restful.Request, *restful.Response) {}))
	container := restful.NewContainer()
	container.Add(ws)
	req := httptest.NewRequest(http.MethodPost, "/v1/iamv4-provider/resources", nil)
	if setAuth != nil {
		setAuth(req)
	}
	rec := httptest.NewRecorder()
	container.ServeHTTP(rec, req)
	return rec
}

func assertErrorEnvelope(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var resp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v body=%s", err, rec.Body.String())
	}
	if resp.Error.Code == "" {
		t.Fatalf("missing error code: %s", rec.Body.String())
	}
}
