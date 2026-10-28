package test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stmcginnis/gofish"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateRedfishClient(t *testing.T, path string) *gofish.APIClient {
	ctx := context.Background()
	container, err := testcontainers.Run(
		ctx, "",
		testcontainers.WithExposedPorts("443/tcp"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("443/tcp"),
		),
		testcontainers.WithDockerfile(testcontainers.FromDockerfile{
			Context:   os.Getenv("FIXTURES_DIR"),
			KeepImage: true,
		}),
		testcontainers.WithCmd("sushy-static", "-p", "443", "-c", "sushy.crt", "-k", "sushy.key", "-m", fmt.Sprintf("mockups/%s", path)),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	// Get host address and port of exposed container port
	endpoint, err := container.Endpoint(ctx, "https")
	require.NoError(t, err)

	config := gofish.ClientConfig{
		Endpoint: endpoint,
		Username: "",
		Password: "",
		Insecure: true,
	}

	redfishClient, err := gofish.Connect(config)
	require.NoError(t, err)

	return redfishClient
}

func CollectAndCompare(t *testing.T, c prometheus.Collector, expected string, metricNames ...string) {
	t.Helper()

	actualBytes, err := testutil.CollectAndFormat(c, expfmt.TypeTextPlain, metricNames...)
	if err != nil {
		t.Fatalf("failed to collect and format: %v", err)
	}

	actualLines := extractLines(string(actualBytes))
	expectedLines := extractLines(expected)

	// Sort both slices so the order is deterministic
	sort.Strings(actualLines)
	sort.Strings(expectedLines)

	// Join them back into multi-line strings for testify's text differ
	actualText := strings.Join(actualLines, "\n") + "\n"
	expectedText := strings.Join(expectedLines, "\n") + "\n"

	assert.Equal(t, expectedText, actualText, "Metrics mismatch")

	for _, name := range metricNames {
		assert.Contains(t, actualText, name)
	}
}

// Returns a slice of strings with headers and empty lines removed
func extractLines(s string) []string {
	var result []string
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			result = append(result, trimmed)
		}
	}
	return result
}
