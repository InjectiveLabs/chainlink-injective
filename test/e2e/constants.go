package e2e

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"strings"

	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	chainhd "github.com/InjectiveLabs/sdk-go/chain/crypto/hd"
	"github.com/cosmos/cosmos-sdk/client"
	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	log "github.com/xlab/suplog"

	ctypes "github.com/InjectiveLabs/sdk-go/chain/types"
)

const (
	daemonGRPCEndpoint       = "tcp://localhost:9900"
	daemonTendermintEndpoint = "tcp://localhost:26657"
	testERC20TokenDenom      = "peggy0x69efCB62D98f4a6ff5a0b0CFaa4AAbB122e85e08"
	testERC20TokenDecimals   = 6
	injCoinDecimals          = 18
)

var (
// initialCommunityPoolBalance  = cosmtypes.NewInt(3_000_000).Mul(cosmtypes.NewInt(1e18))
)

func init() {
	readEnv()

	config := cosmtypes.GetConfig()
	ctypes.SetBech32Prefixes(config)
	ctypes.SetBip44CoinType(config)

	for idx, a := range CosmosAccounts {
		a.Parse()
		CosmosAccounts[idx] = a
	}
}

// readEnv is a special utility that reads `.env` file into actual environment variables
// of the current app, similar to `dotenv` Node package.
func readEnv() {
	if envdata, _ := ioutil.ReadFile(".env"); len(envdata) > 0 {
		s := bufio.NewScanner(bytes.NewReader(envdata))
		for s.Scan() {
			txt := s.Text()
			valIdx := strings.IndexByte(txt, '=')
			if valIdx < 0 {
				continue
			}

			strValue := strings.Trim(txt[valIdx+1:], `"`)
			if err := os.Setenv(txt[:valIdx], strValue); err != nil {
				log.WithField("name", txt[:valIdx]).WithError(err).Warningln("failed to override ENV variable")
			}
		}
	}
}

var CosmosAccounts = []Account{
	{Name: "validator1", Mnemonic: "remember huge castle bottom apology smooth avocado ceiling tent brief detect poem"},
	{Name: "validator2", Mnemonic: "capable dismiss rice income open wage unveil left veteran treat vast brave"},
	{Name: "validator3", Mnemonic: "jealous wrist abstract enter erupt hunt victory interest aim defy camp hair"},
	{Name: "user1", Mnemonic: "divide report just assist salad peanut depart song voice decide fringe stumble"},
	{Name: "user2", Mnemonic: "physical page glare junk return scale subject river token door mirror title"},
}

func getSigningKeys(accounts ...Account) []cryptotypes.PrivKey {
	privkeys := make([]cryptotypes.PrivKey, 0, len(accounts))
	for _, a := range accounts {
		privkeys = append(privkeys, a.PrivKey)
	}

	return privkeys
}

func getKeyrings(accounts ...Account) map[string]keyring.Keyring {
	keyrings := make(map[string]keyring.Keyring, len(accounts))
	for _, a := range accounts {
		keyrings[a.Name] = a.Keyring
	}

	return keyrings
}

func getClientContext(from ...string) client.Context {
	if len(from) == 0 {
		ctx, err := chainclient.NewClientContext("injective-1", "", nil)
		orPanic(err)

		return ctx
	}
	fromName := from[0]

	keyrings := getKeyrings(CosmosAccounts...)
	if _, ok := keyrings[fromName]; !ok {
		orPanic(errors.Errorf("account not found in keyrings: %s", fromName))
	}

	ctx, err := chainclient.NewClientContext("injective-1", fromName, keyrings[fromName])
	orPanic(err)

	ctx = ctx.WithNodeURI(daemonTendermintEndpoint)
	tmRPC, err := rpchttp.New(daemonTendermintEndpoint, "/websocket")
	if err != nil {
		orFail(err)
	}

	ctx = ctx.WithClient(tmRPC)
	return ctx
}

type Account struct {
	Name     string
	Address  string
	Key      string
	Mnemonic string

	CosmosAccAddress cosmtypes.AccAddress
	CosmosValAddress cosmtypes.ValAddress

	PrivKey cryptotypes.PrivKey
	Keyring keyring.Keyring
}

func (a *Account) Parse() {
	if len(a.Mnemonic) > 0 {
		// derive address and privkey from the provided mnemonic

		algo, err := keyring.NewSigningAlgoFromString("secp256k1", keyring.SigningAlgoList{
			hd.Secp256k1,
		})
		orPanic(err)

		pkBytes, err := algo.Derive()(a.Mnemonic, "", cosmtypes.GetConfig().GetFullFundraiserPath())

		cosmosAccPk := &ethsecp256k1.PrivKey{
			Key: pkBytes,
		}

		a.PrivKey = cosmosAccPk
		a.Address = cosmtypes.AccAddress(cosmosAccPk.PubKey().Address().Bytes()).String()

	} else if len(a.Key) > 0 {
		pkBytes, err := hex.DecodeString(a.Key)
		orPanic(err)

		cosmosAccPk := &ethsecp256k1.PrivKey{
			Key: pkBytes,
		}

		a.PrivKey = cosmosAccPk
	}

	if accAddress, err := cosmtypes.AccAddressFromBech32(a.Address); err == nil {
		// provided a Bech32 address

		if a.PrivKey != nil {
			a.Keyring, err = KeyringForPrivKey(a.Name, a.PrivKey)
			orPanic(err)

			if !bytes.Equal(a.PrivKey.PubKey().Address().Bytes(), accAddress.Bytes()) {
				panic(errors.Errorf("privkey doesn't match address: %s", accAddress.String()))
			}
		}

		a.CosmosAccAddress = accAddress
		a.CosmosValAddress = cosmtypes.ValAddress(accAddress.Bytes())
	} else if err != nil {
		panic(errors.Wrapf(err, "failed to parse address: %s", a.Address))
	} else {
		panic(errors.Errorf("unsupported address: %s", a.Address))
	}
}

// KeyringForPrivKey creates a temporary in-mem keyring for a PrivKey.
// Allows to init Context when the key has been provided in plaintext and parsed.
func KeyringForPrivKey(name string, privKey cryptotypes.PrivKey) (keyring.Keyring, error) {
	kb := keyring.NewInMemory(chainhd.EthSecp256k1Option())
	tmpPhrase := randPhrase(64)
	armored := cosmcrypto.EncryptArmorPrivKey(privKey, tmpPhrase, privKey.Type())
	err := kb.ImportPrivKey(name, armored, tmpPhrase)
	if err != nil {
		err = errors.Wrap(err, "failed to import privkey")
		return nil, err
	}

	return kb, nil
}

func randPhrase(size int) string {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	orPanic(err)

	return string(buf)
}
