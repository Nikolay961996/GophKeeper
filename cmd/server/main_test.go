package main

import (
	"testing"
)

func TestMainFunction(_ *testing.T) {
	_ = grpcAddr
	_ = jwtSecret
	_ = dbConnStr
	_ = usePostgres
}

func TestFlagParsing(t *testing.T) {
	if *grpcAddr == "" {
		t.Error("grpcAddr should not be empty")
	}
	if *jwtSecret == "" {
		t.Error("jwtSecret should not be empty")
	}
	if *dbConnStr == "" {
		t.Error("dbConnStr should not be empty")
	}
}
