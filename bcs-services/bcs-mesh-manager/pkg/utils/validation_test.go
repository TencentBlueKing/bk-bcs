package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateResourceLimit(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		err := validateResourceLimit("", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be empty")
	})

	t.Run("only request", func(t *testing.T) {
		err := validateResourceLimit("100m", "")
		assert.NoError(t, err)
	})

	t.Run("only limit", func(t *testing.T) {
		err := validateResourceLimit("", "200m")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be empty")
	})

	t.Run("limit >= request", func(t *testing.T) {
		err := validateResourceLimit("100m", "200m")
		assert.NoError(t, err)
	})

	t.Run("limit == request", func(t *testing.T) {
		err := validateResourceLimit("100m", "100m")
		assert.NoError(t, err)
	})

	t.Run("limit < request", func(t *testing.T) {
		err := validateResourceLimit("200m", "100m")
		assert.Error(t, err)
	})

	t.Run("invalid request", func(t *testing.T) {
		err := validateResourceLimit("bad", "100m")
		assert.Error(t, err)
	})

	t.Run("invalid limit", func(t *testing.T) {
		err := validateResourceLimit("100m", "bad")
		assert.Error(t, err)
	})

	// 测试当limit为0时，不进行资源限制验证
	t.Run("limit is zero - should not validate", func(t *testing.T) {
		err := validateResourceLimit("100m", "0")
		assert.NoError(t, err)
	})

	t.Run("limit is zero with empty request", func(t *testing.T) {
		err := validateResourceLimit("", "0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request cannot be empty")
	})

	t.Run("limit is zero with large request", func(t *testing.T) {
		err := validateResourceLimit("1000m", "0")
		assert.NoError(t, err)
	})

	// 测试带单位的0值
	t.Run("limit is zero with unit - should not validate", func(t *testing.T) {
		err := validateResourceLimit("100m", "0m")
		assert.NoError(t, err)
	})

	t.Run("limit is zero with memory unit - should not validate", func(t *testing.T) {
		err := validateResourceLimit("512Mi", "0Mi")
		assert.NoError(t, err)
	})

	t.Run("limit is zero with different memory unit - should not validate", func(t *testing.T) {
		err := validateResourceLimit("1Gi", "0Gi")
		assert.NoError(t, err)
	})

	// 测试正常的带单位值
	t.Run("limit >= request with units", func(t *testing.T) {
		err := validateResourceLimit("100m", "200m")
		assert.NoError(t, err)
	})

	t.Run("limit >= request with memory units", func(t *testing.T) {
		err := validateResourceLimit("512Mi", "1Gi")
		assert.NoError(t, err)
	})

	t.Run("limit < request with units", func(t *testing.T) {
		err := validateResourceLimit("200m", "100m")
		assert.Error(t, err)
	})

	t.Run("limit < request with memory units", func(t *testing.T) {
		err := validateResourceLimit("1Gi", "512Mi")
		assert.Error(t, err)
	})
}
