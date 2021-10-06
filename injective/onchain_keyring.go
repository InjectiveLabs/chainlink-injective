package injective

import (
	"bytes"

	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	secp256k1 "github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	ethsecp256k1 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

var _ types.OnchainKeyring = &InjectiveModuleOnchainKeyring{}

type InjectiveModuleOnchainKeyring struct {
	Signer  sdk.AccAddress
	Keyring keyring.Keyring
}

// PublicKey returns the acc address of the keypair used by Sign.
func (c *InjectiveModuleOnchainKeyring) PublicKey() types.OnchainPublicKey {
	return types.OnchainPublicKey(c.Signer.Bytes())
}

// Sign returns a signature over ReportContext and Report.
func (c *InjectiveModuleOnchainKeyring) Sign(
	reportCtx types.ReportContext,
	report types.Report,
) (signature []byte, err error) {
	onchainReport := chaintypes.ReportToSign{
		ConfigDigest: reportCtx.ConfigDigest[:],
		Epoch:        uint64(reportCtx.Epoch),
		Round:        uint64(reportCtx.Round),
		ExtraHash:    reportCtx.ExtraHash[:],
		Report:       []byte(report),
	}

	sig, _, err := c.Keyring.SignByAddress(c.Signer, onchainReport.Bytes())
	return sig, err
}

// Verify verifies a signature over ReportContext and Report allegedly
// created from OnchainPublicKey (acc address).
func (c *InjectiveModuleOnchainKeyring) Verify(
	acc types.OnchainPublicKey,
	reportCtx types.ReportContext,
	report types.Report,
	signature []byte,
) bool {
	pubKey, err := ethsecp256k1.RecoverPubkey([]byte(report), signature)
	if err != nil {
		return false
	}

	ecPubKey, err := ethcrypto.UnmarshalPubkey(pubKey)
	if err != nil {
		return false
	}

	signerAccBytes := (&secp256k1.PubKey{
		Key: ethcrypto.CompressPubkey(ecPubKey),
	}).Address().Bytes()

	return bytes.Equal(signerAccBytes, acc)
}

// Maximum length of a signature
func (c *InjectiveModuleOnchainKeyring) MaxSignatureLength() int {
	return 65
}
