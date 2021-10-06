package p2pkey

import (
	"encoding/json"
	"strings"

	keystore "github.com/ethereum/go-ethereum/accounts/keystore"
	cryptop2p "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/utils"
)

const keyTypeIdentifier = "P2P"

// EncryptedP2PKeyExport represents the structure of P2P keys exported and imported
// to/from the disk
type EncryptedP2PKeyExport struct {
	KeyType   string               `json:"keyType"`
	PublicKey PublicKeyBytes       `json:"publicKey"`
	PeerID    PeerID               `json:"peerID"`
	Crypto    *keystore.CryptoJSON `json:"crypto,omitempty"`
}

func FromEncryptedJSON(keyJSON []byte, password string) (Key, error) {
	var export EncryptedP2PKeyExport
	if err := json.Unmarshal(keyJSON, &export); err != nil {
		return Key{}, err
	}

	key, err := export.DecryptPrivateKey(password)
	if err != nil {
		return Key{}, err
	}

	return *key, nil
}

func (k Key) ToEncryptedExport(passphrase string, scryptParams utils.ScryptParams) (export []byte, err error) {
	var marshalledPrivK []byte
	marshalledPrivK, err = cryptop2p.MarshalPrivateKey(k)
	if err != nil {
		return export, err
	}
	cryptoJSON, err := keystore.EncryptDataV3(marshalledPrivK, []byte(adulteratedPassword(passphrase)), scryptParams.N, scryptParams.P)
	if err != nil {
		return export, errors.Wrapf(err, "could not encrypt p2p key")
	}

	pubKeyBytes, err := k.GetPublic().Raw()
	if err != nil {
		return export, errors.Wrapf(err, "could not ger public key bytes from private key")
	}
	peerID, err := k.GetPeerID()
	if err != nil {
		return export, errors.Wrapf(err, "could not ger peerID from private key")
	}

	encryptedP2PKExport := EncryptedP2PKeyExport{
		KeyType:   keyTypeIdentifier,
		PublicKey: pubKeyBytes,
		PeerID:    peerID,
		Crypto:    &cryptoJSON,
	}
	return json.Marshal(encryptedP2PKExport)
}

// DecryptPrivateKey returns the PrivateKey in export, decrypted via passphrase, or an error
func (export EncryptedP2PKeyExport) DecryptPrivateKey(passphrase string) (k *Key, err error) {
	marshalledPrivK, err := keystore.DecryptDataV3(*export.Crypto, adulteratedPassword(passphrase))
	if err != nil {
		return k, errors.Wrapf(err, "could not decrypt key %s", strings.ToUpper(export.PublicKey.String()))
	}
	privK, err := cryptop2p.UnmarshalPrivateKey(marshalledPrivK)
	if err != nil {
		return k, errors.Wrapf(err, "could not unmarshal private key for %s", strings.ToUpper(export.PublicKey.String()))
	}
	return &Key{
		privK,
	}, nil
}

type Raw []byte

func (r Raw) Key() Key {
	privK, err := cryptop2p.UnmarshalPrivateKey([]byte(r))
	if err != nil {
		panic(errors.Wrap(err, "could not unmarshal private key"))
	}

	return Key{
		privK,
	}
}
