package e2e

import (
	cryptorand "crypto/rand"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	. "github.com/onsi/ginkgo"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/ocr2/config"
	ocrconfig "github.com/InjectiveLabs/chainlink-injective/ocr2/config"
	"github.com/InjectiveLabs/chainlink-injective/ocr2/reportingplugin/median"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting/types"
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

		feedFundMsgs := make([]sdk.Msg, 0, len(pairs))

		for _, pair := range pairs {
			feedId := pair[0] + "/" + pair[1]

			proposals = append(proposals, &chaintypes.SetConfigProposal{
				Title:       fmt.Sprintf("SetConfig Proposal for %s/%s", pair[0], pair[1]),
				Description: "Grants transmitter/signer privileges and sets feed config",
				Config: &chaintypes.FeedConfig{
					Signers:      []string{OracleTestAddress},
					Transmitters: []string{OracleTestAddress},
					F:            1,
					OnchainConfig: &chaintypes.OnchainConfig{
						FeedId:              feedId,
						MinAnswer:           sdk.SmallestDec(),
						MaxAnswer:           sdk.NewDec(99999999999999999),
						LinkPerObservation:  sdk.NewInt(10),
						LinkPerTransmission: sdk.NewInt(69),
						LinkDenom:           "peggy0x514910771AF9Ca656af840dff83E8264EcF986CA",
						UniqueReports:       false,
						Description:         fmt.Sprintf("%s/%s Feed", pair[0], pair[1]),
					},
					OffchainConfigVersion: 2, // OCR2
					OffchainConfig:        makeFastChainOffchainConfig(),
				},
			})

			feedFundMsgs = append(feedFundMsgs, &chaintypes.MsgFundFeedRewardPool{
				Sender: clientCtx["user2"].FromAddress.String(),
				FeedId: feedId,
				Amount: sdk.NewCoin("peggy0x514910771AF9Ca656af840dff83E8264EcF986CA", sdk.NewInt(1000000000000000000)),
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

			// wait until feeds are approved
			time.Sleep(10 * time.Second)

			submitMsgs(proposerCtx, feedFundMsgs...)
		})
	})
})

func makeFastChainOffchainConfig() []byte {
	medianPluginConfig := &median.MedianPluginConfig{
		AlphaPpb: 1000000000 / 100, // threshold PPB
		DeltaC:   uint64(10 * time.Second),
	}
	medianPluginConfigBytes, err := medianPluginConfig.Encode()
	orFail(err)

	sharedSecretEncryptionPublicKeys := []ocrtypes.SharedSecretEncryptionPublicKey{
		fromHex32("376f1e7c6dcc5248fcf439471dc00bacdd195275d07fbd3245c2d941eec2e91e"),
	}

	// expected to be shared only with oracles in a super secret private chat I guess
	sharedSecret := [config.SharedSecretSize]byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
	}

	sharedSecretEncryptions := config.EncryptSharedSecret(
		sharedSecretEncryptionPublicKeys,
		&sharedSecret,
		cryptorand.Reader,
	)

	config := &ocrconfig.OffchainConfig{
		DeltaStage:    uint64(5 * time.Second),
		DeltaRound:    uint64(5 * time.Second),
		DeltaProgress: uint64(8 * time.Second),
		DeltaResend:   uint64(5 * time.Second),
		DeltaGrace:    uint64(2500 * time.Millisecond),

		RMax: 254,
		S: []uint32{
			1,
		},
		OffchainPublicKeys: [][]byte{
			// Loaded OCRKeyV2{ID: f7b80d092a4c328ef52508d2cef17f4f31d16293729e19c62f9ad6cb59a961a0}
			// {
			// 	"keyType": "OCR",
			// 	"id": "f7b80d092a4c328ef52508d2cef17f4f31d16293729e19c62f9ad6cb59a961a0",
			// 	"offChainPublicKey": "ocroff_4d203a02f68441df3ce8a8678f5c8d2a9628df25bb9d98d4425db25c29df3422",
			// 	"configPublicKey": "ocrcfg_376f1e7c6dcc5248fcf439471dc00bacdd195275d07fbd3245c2d941eec2e91e"
			// }
			fromHex("4d203a02f68441df3ce8a8678f5c8d2a9628df25bb9d98d4425db25c29df3422"),
		},
		PeerIds: []string{
			"12D3KooWPaHvunmPm3qjhsffgZBd2rQS4tdCgSYWEeRiX6hDsrdq",
		},

		ReportingPluginConfig: medianPluginConfigBytes,

		MaxDurationQuery:                        uint64(250 * time.Millisecond),
		MaxDurationObservation:                  uint64(250 * time.Millisecond),
		MaxDurationReport:                       uint64(250 * time.Millisecond),
		MaxDurationShouldAcceptFinalizedReport:  uint64(250 * time.Millisecond),
		MaxDurationShouldTransmitAcceptedReport: uint64(250 * time.Millisecond),

		SharedSecretEncryptions: sharedSecretEncryptions.Proto(),
	}

	configBytes, err := config.Encode()
	orFail(err)

	return configBytes
}
