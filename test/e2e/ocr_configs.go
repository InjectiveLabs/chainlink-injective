package e2e

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	. "github.com/onsi/ginkgo"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/ocr2/config"
	ocrconfig "github.com/InjectiveLabs/chainlink-injective/ocr2/config"
)

var _ = Describe("OCR Feed Configs", func() {
	_ = Describe("Proposals to set configs", func() {
		clientCtx := map[string]client.Context{
			"validator1": getClientContext("validator1"),
			"validator2": getClientContext("validator2"),
			"validator3": getClientContext("validator3"),
			"user2":      getClientContext("user2"),
		}

		pairs := [][2]string{
			{"LINK", "USDC"},
			{"INJ", "USDC"},
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
					Signers: []string{
						getAddressOrFail("oracle0").String(),
						getAddressOrFail("oracle1").String(),
						getAddressOrFail("oracle2").String(),
						getAddressOrFail("oracle3").String(),
					},
					Transmitters: []string{
						getAddressOrFail("oracle0").String(),
						getAddressOrFail("oracle1").String(),
						getAddressOrFail("oracle2").String(),
						getAddressOrFail("oracle3").String(),
					},
					F: 1,
					OnchainConfig: makeMedianReportingOnchainConfig(
						sdk.SmallestDec().BigInt(),
						sdk.NewDec(99999999999999999).BigInt(),
					),
					OffchainConfigVersion: 2, // OCR2
					OffchainConfig:        makeFastChainOffchainConfig(),
					ModuleParams: &chaintypes.ModuleParams{
						FeedId:              feedId,
						MinAnswer:           sdk.SmallestDec(),
						MaxAnswer:           sdk.NewDec(99999999999999999),
						LinkPerObservation:  sdk.NewInt(10),
						LinkPerTransmission: sdk.NewInt(69),
						LinkDenom:           "peggy0x514910771AF9Ca656af840dff83E8264EcF986CA",
						UniqueReports:       false,
						Description:         fmt.Sprintf("%s/%s Feed", pair[0], pair[1]),
					},
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

func makeMedianReportingOnchainConfig(min, max *big.Int) []byte {
	config := &median.OnchainConfig{
		Min: min,
		Max: max,
	}

	configBytes, err := config.Encode()
	orFail(err)

	return configBytes
}

func makeFastChainOffchainConfig() []byte {
	medianPluginConfig := &median.OffchainConfig{
		AlphaReportPPB: 1000000000 / 100, // threshold PPB
		AlphaAcceptPPB: 1000000000 / 100, // threshold PPB
		DeltaC:         10 * time.Second,
	}

	sharedSecretEncryptionPublicKeys := []ocrtypes.SharedSecretEncryptionPublicKey{
		fromHex32("2f70f0dda48830c8bcbe465cf3f5b5712a2abf5b1753e9116246a3f67d29b61b"),
		fromHex32("15acff142c476a20769bdffc28c32414a0c75cd1769c26b683804fd5f163a852"),
		fromHex32("7d18b47f02293cc9ef3746e593bda0ee3aed6cf70943a585213c7c33fa77d314"),
		fromHex32("419c1a2fe81f0ce832cf471e7c294daf7479ec9f9a1ce935ab27eaeafd348876"),
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

	// See https://research.chain.link/ocr.pdf
	// for the reference on protocol variables
	//
	config := &ocrconfig.OffchainConfig{
		DeltaStage:    uint64(5 * time.Second),
		DeltaRound:    uint64(5 * time.Second),
		DeltaProgress: uint64(8 * time.Second),
		DeltaResend:   uint64(5 * time.Second),
		DeltaGrace:    uint64(3 * time.Second),
		RMax:          254,
		S: []uint32{
			1, 1, 1, 1,
		},
		OffchainPublicKeys: [][]byte{
			fromHex("35c5877d26acddadb4d915edfb5c66a427a3ced8328292159afc90980d145c5c"),
			fromHex("c863ac73bc720c79b34cb053d81a9bdf2c7094f7314ff32e6ca6ea7519da220a"),
			fromHex("b73208d0b23f82c20b10ef659bffcea7137e404ce31cf44cf7e6656b06c6ebd2"),
			fromHex("6a125a2236905c16615977b6d0b059b19857d5fa10d252e85f9f58078a02470a"),
		},
		PeerIds: []string{
			"12D3KooWEoy4KrP3uwd4uZmDFBfKur2F5zSNTVMSwymQ9iNCFt7Z",
			"12D3KooWHgoKkzaNGKYK39PMjyH3tPBx1iDHmEHzrBCmuKhn4C8F",
			"12D3KooWJLRX7N1aP1XSS7vHzireeBcs7m9Kv321FqXCCPcwB2P2",
			"12D3KooWT2mPa5onqXGkicvaQUHSW6d6AVWc5CLqxMSQTfQCDgcq",
		},

		ReportingPluginConfig: medianPluginConfig.Encode(),

		// NOTE: sum of MaxDurationQuery/Observation/Report (7.5s) must be less than DeltaProgress (8s)
		MaxDurationQuery:       uint64(2500 * time.Millisecond),
		MaxDurationObservation: uint64(2500 * time.Millisecond),
		MaxDurationReport:      uint64(2500 * time.Millisecond),

		MaxDurationShouldAcceptFinalizedReport:  uint64(2500 * time.Millisecond),
		MaxDurationShouldTransmitAcceptedReport: uint64(2500 * time.Millisecond),

		SharedSecretEncryptions: sharedSecretEncryptions.Proto(),
	}

	configBytes, err := config.Encode()
	orFail(err)

	return configBytes
}
