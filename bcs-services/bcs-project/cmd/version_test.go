package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	// 输出到标准输出
	// ref: https://github.com/spf13/cobra/blob/e04ec725508c760e70263b031e5697c232d5c3fa/command_test.go#L34
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	// 执行version命令
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("exec version command failed, %v", err)
	}

	// 判断输出不为空
	assert.NotEmpty(t, buf.String())
}
