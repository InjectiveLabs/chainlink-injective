package injective

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/ocr2/reportingplugin/median"
	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
)

var _ types.ContractTransmitter = &CosmosModuleTransmitter{}

type CosmosModuleTransmitter struct {
	FeedId       string
	QueryClient  chaintypes.QueryClient
	CosmosClient chainclient.CosmosClient
}

func (c *CosmosModuleTransmitter) FromAccount() types.Account {
	return types.Account(c.CosmosClient.FromAddress().String())
}

// Transmit sends the report to the on-chain OCR2Aggregator smart contract's Transmit method
func (c *CosmosModuleTransmitter) Transmit(
	ctx context.Context,
	reportCtx types.ReportContext,
	report types.Report,
	signatures []types.AttributedOnChainSignature,
) error {
	if len(c.FeedId) == 0 {
		err := errors.New("CosmosModuleTransmitter has no FeedId set")
		return err
	}

	reportToSign, err := chaintypes.ReportFromBytes([]byte(report))
	if err != nil {
		return err
	}

	// TODO: design how to decouple Cosmos reporting from reportingplugins of OCR2
	// The reports are not necessarily numeric (see: titlerequest).
	var reportRaw median.Report
	if err := proto.Unmarshal([]byte(reportToSign.Report), &reportRaw); err != nil {
		err = errors.Wrap(err, "failed to unmarshal opaque report as median.Report")
		return err
	}

	msgTransmit := &chaintypes.MsgTransmit{
		Transmitter:  c.CosmosClient.FromAddress().String(),
		ConfigDigest: reportCtx.ConfigDigest[:],
		FeedId:       c.FeedId,
		Epoch:        uint64(reportCtx.Epoch),
		Round:        uint64(reportCtx.Round),
		ExtraHash:    reportCtx.ExtraHash[:],
		Report: &chaintypes.Report{ // chain only understands median.Report for now
			ObservationsTimestamp: reportRaw.ObservationsTimestamp,
			Observers:             reportRaw.Observers,
			Observations:          reportRaw.Observations,
		},
		Signatures: make([][]byte, len(signatures)),
	}
	for _, sig := range signatures {
		msgTransmit.Signatures[sig.Signer] = sig.Signature
	}

	txResp, err := c.CosmosClient.SyncBroadcastMsg(msgTransmit)
	if err != nil {
		return err
	}

	if txResp.Code != 0 {
		raw, _ := json.Marshal(txResp)
		return fmt.Errorf("cosmos Tx error: %s", string(raw))
	}

	return nil
}

func (c *CosmosModuleTransmitter) LatestConfigDigestAndEpoch(
	ctx context.Context,
) (
	configDigest types.ConfigDigest,
	epoch uint32,
	err error,
) {
	if len(c.FeedId) == 0 {
		err := errors.New("CosmosModuleTransmitter has no FeedId set")
		return types.ConfigDigest{}, 0, err
	}

	if c.QueryClient == nil {
		err := errors.New("cannot query LatestConfigDigestAndEpoch: no QueryClient set")
		return types.ConfigDigest{}, 0, err
	}

	resp, err := c.QueryClient.FeedConfigInfo(ctx, &chaintypes.QueryFeedConfigInfoRequest{
		FeedId: c.FeedId,
	})
	if err != nil {
		return types.ConfigDigest{}, 0, err
	}

	configDigest = configDigestFromBytes(resp.FeedConfigInfo.LatestConfigDigest)
	epoch = uint32(resp.EpochAndRound.Epoch)
	return configDigest, epoch, nil
}
