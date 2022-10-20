package printer

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
	"k8s.io/klog"
	"strconv"
)

// PrintCreateAdminUserCmdResult prints the response that create admin user
func PrintCreateAdminUserCmdResult(flagOutput string, resp *pkg.CreateAdminUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create admin user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "Name", "UserType", "UserToken", "CreatedBy", "CreatedAt", "UpdatedAt", "ExpiresAt", "DeletedAt",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.FormatUint(uint64(data.ID), 10),
			data.Name,
			strconv.FormatUint(uint64(data.UserType), 10),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.String(),
			data.UpdatedAt.String(),
			data.ExpiresAt.String(),
			data.DeletedAt.String(),
		}
	}())
	tw.Render()
}
