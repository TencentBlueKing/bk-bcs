package web

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"

	"github.com/gin-gonic/gin"
)

// UserLoginRedirect 用户登入跳转URL
func (s *service) UserLoginRedirect(c *gin.Context) {
	values := url.Values{}
	values.Set("c_url", config.G.Web.Host+c.Request.URL.String())

	redirectUrl := fmt.Sprintf("%s?%s", config.G.BkLogin.Host, values.Encode())
	c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

// PermRequestRedirect 用户权限申请URL
func (s *service) UserPermRequestRedirect(c *gin.Context) {
	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")
	if projectId == "" {
		api.APIError(c, i18n.GetMessage("project_id is required"))
		return
	}

	redirectUrl, err := iam.MakeResourceApplyUrl(c.Request.Context(), projectId, clusterId, "")
	if err != nil {
		api.APIError(c, i18n.GetMessage(err.Error()))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
