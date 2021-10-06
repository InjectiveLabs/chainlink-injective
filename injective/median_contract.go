package injective

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/ocr2/reportingplugin/median"
)

var _ median.MedianContract = &CosmosMedianReporter{}

type CosmosMedianReporter struct {
	FeedId      string
	QueryClient chaintypes.QueryClient
}

func (c *CosmosMedianReporter) LatestTransmissionDetails(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	latestAnswer *big.Int,
	latestTimestamp time.Time,
	err error,
) {
	if len(c.FeedId) == 0 {
		err = errors.New("CosmosMedianReporter has no FeedId set")
		return
	}

	if c.QueryClient == nil {
		err = errors.New("cannot query LatestTransmissionDetails: no QueryClient set")
		return
	}

	var resp *chaintypes.QueryLatestTransmissionDetailsResponse
	if resp, err = c.QueryClient.LatestTransmissionDetails(ctx, &chaintypes.QueryLatestTransmissionDetailsRequest{
		FeedId: c.FeedId,
	}); err != nil {
		return
	}

	configDigest = configDigestFromBytes(resp.ConfigDigest)
	epoch = uint32(resp.EpochAndRound.Epoch)
	round = uint8(resp.EpochAndRound.Round)
	latestAnswer = resp.Data.Answer.BigInt()
	latestTimestamp = time.Unix(resp.Data.TransmissionTimestamp, 0)
	err = nil

	return
}

func (c *CosmosMedianReporter) LatestRoundRequested(
	ctx context.Context,
	lookback time.Duration,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	round uint8,
	err error,
) {
	return
}
