package printer

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
	"github.com/olekukonko/tablewriter"
	"k8s.io/klog"
	"os"
	"strconv"
)

// PrintApplyTkeCidrCmdResult prints the response that apply tke cidrs
func PrintApplyTkeCidrCmdResult(flagOutput string, resp *pkg.ApplyTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("apply tke cidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.SetHeader(func() []string {
		return []string{
			"VPC", "CIDR", "IP_NUMBER", "STATUS",
		}
	}())
	tw.SetAutoMergeCells(true)
	data := resp.Data
	tw.Append(func() []string {
		return []string{
			data.Vpc,
			data.Cidr,
			strconv.Itoa(int(data.IpNumber)),
			data.Status,
		}
	}())
	tw.Render()
}

// PrintAddTkeCidrCmdResult prints the response that add tkecidrs
func PrintAddTkeCidrCmdResult(flagOutput string, resp *pkg.AddTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("add tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.Render()
}

// PrintListTkeCidrCmdResult prints the response that list TkeCidr
func PrintListTkeCidrCmdResult(flagOutput string, resp *pkg.ListTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("list TkeCidr output json to stdout failed: %s", err.Error())
		}
	}

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetHeader(func() []string {
		return []string{
			"COUNT", "VPC", "IP_NUMBER", "STATUS",
		}
	}())
	// 添加页脚
	tw.SetFooter([]string{"", "Total", strconv.Itoa(len(resp.Data))})
	// 合并相同值的列
	//tw.SetAutoMergeCells(true)
	for _, item := range resp.Data {
		tw.Append(func() []string {
			return []string{
				strconv.Itoa(item.Count),
				item.Vpc,
				strconv.Itoa(int(item.IpNumber)),
				item.Status,
			}
		}())
	}
	tw.Render()
}

// PrintReleaseTkeCidrCmdResult prints the response that release tkecidrs
func PrintReleaseTkeCidrCmdResult(flagOutput string, resp *pkg.ReleaseTkeCidrResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("release tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.Render()
}

// PrintSyncTkeClusterCredentialsCmdResult prints the response that sync cluster tkecidrs
func PrintSyncTkeClusterCredentialsCmdResult(flagOutput string, resp *pkg.SyncTkeClusterCredentialsResponse) {
	if flagOutput == outputTypeJSON {
		if err := encodeJSON(resp); err != nil {
			klog.Fatalf("sync cluster tkecidrs output json to stdout failed: %s", err.Error())
		}
	}
	tw := defaultTableWriter()
	tw.Render()
}
