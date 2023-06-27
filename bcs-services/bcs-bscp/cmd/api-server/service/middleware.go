package service

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/kit"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/rest"
)

// CheckDefaultTmplSpace create default template space if not existent
func (p *proxy) CheckDefaultTmplSpace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bizIDStr := chi.URLParam(r, "biz_id")
		bizIDInt, err := strconv.Atoi(bizIDStr)
		if err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		bizID := uint32(bizIDInt)
		if bizsOfTS.Has(bizID) {
			next.ServeHTTP(w, r)
			return
		}

		// use system user to create default template space
		kt := kit.MustGetKit(r.Context())
		kt.User = constant.BKSystemUser

		// create default template space when not existent
		in := &pbcs.CreateDefaultTemplateSpaceReq{BizId: bizID}
		if _, err := p.cfgClient.CreateDefaultTemplateSpace(kt.RpcCtx(), in); err != nil {
			render.Render(w, r, rest.BadRequest(err))
			return
		}
		bizsOfTS.Set(bizID)

		next.ServeHTTP(w, r)
	})
}
