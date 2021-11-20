package model

import (
	"encoding/hex"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

type ID string

type Job struct {
	ObjectID primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	JobID     ID        `json:"jobId" bson:"jobId"`
	Spec      *JobSpec  `json:"spec" bson:"spec"`
	IsActive  bool      `json:"isActive" bson:"isActive"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type JobSpec struct {
	IsBootstrapPeer                        bool     `json:"isBootstrapPeer" bson:"isBootstrapPeer"`
	FeedID                                 ID       `json:"feedId" bson:"feedId"`
	KeyID                                  ID       `json:"keyId" bson:"keyId"`
	P2PBootstrapPeers                      []string `json:"p2pBootstrapPeers" bson:"p2pBootstrapPeers"`
	ContractConfigConfirmations            int      `json:"contractConfigConfirmations" bson:"contractConfigConfirmations"`
	ContractConfigTrackerSubscribeInterval string   `json:"contractConfigTrackerSubscribeInterval" bson:"contractConfigTrackerSubscribeInterval"`
	ObservationTimeout                     string   `json:"observationTimeout" bson:"observationTimeout"`
	BlockchainTimeout                      string   `json:"blockchainTimeout" bson:"blockchainTimeout"`
}

type JobPersistentState struct {
	ObjectID primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	JobID        ID `json:"jobId" bson:"jobId"`
	ConfigDigest ID `json:"configDigest" bson:"configDigest"`

	Epoch                uint32   `json:"epoch" bson:"epoch"`
	HighestSentEpoch     uint32   `json:"highestSentEpoch" bson:"highestSentEpoch"`
	HighestReceivedEpoch []uint32 `json:"highestReceivedEpoch" bson:"highestReceivedEpoch"`
}

type JobContractConfig struct {
	ObjectID primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	JobID        ID `json:"jobId" bson:"jobId"`
	ConfigDigest ID `json:"configDigest" bson:"configDigest"`

	ConfigCount           uint64     `json:"configCount" bson:"configCount"`
	Signers               []HexBytes `json:"signers" bson:"signers"`
	Transmitters          []Account  `json:"transmitters" bson:"transmitters"`
	F                     uint8      `json:"f" bson:"f"`
	OnchainConfig         HexBytes   `json:"onchainConfig" bson:"onchainConfig"`
	OffchainConfigVersion uint64     `json:"offchainConfigVersion" bson:"offchainConfigVersion"`
	OffchainConfig        []byte     `json:"offchainConfig" bson:"offchainConfig"`
}

type Account string

type JobPendingTransmission struct {
	ObjectID primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	JobID        ID `json:"jobId" bson:"jobId"`
	ConfigDigest ID `json:"configDigest" bson:"configDigest"`

	ReportTimestamp ReportTimestamp     `json:"reportTimestamp" bson:"reportTimestamp"`
	Transmission    PendingTransmission `json:"tx" bson:"tx"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type ReportTimestamp struct {
	Epoch uint32 `json:"epoch" bson:"epoch"`
	Round uint8  `json:"round" bson:"round"`
}

type PendingTransmission struct {
	Time                 time.Time                    `json:"time" bson:"time"`
	ExtraHash            HexBytes                     `json:"extraHash" bson:"extraHash"`
	Report               HexBytes                     `json:"report" bson:"report"`
	AttributedSignatures []AttributedOnchainSignature `json:"signatures" bson:"signatures"`
}

type AttributedOnchainSignature struct {
	Signature HexBytes `json:"signature" bson:"signature"`
	Signer    int      `json:"signer" bson:"signer"`
}

type JobPeerAnnouncement struct {
	ObjectID primitive.ObjectID `json:"-" bson:"_id,omitempty"`

	JobID     ID        `json:"jobId" bson:"jobId"`
	PeerID    ID        `json:"peerId" bson:"peerId"`
	Announce  []byte    `json:"announce" bson:"announce"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}

type Cursor struct {
	From  primitive.ObjectID  `json:"from,omitempty"`
	To    *primitive.ObjectID `json:"to,omitempty"`
	Limit int                 `json:"limit,omitempty"`
}

type HexBytes []byte

func (h HexBytes) MarshalJSON() ([]byte, error) {
	hex := hex.EncodeToString(h)
	buf := make([]byte, 0, len(hex)+2)
	buf = append(buf, '"')
	buf = append(buf, hex...)
	buf = append(buf, '"')
	return buf, nil
}

func (h HexBytes) MarshalBSONValue() (bsontype.Type, []byte, error) {
	buf := bsoncore.AppendString(nil, hex.EncodeToString([]byte(h)))
	return bsontype.String, buf, nil
}

func (h HexBytes) String() string {
	return hex.EncodeToString([]byte(h))
}

func (h *HexBytes) UnmarshalBSONValue(t bsontype.Type, src []byte) error {
	if t != bsontype.String {
		return errors.Errorf("bsontype(%s) not allowed in HexBytes.UnmarshalBSONValue", t.String())
	}

	v, _, ok := bsoncore.ReadString(src)
	if !ok {
		return errors.Errorf("bsoncore failed to read String")
	}

	data, err := hex.DecodeString(v)
	if err != nil {
		return errors.Wrapf(err, "bsoncore failed to decode hex string: %s", v)
	}

	*h = HexBytes(data)
	return nil
}
