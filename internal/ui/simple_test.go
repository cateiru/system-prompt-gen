package ui

import (
	"testing"

	"github.com/cateiru/system-prompt-gen/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSimpleModel(t *testing.T) {
	cfg := config.DefaultConfig()
	m := initialModel(cfg)

	assert.NotNil(t, m)
	assert.Equal(t, stateLoading, m.state)
}
