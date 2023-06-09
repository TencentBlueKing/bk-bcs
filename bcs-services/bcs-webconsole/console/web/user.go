package web

import (
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
)

// UserPermRequestRedirect 用户权限申请URL
func (s *service) UserPermRequestRedirect(c *gin.Context) {
	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")
	if projectId == "" {
		api.APIError(c, i18n.GetMessage("project_id is required"))
		return
	}

	redirectUrl, err := iam.MakeResourceApplyUrl(c.Request.Context(), projectId, clusterId, route.GetNamespace(c), "")
	if err != nil {
		api.APIError(c, i18n.GetMessage(err.Error()))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
