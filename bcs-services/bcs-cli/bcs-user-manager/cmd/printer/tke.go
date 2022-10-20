package printer

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-user-manager/pkg"
	"k8s.io/klog"
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
			strconv.FormatUint(uint64(data.IpNumber), 10),
			data.Status,
		}
	}())
	tw.Render()
}
