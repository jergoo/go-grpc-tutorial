package main

import (
	"testing"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ping"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Ping()
		})
	}
}
