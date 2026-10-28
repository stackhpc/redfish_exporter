package collector

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/prometheus/common/promslog"
	"github.com/stackhpc/redfish_exporter/common"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateStaticBMC(t *testing.T, path string, collect_logs bool) *RedfishCollector {
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
	endpoint, err := container.Endpoint(ctx, "")
	require.NoError(t, err)

	promslogConfig := promslog.Config{}

	targetLoggerCtx := promslog.New(&promslogConfig)

	hostConfig := common.HostConfig{
		Collectlogs: collect_logs,
	}
	require.NoError(t, err)

	collectionCtx, err := common.NewCollectionContext(endpoint, &hostConfig, targetLoggerCtx, collect_logs)
	require.NoError(t, err)
	collector := NewRedfishCollector(collectionCtx, targetLoggerCtx)
	return collector
}
