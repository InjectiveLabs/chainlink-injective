package e2e

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"strings"
	"time"

	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/itchyny/gojq"
	. "github.com/onsi/ginkgo"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

// maxUInt256 returns a value equal to 2**256 - 1 (MAX_UINT in Solidity).
func maxUInt256() *big.Int {
	return new(big.Int).Sub(new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil), big.NewInt(1))
}

func toBool(s string) bool {
	switch strings.ToLower(s) {
	case "true", "1", "t", "yes":
		return true
	default:
		return false
	}
}

func orFail(err error) {
	if err != nil {
		Fail(err.Error(), 1)
	}
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func sumInts(n0 *big.Int, n ...*big.Int) *big.Int {
	sum := new(big.Int)
	if n0 != nil {
		sum.Set(n0)
	}

	for _, i := range n {
		sum.Add(sum, i)
	}

	return sum
}

func fromHex(h string) []byte {
	b, err := hex.DecodeString(h)
	orFail(err)

	return b
}

func fromHex32(h string) (bytes32 [32]byte) {
	b, err := hex.DecodeString(h)
	orFail(err)

	if n := copy(bytes32[:], b); n != 32 {
		panic("short read")
	}

	return bytes32
}

func getAddressOrFail(name string) cosmtypes.AccAddress {
	for _, a := range CosmosAccounts {
		if a.Name == name {
			return a.CosmosAccAddress
		}
	}

	orFail(errors.Errorf("unable to load Cosmos Account for: %s", name))
	return nil
}

func waitForService(ctx context.Context, conn *grpc.ClientConn) error {
	t := time.NewTimer(1 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("Service wait timed out. Please run injective exchange service:\n\nmake install && injective-exchange")
		case <-t.C:
			state := conn.GetState()

			if state != connectivity.Ready {
				t.Reset(5 * time.Second)
				continue
			}

			return nil
		}
	}
}

func grpcDialEndpoint(protoAddr string) (conn *grpc.ClientConn, err error) {
	conn, err = grpc.Dial(protoAddr, grpc.WithInsecure(), grpc.WithContextDialer(dialerFunc))
	if err != nil {
		err := errors.Wrapf(err, "failed to connect to the gRPC: %s", protoAddr)
		return nil, err
	}

	return conn, nil
}

// dialerFunc dials the given address and returns a net.Conn. The protoAddr argument should be prefixed with the protocol,
// eg. "tcp://127.0.0.1:8080" or "unix:///tmp/test.sock"
func dialerFunc(ctx context.Context, protoAddr string) (net.Conn, error) {
	proto, address := protocolAndAddress(protoAddr)
	conn, err := net.Dial(proto, address)
	return conn, err
}

// protocolAndAddress splits an address into the protocol and address components.
// For instance, "tcp://127.0.0.1:8080" will be split into "tcp" and "127.0.0.1:8080".
// If the address has no protocol prefix, the default is "tcp".
func protocolAndAddress(listenAddr string) (string, string) {
	protocol, address := "tcp", listenAddr
	parts := strings.SplitN(address, "://", 2)
	if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
	return protocol, address
}

func runJSONQuery(data []byte, q *gojq.Query, limit int) ([]interface{}, error) {
	var in interface{}
	if err := json.Unmarshal(data, &in); err != nil {
		err = errors.Wrapf(err, "failed to unmarshal JSON data: %s", string(data))
		return nil, err
	}

	results := make([]interface{}, 0, limit)

	iter := q.Run(in)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		results = append(results, v)
	}

	return results, nil
}

func decstr(str string) cosmtypes.Dec {
	d, _ := cosmtypes.NewDecFromStr(str)
	return d
}

func dec(v float64) cosmtypes.Dec {
	d := decimal.NewFromFloat(v)
	return cosmtypes.NewDecFromBigInt(d.BigInt())
}

func amount(v float64, decimals ...int) cosmtypes.Dec {
	var decimalsOpt int32 = 18
	if len(decimals) == 1 {
		decimalsOpt = int32(decimals[0])
	} else if len(decimals) > 1 {
		panic("too many arguments")
	}

	d := decimal.NewFromFloat(v).Shift(decimalsOpt)
	return cosmtypes.NewDecFromInt(cosmtypes.NewIntFromBigInt(d.BigInt()))
}

func coinAmount(v float64, denom string, decimals ...int) cosmtypes.Coin {
	var decimalsOpt int32 = 18
	if len(decimals) == 1 {
		decimalsOpt = int32(decimals[0])
	} else if len(decimals) > 1 {
		panic("too many arguments")
	}

	d := decimal.NewFromFloat(v).Shift(decimalsOpt)
	return cosmtypes.NewCoin(denom, cosmtypes.NewIntFromBigInt(d.BigInt()))
}

func coinsAmount(v float64, denom string, decimals ...int) cosmtypes.Coins {
	return cosmtypes.NewCoins(coinAmount(v, denom, decimals...))
}

func feedId(base, quote string) []byte {
	return keccak256([]byte(base + "\x01\x02" + quote))
}

func keccak256(data ...[]byte) (h []byte) {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}

	return d.Sum(nil)
}
