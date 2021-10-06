package e2e

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
)

func TestInjectiveExchangeE2E(t *testing.T) {
	if !testing.Verbose() {
		// avoid errors from suites that would try to break things
		log.DefaultLogger.SetLevel(log.FatalLevel)
	} else {
		log.DefaultLogger.SetLevel(log.InfoLevel)
	}

	RegisterFailHandler(Fail)

	BeforeSuite(func() {
		checkMultinodeSetup()
	})

	RunSpecs(t, "Injective/Cosmos OCR module E2E Test Suite")
}

func checkMultinodeSetup() {
	clientCtx := getClientContext("user1")

	cc, err := chainclient.NewCosmosClient(clientCtx, daemonGRPCEndpoint)
	if err != nil {
		err = errors.Wrapf(err, "Couldn't connect to injectived. Make sure to run:\n\nCLEANUP=1 TEST_ERC20_DENOM=%s ./test/e2e_multinode.sh injectived\n\n", testERC20TokenDenom)
		orFail(err)
	}
	defer cc.Close()

	log.Infof("Connected to Injectived (GRPC) -> %s", daemonGRPCEndpoint)
	log.Infof("Connected to Injectived (Tendermint) -> %s", daemonTendermintEndpoint)

}
