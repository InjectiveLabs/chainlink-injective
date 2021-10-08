//go:generate protoc -I. --go_out=. ./median_plugin_config.proto

package median

import (
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func DecodeConfig(b []byte) (*MedianPluginConfig, error) {
	var config MedianPluginConfig

	if err := proto.Unmarshal(b, &config); err != nil {
		err = errors.Wrap(err, "failed to proto–unmarshal the median plugin config")
		return nil, err
	}

	return &config, nil
}

func (c *MedianPluginConfig) Encode() ([]byte, error) {
	return proto.Marshal(c)
}
