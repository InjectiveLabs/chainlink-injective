package e2e

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	. "github.com/onsi/ginkgo"

	ocrtypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
)

const OracleTestAddress = "inj128jwakuw3wrq6ye7m4p64wrzc5rfl8tvwzc6s8"

var _ = Describe("OCR Feed Configs", func() {
	_ = Describe("Proposals to set configs", func() {
		clientCtx := map[string]client.Context{
			"validator1": getClientContext("validator1"),
			"validator2": getClientContext("validator2"),
			"validator3": getClientContext("validator3"),
			"user2":      getClientContext("user2"),
		}

		pairs := [][2]string{
			{"LINK", "USDT"},
			{"INJ", "USDT"},
		}

		proposals := make([]govtypes.Content, 0, len(pairs))
		proposerCtx := clientCtx["user2"]

		msgs := make([]sdk.Msg, 0, len(pairs))

		for _, pair := range pairs {
			feedId := pair[0] + "/" + pair[1]

			msgs = append(msgs, &ocrtypes.MsgFundFeedRewardPool{
				Sender: clientCtx["user2"].FromAddress.String(),
				FeedId: feedId,
				Amount: sdk.NewCoin("peggy0x514910771AF9Ca656af840dff83E8264EcF986CA", sdk.NewInt(1000000000000000000)),
			})

			proposals = append(proposals, &ocrtypes.SetConfigProposal{
				Title:       fmt.Sprintf("SetConfig Proposal for %s/%s", pair[0], pair[1]),
				Description: "Grants transmitter/signer privileges and sets feed config",
				Config: &ocrtypes.FeedConfig{
					FeedId:                feedId,
					Signers:               []string{OracleTestAddress},
					Transmitters:          []string{OracleTestAddress},
					F:                     0,
					OffchainConfigVersion: 0,
					OffchainConfig:        []byte("TEST_ENV"),
					MinAnswer:             sdk.SmallestDec(),
					MaxAnswer:             sdk.NewDec(99999999999999999),
					LinkPerObservation:    sdk.NewInt(10),
					LinkPerTransmission:   sdk.NewInt(69),
					LinkDenom:             "peggy0x514910771AF9Ca656af840dff83E8264EcF986CA",
					UniqueReports:         false,
					Description:           fmt.Sprintf("%s/%s Feed", pair[0], pair[1]),
				},
			})
		}

		It("Submits Governance Proposals and Funds Feed Reward Pool ", func() {
			submitGovernanceProposals(
				proposals,
				proposerCtx, // inj1hkhdaj2a2clmq5jq6mspsggqs32vynpk228q3r
				clientCtx["validator1"],
				clientCtx["validator2"],
				clientCtx["validator3"],
			)
		})

		It("Funds Feed Reward Pool ", func() {
			submitMsgs(proposerCtx, msgs...)
		})
	})
})
