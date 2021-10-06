package median

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

type OffchainConfig struct {
	AlphaPPB uint64        `json:"alphaPPB"`
	DeltaC   time.Duration `json:"deltaC"`
}

func DecodeConfig(b []byte) (OffchainConfig, error) {
	var config OffchainConfig

	if err := json.Unmarshal(b, &config); err != nil {
		err = errors.Wrap(err, "failed to JSON unmarshal the offchain config")
		return OffchainConfig{}, err
	}

	return config, nil
}

func (c OffchainConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}
