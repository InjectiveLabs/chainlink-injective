//go:generate protoc -I. --go_out=. ./offchain_config.proto

package config

import (
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func DecodeConfig(b []byte) (*OffchainConfig, error) {
	var config OffchainConfig

	if err := proto.Unmarshal(b, &config); err != nil {
		err = errors.Wrap(err, "failed to proto–unmarshal the offchain config")
		return nil, err
	}

	return &config, nil
}

func (c *OffchainConfig) Encode() ([]byte, error) {
	return proto.Marshal(c)
}
