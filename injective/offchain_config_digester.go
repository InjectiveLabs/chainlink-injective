package injective

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
)

const ConfigDigestPrefixCosmos types.ConfigDigestPrefix = 2

var _ types.OffchainConfigDigester = CosmosOffchainConfigDigester{}

type CosmosOffchainConfigDigester struct {
	ChainID string
	FeedId  string

	MinAnswer   sdk.Dec
	MaxAnswer   sdk.Dec
	Description string
}

func (d CosmosOffchainConfigDigester) ConfigDigest(cc types.ContractConfig) (types.ConfigDigest, error) {
	if len(d.ChainID) == 0 {
		err := errors.New("ChainID is not set, but required")
		return types.ConfigDigest{}, err
	} else if len(d.FeedId) == 0 {
		err := errors.New("FeedId is not set, but required")
		return types.ConfigDigest{}, err
	}

	// TODO: figure out discrpancy between ContractConfig.Signers
	// and onchain values. Signers in offchain are expected to be pubkeys,
	// while chain code expects addresses.
	// ---------
	//
	// for _, pubkey := range cc.Signers {
	// 	signerAcc := sdk.AccAddress((&secp256k1.PubKey{
	// 		Key: pubkey,
	// 	}).Address().Bytes())

	// 	signers = append(signers, signerAcc.String())
	// }

	signers := make([]string, 0, len(cc.Signers))
	for _, acc := range cc.Signers {
		signers = append(signers, sdk.AccAddress(acc).String())
	}

	transmitters := make([]string, 0, len(cc.Transmitters))
	for _, acc := range cc.Transmitters {
		addr, err := sdk.AccAddressFromBech32(string(acc))
		if err != nil {
			return types.ConfigDigest{}, err
		}

		transmitters = append(transmitters, addr.String())
	}

	config := &chaintypes.FeedConfig{
		FeedId:                d.FeedId,
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint32(cc.F),
		OffchainConfigVersion: cc.OffchainConfigVersion,
		OffchainConfig:        cc.OffchainConfig,
		MinAnswer:             d.MinAnswer,
		MaxAnswer:             d.MaxAnswer,
		Description:           d.Description,
	}

	configDigest := configDigestFromBytes(config.Digest(d.ChainID))

	return configDigest, nil
}

func (d CosmosOffchainConfigDigester) ConfigDigestPrefix() types.ConfigDigestPrefix {
	return ConfigDigestPrefixCosmos
}

func configDigestFromBytes(buf []byte) types.ConfigDigest {
	var configDigest types.ConfigDigest

	if len(buf) != len(configDigest) {
		// assertion
		panic("buffer is not matching digest/hash length (32)")
	}

	if n := copy(configDigest[:], buf); n != len(configDigest) {
		// assertion
		panic("unexpectedly short read")
	}

	if configDigest[0] != 0 || types.ConfigDigestPrefix(configDigest[1]) != ConfigDigestPrefixCosmos {
		// assertion
		panic("wrong ConfigDigestPrefix")
	}

	return configDigest
}
