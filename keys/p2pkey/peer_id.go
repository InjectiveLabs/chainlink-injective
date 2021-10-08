package p2pkey

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
)

const peerIDPrefix = "p2p_"

type PeerID peer.ID

func (p PeerID) String() string {
	return fmt.Sprintf("%s%s", peerIDPrefix, peer.ID(p).String())
}

func (p PeerID) Raw() string {
	return peer.ID(p).String()
}

func (p *PeerID) UnmarshalText(bs []byte) error {
	input := string(bs)
	if strings.HasPrefix(input, peerIDPrefix) {
		input = string(bs[len(peerIDPrefix):])
	}

	peerID, err := peer.Decode(input)
	if err != nil {
		return errors.Wrapf(err, `PeerID#UnmarshalText("%v")`, input)
	}
	*p = PeerID(peerID)
	return nil
}

func (p PeerID) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PeerID) UnmarshalJSON(input []byte) error {
	var result string
	if err := json.Unmarshal(input, &result); err != nil {
		return err
	}

	return p.UnmarshalText([]byte(result))
}