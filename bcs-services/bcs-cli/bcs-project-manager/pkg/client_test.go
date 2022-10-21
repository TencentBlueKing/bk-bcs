/**
 * @Author: Ambition
 * @Description:
 * @File: client_test
 * @Version: 1.0.0
 * @Date: 2022/10/17 11:48
 */

package pkg

import (
	"context"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func Test_GeProjectList(t *testing.T) {
	client, _, err := NewBcsProjectCli(context.Background(), &Config{
		APIServer: "127.0.0.1:8091",
		AuthToken: "",
	})
	if err != nil {
		log.Fatal(err)
	}
	rsp, err := client.ListProjects(context.Background(), &bcsproject.ListProjectsRequest{})
	assert.Nil(t, err)
	assert.NotNil(t, rsp)
}
