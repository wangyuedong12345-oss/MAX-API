package constant

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPath2RelayModeAlphaSearch(t *testing.T) {
	tests := []string{
		"/v1/alpha/search",
		"/v1/alpha/search?foo=1",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			require.Equal(t, RelayModeAlphaSearch, Path2RelayMode(path))
		})
	}
}

func TestPath2RelayModeOpenAIVideoSubmit(t *testing.T) {
	tests := []string{
		"/v1/videos",
		"/v1/videos/generations",
		"/v1/video/generations",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			require.Equal(t, RelayModeVideoSubmit, Path2RelayMode(path))
		})
	}
}

func TestPath2RelayModeOpenAIVideoFetchRemainsMiddlewareDriven(t *testing.T) {
	require.Equal(t, RelayModeUnknown, Path2RelayMode("/v1/videos/task_123"))
}
