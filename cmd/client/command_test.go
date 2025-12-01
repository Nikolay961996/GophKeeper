package main

import (
	"gophkeeper/internal/client/cobra"
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/stretchr/testify/assert"
)

func TestGetDataManager_ErrorCases(t *testing.T) {
	// No token
	cfg := &config.Config{Token: ""}
	manager, err := cobrakeeper.GetDataManager(false, cfg, masterPassword)
	assert.Error(t, err)
	assert.Nil(t, manager)

	// No master password
	cfg.Token = "test-token"
	manager, err = cobrakeeper.GetDataManager(true, cfg, masterPassword)
	assert.Error(t, err)
	assert.Nil(t, manager)
}
