package permission

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/cmd/utils"
	userV1 "github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/usermanager/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http"
)

func getPermission(c *utils.ClientContext) error {
	if err := c.MustSpecified(utils.OptionUserName, utils.OptionResourceType); err != nil {
		return err
	}

	userManager := userV1.NewBcsUserManager(utils.GetClientOption())
	pf := v1http.GetPermissionForm{
		UserName:     c.String(utils.OptionUserName),
		ResourceType: c.String(utils.OptionResourceType),
	}
	data, err := json.Marshal(pf)
	if err != nil {
		return err
	}
	permissions, err := userManager.GetPermission(http.MethodGet, data)
	if err != nil {
		return fmt.Errorf("failed to grant permission: %v", err)
	}

	return printGet(permissions)
}

func printGet(single interface{}) error {
	fmt.Printf("%s\n", utils.TryIndent(single))
	return nil
}
