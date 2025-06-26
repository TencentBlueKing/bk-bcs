package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateResourceLimit(t *testing.T) {
	t.Run("both empty", func(t *testing.T) {
		err := ValidateResourceLimit("", "")
		assert.NoError(t, err)
	})

	t.Run("only request", func(t *testing.T) {
		err := ValidateResourceLimit("100m", "")
		assert.NoError(t, err)
	})

	t.Run("only limit", func(t *testing.T) {
		err := ValidateResourceLimit("", "200m")
		assert.NoError(t, err)
	})

	t.Run("limit >= request", func(t *testing.T) {
		err := ValidateResourceLimit("100m", "200m")
		assert.NoError(t, err)
	})

	t.Run("limit == request", func(t *testing.T) {
		err := ValidateResourceLimit("100m", "100m")
		assert.NoError(t, err)
	})

	t.Run("limit < request", func(t *testing.T) {
		err := ValidateResourceLimit("200m", "100m")
		assert.Error(t, err)
	})

	t.Run("invalid request", func(t *testing.T) {
		err := ValidateResourceLimit("bad", "100m")
		assert.Error(t, err)
	})

	t.Run("invalid limit", func(t *testing.T) {
		err := ValidateResourceLimit("100m", "bad")
		assert.Error(t, err)
	})
}
