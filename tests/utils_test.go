package tests

import (
	"github.com/goal-web/goal-cli/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTitle(t *testing.T) {
	tests := map[string]string{
		"user":      "User",
		"user_post": "User_post",
	}
	for key, value := range tests {
		assert.True(t, utils.Capitalize(key) == value)

	}
}
