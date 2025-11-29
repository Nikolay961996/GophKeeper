package main

import (
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/stretchr/testify/assert"
)

func TestGetDataManager_ErrorCases(t *testing.T) {
	// No token
	cfg := &config.Config{Token: ""}
	manager, err := getDataManager(false, cfg)
	assert.Error(t, err)
	assert.Nil(t, manager)

	// No master password
	cfg.Token = "test-token"
	manager, err = getDataManager(true, cfg)
	assert.Error(t, err)
	assert.Nil(t, manager)
}
