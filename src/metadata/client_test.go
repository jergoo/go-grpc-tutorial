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

func TestMultiPong(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "multi pong"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			MultiPong()
		})
	}
}
