package api

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

/*
	资源
*/

// ListResourceByLocation 从区域获取资源
func (aks *AksServiceImpl) ListResourceByLocation(ctx context.Context, location string) ([]*armcompute.ResourceSKU,
	error) {
	resp := make([]*armcompute.ResourceSKU, 0)
	pager := aks.resourceClient.NewListPager(&armcompute.ResourceSKUsClientListOptions{
		Filter: to.Ptr(fmt.Sprintf("location eq '%s'", location)),
	})
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		resp = append(resp, nextResult.Value...)
	}
	return resp, nil
}
