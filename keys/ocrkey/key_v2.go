package ocrkey

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/utils"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"
	"golang.org/x/crypto/curve25519"
)

var (
	ErrScalarTooBig = errors.Errorf("can't handle scalars greater than %d", curve25519.PointSize)
)

type Raw []byte

func (raw Raw) Key() KeyV2 {
	if l := len(raw); l != 96 {
		panic(fmt.Sprintf("invalid raw key length: %d", l))
	}
	var key KeyV2
	var ed25519PrivKey []byte = raw[0:64]
	var offChainEncryption [32]byte
	copy(offChainEncryption[:], raw[64:])
	OffChainSigning := offChainPrivateKey(ed25519PrivKey)
	key.OffChainSigning = &OffChainSigning
	key.OffChainEncryption = &offChainEncryption
	return key
}

func (raw Raw) String() string {
	return "<OCR Raw Private Key>"
}

func (raw Raw) GoString() string {
	return raw.String()
}

type KeyV2 struct {
	OffChainSigning    *offChainPrivateKey
	OffChainEncryption *[curve25519.ScalarSize]byte
}

func NewV2() (KeyV2, error) {
	_, offChainPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyV2{}, err
	}
	var encryptionPriv [curve25519.ScalarSize]byte
	_, err = rand.Reader.Read(encryptionPriv[:])
	if err != nil {
		return KeyV2{}, err
	}
	return KeyV2{
		OffChainSigning:    (*offChainPrivateKey)(&offChainPriv),
		OffChainEncryption: &encryptionPriv,
	}, nil
}

func MustNewV2XXXTestingOnly(k *big.Int) KeyV2 {
	var seed [32]byte
	copy(seed[:], k.Bytes())
	offChainPriv := ed25519.NewKeyFromSeed(seed[:])
	return KeyV2{
		OffChainSigning:    (*offChainPrivateKey)(&offChainPriv),
		OffChainEncryption: &seed,
	}
}

func (key KeyV2) ID() string {
	sha := sha256.Sum256(key.Raw())
	return hex.EncodeToString(sha[:])
}

func (key KeyV2) Raw() Raw {
	return utils.ConcatBytes(
		[]byte(*key.OffChainSigning),
		key.OffChainEncryption[:],
	)
}

// SignOffChain returns an EdDSA-Ed25519 signature on msg.
func (key KeyV2) SignOffChain(msg []byte) (signature []byte, err error) {
	return key.OffChainSigning.Sign(msg)
}

// ConfigDiffieHellman returns the shared point obtained by multiplying someone's
// public key by a secret scalar ( in this case, the OffChainEncryption key.)
func (key KeyV2) ConfigDiffieHellman(base *[curve25519.PointSize]byte) (
	sharedPoint *[curve25519.PointSize]byte, err error,
) {
	p, err := curve25519.X25519(key.OffChainEncryption[:], base[:])
	if err != nil {
		return nil, err
	}
	sharedPoint = new([ed25519.PublicKeySize]byte)
	copy(sharedPoint[:], p)
	return sharedPoint, nil
}

// PublicKeyOffChain returns the pbulic component of the keypair used in SignOffChain
func (key KeyV2) PublicKeyOffChain() ocrtypes.OffchainPublicKey {
	return ocrtypes.OffchainPublicKey(key.OffChainSigning.PublicKey())
}

// PublicKeyConfig returns the public component of the keypair used in ConfigKeyShare
func (key KeyV2) PublicKeyConfig() [curve25519.PointSize]byte {
	rv, err := curve25519.X25519(key.OffChainEncryption[:], curve25519.Basepoint)
	if err != nil {
		log.Println("failure while computing public key: " + err.Error())
	}
	var rvFixed [curve25519.PointSize]byte
	copy(rvFixed[:], rv)
	return rvFixed
}

func (key KeyV2) GetID() string {
	return key.ID()
}

func (key KeyV2) String() string {
	return fmt.Sprintf("OCRKeyV2{ID: %s}", key.ID())
}

func (key KeyV2) GoString() string {
	return key.String()
}
