package qcloud

import (
	"fmt"
	"testing"
)

func TestGetQcloudNodeInfo(t *testing.T) {
	result, err := GetQcloudNodeMetadata()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(result)

}
