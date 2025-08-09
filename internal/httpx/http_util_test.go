package httpx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUrlContent(t *testing.T) {
	url := "https://raw.githubusercontent.com/Guovin/iptv-api/gd/output/result.m3u"

	content, err := LoadUrlContent(context.Background(), url)

	require.NoError(t, err, "LoadUrlContent failed")
	require.NotEmpty(t, content)
}

func TestPingURL(t *testing.T) {
	url := "https://www.github.com"
	lagency, err := PingURL(url)
	require.NoError(t, err, "PingURL failed")
	require.NotZero(t, lagency)
}

func TestTestDownloadSpeed(t *testing.T) {
	url := "https://download.jetbrains.com/go/goland-2025.1.3-aarch64.dmg"
	kbps, err := TestDownloadSpeed(context.Background(), url)
	require.NoError(t, err, "TestDownloadSpeed failed")
	require.NotZero(t, kbps)
}
