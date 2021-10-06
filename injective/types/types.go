package types

import (
	"bytes"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

const FeedIDMaxLength = 20

var digestPrefixCosmos = []byte("\x00\x02")
var digestSeparator = []byte("\x01\x02")

func (cfg *FeedConfig) Digest(chainID string) []byte {
	data, err := proto.Marshal(cfg)
	if err != nil {
		panic("unmarshable")
	}

	w := sha3.NewLegacyKeccak256()
	w.Write(data)
	w.Write(digestSeparator)
	w.Write([]byte(chainID))

	configDigest := w.Sum(nil)
	configDigest[0] = digestPrefixCosmos[0]
	configDigest[1] = digestPrefixCosmos[1]
	return configDigest
}

func (cfg *FeedConfig) ValidTransmitters() map[string]struct{} {
	transmitters := make(map[string]struct{})
	for _, transmitter := range cfg.Transmitters {
		transmitters[transmitter] = struct{}{}
	}
	return transmitters
}

func (cfg *FeedConfig) TransmitterFromSigner() map[string]sdk.AccAddress {
	transmitterFromSigner := make(map[string]sdk.AccAddress)
	for idx, signer := range cfg.Signers {
		addr, _ := sdk.AccAddressFromBech32(cfg.Transmitters[idx])
		transmitterFromSigner[signer] = addr
	}
	return transmitterFromSigner
}

func (cfg *FeedConfig) ValidateBasic() error {
	if err := checkConfigValid(
		len(cfg.Signers),
		len(cfg.Transmitters),
		int(cfg.F),

		// usually opaque to chain, but we use this for testing
		cfg.OffchainConfig,
	); err != nil {
		return err
	}

	// TODO: determine whether this is a sensible enough limitation
	if len(cfg.FeedId) > FeedIDMaxLength {
		return sdkerrors.Wrap(ErrIncorrectConfig, "feed_id is missing or incorrect length")
	}

	if strings.TrimSpace(cfg.FeedId) != cfg.FeedId {
		return sdkerrors.Wrap(ErrIncorrectConfig, "feed_id cannot have leading or trailing space characters")
	}

	if cfg.FeedAdmin != "" {
		if _, err := sdk.AccAddressFromBech32(cfg.FeedAdmin); err != nil {
			return err
		}
	}

	if cfg.BillingAdmin != "" {
		if _, err := sdk.AccAddressFromBech32(cfg.BillingAdmin); err != nil {
			return err
		}
	}

	if cfg.MinAnswer.IsNil() || cfg.MaxAnswer.IsNil() {
		return sdkerrors.Wrap(ErrIncorrectConfig, "MinAnswer and MaxAnswer cannot be nil")
	}

	if cfg.LinkPerTransmission.IsNil() || !cfg.LinkPerTransmission.IsPositive() {
		return sdkerrors.Wrap(ErrIncorrectConfig, "LinkPerTransmission must be positive")
	}

	if cfg.LinkPerObservation.IsNil() || !cfg.LinkPerObservation.IsPositive() {
		return sdkerrors.Wrap(ErrIncorrectConfig, "LinkPerObservation must be positive")
	}

	seenTransmitters := make(map[string]struct{}, len(cfg.Transmitters))
	for _, transmitter := range cfg.Transmitters {
		addr, err := sdk.AccAddressFromBech32(transmitter)
		if err != nil {
			return err
		}

		if _, ok := seenTransmitters[addr.String()]; ok {
			return ErrRepeatedAddress
		} else {
			seenTransmitters[addr.String()] = struct{}{}
		}
	}

	seenSigners := make(map[string]struct{}, len(cfg.Signers))
	for _, signer := range cfg.Signers {
		addr, err := sdk.AccAddressFromBech32(signer)
		if err != nil {
			return err
		}

		if _, ok := seenSigners[addr.String()]; ok {
			return ErrRepeatedAddress
		} else {
			seenSigners[addr.String()] = struct{}{}
		}
	}

	if cfg.LinkDenom == "" {
		return sdkerrors.ErrInvalidCoins
	}

	return nil
}

var testEnvHeader = []byte("TEST_ENV")

func checkConfigValid(
	numSigners, numTransmitters, f int,
	offchainConfig []byte,
) error {
	if numSigners > MaxNumOracles {
		return ErrTooManySigners
	}

	if f == 0 && bytes.Equal(offchainConfig, testEnvHeader) {
		// a special case for testing
	} else if f <= 0 {
		return sdkerrors.Wrap(ErrIncorrectConfig, "f must be positive")
	}

	if numSigners != numTransmitters {
		return sdkerrors.Wrap(ErrIncorrectConfig, "oracle addresses out of registration")
	}

	if numSigners <= 3*f {
		return sdkerrors.Wrapf(ErrIncorrectConfig, "faulty-oracle f too high: %d", f)
	}

	return nil
}

func ReportFromBytes(buf []byte) (*ReportToSign, error) {
	var r ReportToSign
	if err := proto.Unmarshal(buf, &r); err != nil {
		err = errors.Wrap(err, "failed to proto-decode ReportToSign from bytes")
		return nil, err
	}

	return &r, nil
}

func (r *ReportToSign) Bytes() []byte {
	data, err := proto.Marshal(r)
	if err != nil {
		panic("unmarshable")
	}

	return data
}

func (r *ReportToSign) Digest() []byte {
	w := sha3.NewLegacyKeccak256()
	w.Write(r.Bytes())
	return w.Sum(nil)
}

type Reward struct {
	Addr   sdk.AccAddress
	Amount sdk.Coin
}
