package p2pkey

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"

	cryptop2p "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

// Key represents a libp2p private key
type Key struct {
	cryptop2p.PrivKey
}

// PublicKeyBytes is generated using cryptop2p.PubKey.Raw()
type PublicKeyBytes []byte

func (pkb PublicKeyBytes) String() string {
	return hex.EncodeToString(pkb)
}

func (pkb PublicKeyBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(pkb))
}

func (pkb *PublicKeyBytes) UnmarshalJSON(input []byte) error {
	var hexString string
	if err := json.Unmarshal(input, &hexString); err != nil {
		return err
	}

	result, err := hex.DecodeString(hexString)
	if err != nil {
		return err
	}

	*pkb = PublicKeyBytes(result)
	return nil
}

func (k Key) GetPeerID() (PeerID, error) {
	peerID, err := peer.IDFromPrivateKey(k)
	if err != nil {
		return PeerID(""), errors.WithStack(err)
	}
	return PeerID(peerID), err
}

func (k Key) MustGetPeerID() PeerID {
	peerID, err := peer.IDFromPrivateKey(k)
	if err != nil {
		panic(err)
	}
	return PeerID(peerID)
}

// CreateKey makes a new libp2p keypair from a crytographically secure entropy source
func CreateKey() (Key, error) {
	p2pPrivkey, _, err := cryptop2p.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return Key{}, nil
	}
	return Key{
		p2pPrivkey,
	}, nil
}

// type is added to the beginning of the passwords for
// P2P key, so that the keys can't accidentally be mis-used
// in the wrong place
func adulteratedPassword(auth string) string {
	s := "p2pkey" + auth
	return s
}
