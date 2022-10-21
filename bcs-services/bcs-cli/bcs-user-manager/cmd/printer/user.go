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
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintCreateSaasUserCmdResult prints the response that create saas user
func PrintCreateSaasUserCmdResult(flagOutput string, resp *pkg.CreateSaasUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create saas user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintCreatePlainUserCmdResult prints the response that create plain user
func PrintCreatePlainUserCmdResult(flagOutput string, resp *pkg.CreatePlainUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("create plain user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintGetAdminUserCmdResult prints the response that get admin user
func PrintGetAdminUserCmdResult(flagOutput string, resp *pkg.GetAdminUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get admin user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintGetSaasUserCmdResult prints the response that get saas user
func PrintGetSaasUserCmdResult(flagOutput string, resp *pkg.GetSaasUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get saas user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintGetPlainUserCmdResult prints the response that get plain user
func PrintGetPlainUserCmdResult(flagOutput string, resp *pkg.GetPlainUserResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("get plain user output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintRefreshSaasTokenCmdResult prints the response that refresh saas user token
func PrintRefreshSaasTokenCmdResult(flagOutput string, resp *pkg.RefreshSaasTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("refresh saas user token output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}

// PrintRefreshPlainTokenCmdResult prints the response that refresh plain user token
func PrintRefreshPlainTokenCmdResult(flagOutput string, resp *pkg.RefreshPlainTokenResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("refresh saas user token output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"ID", "NAME", "USER_TYPE", "USER_TOKEN", "CREATED_BY", "CREATED_AT", "UPDATED_AT", "EXPIRES_AT", "DELETED_AT",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			strconv.Itoa(int(data.ID)),
			data.Name,
			strconv.Itoa(int(data.UserType)),
			data.UserToken,
			data.CreatedBy,
			data.CreatedAt.Format(timeFormatter),
			data.UpdatedAt.Format(timeFormatter),
			data.ExpiresAt.Format(timeFormatter),
			data.DeletedAt.Format(timeFormatter),
		}
	}())
	tw.Render()
}
