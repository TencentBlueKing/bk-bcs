/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

package iamv4

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	restful "github.com/emicklei/go-restful/v3"
)

type fakeQuerier struct {
	list  func(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error)
	fetch func(ctx context.Context, resType string, filter Filter) ([]Instance, error)
}

func (f fakeQuerier) List(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error) {
	return f.list(ctx, tenantID, resType, filter, page)
}

func (f fakeQuerier) Fetch(ctx context.Context, resType string, filter Filter) ([]Instance, error) {
	return f.fetch(ctx, resType, filter)
}

func serveIAMV4(t *testing.T, h *Handler, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/iamv4-provider/resources", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	httpReq := restful.NewRequest(req)
	httpResp := restful.NewResponse(rec)
	h.Serve(httpReq, httpResp)
	return rec
}

func TestHandlerListInstanceEnvelope(t *testing.T) {
	h := &Handler{Querier: fakeQuerier{
		list: func(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error) {
			if resType != Project || page.Page != 1 || page.PageSize != 20 {
				t.Fatalf("unexpected list args type=%s page=%v", resType, page)
			}
			if filter.Keyword != "demo" {
				t.Fatalf("keyword=%s", filter.Keyword)
			}
			return 2, []Instance{{ID: "p1", DisplayName: "demo"}}, nil
		},
	}}
	rec := serveIAMV4(t, h, `{
		"type":"project","method":"list_instance",
		"filter":{"keyword":"demo"},
		"page":{"page":1,"page_size":20}
	}`, map[string]string{"X-Request-Id": "rid-1"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-Id") != "rid-1" {
		t.Fatalf("missing request id")
	}
	var resp SuccessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(resp.Data)
	if !bytes.Contains(raw, []byte(`"count":2`)) || bytes.Contains(rec.Body.Bytes(), []byte(`"code"`)) {
		t.Fatalf("unexpected envelope %s", rec.Body.String())
	}
}

func TestHandlerFetchRequires(t *testing.T) {
	h := &Handler{Querier: fakeQuerier{
		fetch: func(ctx context.Context, resType string, filter Filter) ([]Instance, error) {
			return []Instance{{
				ID: "c1", DisplayName: "cls", IAMPath: "/project,p1/", Approvers: []string{"u"},
			}}, nil
		},
	}}
	rec := serveIAMV4(t, h, `{
		"type":"cluster","method":"fetch_instance_info",
		"filter":{"ids":["c1"]},
		"requires":["display_name"]
	}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("_bk_iam_path_")) {
		t.Fatalf("requires should drop path: %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"display_name"`)) ||
		!bytes.Contains(rec.Body.Bytes(), []byte(`"cls"`)) {
		t.Fatalf("missing display_name: %s", rec.Body.String())
	}
}

func TestHandlerValidationErrors(t *testing.T) {
	h := &Handler{Querier: fakeQuerier{}}
	cases := []struct {
		name string
		body string
	}{
		{"bad type", `{"type":"foo","method":"list_instance"}`},
		{"bad method", `{"type":"project","method":"search_instance"}`},
		{"page size", `{"type":"project","method":"list_instance","page":{"page_size":2000}}`},
		{"empty ids", `{"type":"project","method":"fetch_instance_info","filter":{"ids":[]}}`},
		{"keyword", `{"type":"project","method":"list_instance","filter":{"keyword":"` + strings.Repeat("k", 200) + `"}}`},
	}
	for _, c := range cases {
		rec := serveIAMV4(t, h, c.body, nil)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("%s status=%d body=%s", c.name, rec.Code, rec.Body.String())
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(`"error"`)) {
			t.Fatalf("%s missing error envelope: %s", c.name, rec.Body.String())
		}
	}
}

func TestHandlerQuerierInvalidParent(t *testing.T) {
	h := &Handler{Querier: fakeQuerier{
		list: func(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error) {
			return 0, nil, invalidRequest("parent is required")
		},
	}}
	rec := serveIAMV4(t, h, `{"type":"cluster","method":"list_instance"}`, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandlerInternalErrorHidden(t *testing.T) {
	h := &Handler{Querier: fakeQuerier{
		list: func(ctx context.Context, tenantID, resType string, filter Filter, page Page) (int, []Instance, error) {
			return 0, nil, context.DeadlineExceeded
		},
	}}
	rec := serveIAMV4(t, h, `{"type":"project","method":"list_instance"}`, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte("Deadline")) {
		t.Fatalf("leaked internal error: %s", rec.Body.String())
	}
}
