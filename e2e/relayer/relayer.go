package relayer

import (
	"fmt"
	"testing"

	dockerclient "github.com/docker/docker/client"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"go.uber.org/zap"
)

const (
	Rly    = "rly"
	Hermes = "hermes"

	cosmosRelayerRepository = "ghcr.io/cosmos/relayer"
	cosmosRelayerUser       = "100:1000" // docker run -it --rm --entrypoint echo ghcr.io/cosmos/relayer "$(id -u):$(id -g)"
)

// Config holds configuration values for the relayer used in the tests.
type Config struct {
	// Tag is the tag used for the relayer image.
	Tag string
	// Type specifies the type of relayer that this is.
	Type string
}

// New returns an implementation of ibc.Relayer depending on the provided RelayerType.
func New(t *testing.T, cfg Config, logger *zap.Logger, dockerClient *dockerclient.Client, network string) ibc.Relayer {
	switch cfg.Type {
	case Rly:
		return newCosmosRelayer(t, cfg.Tag, logger, dockerClient, network)
	case Hermes:
		return newHermesRelayer()
	default:
		panic(fmt.Sprintf("unknown relayer specified: %s", cfg.Type))
	}
}

// newCosmosRelayer returns an instance of the go relayer.
// Options are used to allow for relayer version selection and specifying the default processing option.
func newCosmosRelayer(t *testing.T, tag string, logger *zap.Logger, dockerClient *dockerclient.Client, network string) ibc.Relayer {
	customImageOption := relayer.CustomDockerImage(cosmosRelayerRepository, tag, cosmosRelayerUser)
	relayerProcessingOption := relayer.StartupFlags("-p", "events") // relayer processes via events

	relayerFactory := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, logger, customImageOption, relayerProcessingOption)

	return relayerFactory.Build(
		t, dockerClient, network,
	)
}

// newHermesRelayer returns an instance of the hermes relayer.
func newHermesRelayer() ibc.Relayer {
	panic("hermes relayer not yet implemented for interchaintest")
}
