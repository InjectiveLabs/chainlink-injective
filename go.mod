module github.com/InjectiveLabs/chainlink-injective

go 1.16

require (
	github.com/InjectiveLabs/sdk-go v1.26.1
	github.com/alexcesaro/statsd v2.0.0+incompatible
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/bugsnag/panicwrap v1.3.4 // indirect
	github.com/cosmos/cosmos-sdk v0.44.0
	github.com/ethereum/go-ethereum v1.10.8
	github.com/gin-gonic/gin v1.7.2
	github.com/gogo/protobuf v1.3.3
	github.com/itchyny/gojq v0.12.5
	github.com/jawher/mow.cli v1.2.0
	github.com/libp2p/go-libp2p-core v0.8.5
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.2.0
	github.com/smartcontractkit/chainlink v0.10.14
	github.com/smartcontractkit/libocr v0.0.0-20211027142358-580c720a7bcf
	github.com/tendermint/tendermint v0.34.13
	github.com/xlab/closer v0.0.0-20190328110542-03326addb7c2
	github.com/xlab/suplog v1.3.1
	go.mongodb.org/mongo-driver v1.7.2
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/btcsuite/btcutil => github.com/btcsuite/btcutil v1.0.2
