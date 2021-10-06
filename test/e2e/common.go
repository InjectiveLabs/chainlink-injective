package e2e

import (
	"context"
	"encoding/json"
	"strconv"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
	"github.com/cosmos/cosmos-sdk/client"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
)

func submitGovernanceProposal(content govtypes.Content, proposerCtx client.Context, voterCtxs ...client.Context) {
	cc, err := chainclient.NewCosmosClient(proposerCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()

	proposalMsg, err := govtypes.NewMsgSubmitProposal(
		content,
		coinsAmount(1, "inj"),
		proposerCtx.FromAddress,
	)
	orFail(err)

	txResp, err := cc.SyncBroadcastMsg(proposalMsg)
	orFail(err)

	if txResp.Code != 0 {
		log.WithField("hash", txResp.TxHash).Warningln("Tx Error")
		orFail(errors.Errorf("deploy %s Tx error: %s", content.GetDescription(), txResp.String()))
	}

	log.WithField("hash", txResp.TxHash).Infoln("Sent Tx")

	jsonProposalID, _ := gojq.Parse(`.. | select(.key? == "proposal_id") | .value`)
	ids, err := runJSONQuery([]byte(txResp.RawLog), jsonProposalID, 1)
	orFail(err)

	proposalID, _ := strconv.ParseUint(ids[0].(string), 10, 64)
	log.Infof("Submitted proposal ID=%d", proposalID)

	for _, voterCtx := range voterCtxs {
		voteForProposal(voterCtx, proposalID)
	}

	log.Infof("Voted %d times YES", len(voterCtxs))
}

func submitGovernanceProposals(proposals []govtypes.Content, proposerCtx client.Context, voterCtxs ...client.Context) {
	cc, err := chainclient.NewCosmosClient(proposerCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()

	msgs := make([]cosmtypes.Msg, 0, len(proposals))
	for _, content := range proposals {
		proposalMsg, err := govtypes.NewMsgSubmitProposal(
			content,
			coinsAmount(1, "inj"),
			proposerCtx.FromAddress,
		)
		orFail(err)

		msgs = append(msgs, proposalMsg)
	}

	txResp, err := cc.SyncBroadcastMsg(msgs...)
	orFail(err)

	if txResp.Code != 0 {
		log.WithField("hash", txResp.TxHash).Warningln("Tx Error")
		proposalsText, _ := json.Marshal(proposals)
		orFail(errors.Errorf("submitting proposals %s Tx error: %s", string(proposalsText), txResp.String()))
	}

	log.WithField("hash", txResp.TxHash).Infoln("Sent Tx")

	jsonProposalID, _ := gojq.Parse(`.. | select(.key? == "proposal_id") | .value`)
	ids, err := runJSONQuery([]byte(txResp.RawLog), jsonProposalID, 1)
	orFail(err)

	proposalIDs := make([]uint64, 0, len(ids))
	seenIDs := make(map[uint64]struct{}, len(ids))

	for idx := range ids {
		proposalID, _ := strconv.ParseUint(ids[idx].(string), 10, 64)

		if _, ok := seenIDs[proposalID]; ok {
			continue
		} else {
			seenIDs[proposalID] = struct{}{}
		}

		log.Infof("Submitting proposal ID=%d", proposalID)
		proposalIDs = append(proposalIDs, uint64(proposalID))
	}

	for _, voterCtx := range voterCtxs {
		voteForProposals(voterCtx, proposalIDs...)
	}

	log.Infof("Voted %d times YES for %d proposals", len(voterCtxs), len(proposalIDs))
}

func voteForProposal(voterCtx client.Context, id uint64) {
	cc, err := chainclient.NewCosmosClient(voterCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()

	msgVote := govtypes.NewMsgVote(voterCtx.FromAddress, id, govtypes.OptionYes)

	txResp, err := cc.SyncBroadcastMsg(msgVote)
	orFail(err)

	if txResp.Code != 0 {
		log.WithField("hash", txResp.TxHash).Warningln("Tx Error")
		orFail(errors.Errorf("vote for a proposal (id=%d) Tx error: %s", id, txResp.String()))
	}

	log.WithField("hash", txResp.TxHash).Infoln("Sent Tx")
}

func voteForProposals(voterCtx client.Context, ids ...uint64) {
	cc, err := chainclient.NewCosmosClient(voterCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()

	msgs := make([]cosmtypes.Msg, 0, len(ids))

	for _, id := range ids {
		msgVote := govtypes.NewMsgVote(voterCtx.FromAddress, id, govtypes.OptionYes)
		msgs = append(msgs, msgVote)
	}

	txResp, err := cc.SyncBroadcastMsg(msgs...)
	orFail(err)

	if txResp.Code != 0 {
		log.WithField("hash", txResp.TxHash).Warningln("Tx Error")
		orFail(errors.Errorf("vote for proposals (ids: %v) Tx error: %s", ids, txResp.String()))
	}

	log.WithField("hash", txResp.TxHash).Infoln("Sent Tx")
}

func getBalanceOf(senderCtx client.Context, accountAddress, coinDenom string) cosmtypes.Int {
	cc, err := chainclient.NewCosmosClient(senderCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()

	bankClient := banktypes.NewQueryClient(cc.QueryClient())

	ctx := context.Background()
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: accountAddress,
		Denom:   coinDenom,
	})
	orFail(err)

	return resp.Balance.Amount
}

func submitMsgs(senderCtx client.Context, msgs ...cosmtypes.Msg) {
	cc, err := chainclient.NewCosmosClient(senderCtx, daemonGRPCEndpoint)
	orFail(err)
	defer cc.Close()


	txResp, err := cc.SyncBroadcastMsg(msgs...)
	orFail(err)

	if txResp.Code != 0 {
		log.WithField("hash", txResp.TxHash).Warningln("Tx Error")
		orFail(errors.Errorf("sending derivative limit order Tx error: %s", txResp.String()))
	}

	log.WithField("hash", txResp.TxHash).Infoln("Sent Tx")
}