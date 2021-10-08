package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	cli "github.com/jawher/mow.cli"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/utils"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chainlink-injective/keys/ocrkey"
	"github.com/InjectiveLabs/chainlink-injective/keys/p2pkey"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/ethsecp256k1"
	"github.com/InjectiveLabs/sdk-go/chain/crypto/hd"
)

const defaultKeyringKeyName = "cosmos"

var emptyCosmosAddress = cosmtypes.AccAddress{}

func initCosmosKeyring(
	cosmosKeyringDir *string,
	cosmosKeyringAppName *string,
	cosmosKeyringBackend *string,
	cosmosKeyFrom *string,
	cosmosKeyPassphrase *string,
	cosmosPrivKey *string,
	cosmosUseLedger *bool,
) (cosmtypes.AccAddress, keyring.Keyring, error) {
	switch {
	case len(*cosmosPrivKey) > 0:
		if *cosmosUseLedger {
			err := errors.New("cannot combine ledger and privkey options")
			return emptyCosmosAddress, nil, err
		}

		pkBytes, err := hexToBytes(*cosmosPrivKey)
		if err != nil {
			err = errors.Wrap(err, "failed to hex-decode cosmos account privkey")
			return emptyCosmosAddress, nil, err
		}

		// Specfic to Injective chain with Ethermint keys
		// Should be secp256k1.PrivKey for generic Cosmos chain
		cosmosAccPk := &ethsecp256k1.PrivKey{
			Key: pkBytes,
		}

		addressFromPk := cosmtypes.AccAddress(cosmosAccPk.PubKey().Address().Bytes())

		var keyName string

		// check that if cosmos 'From' specified separately, it must match the provided privkey,
		if len(*cosmosKeyFrom) > 0 {
			addressFrom, err := cosmtypes.AccAddressFromBech32(*cosmosKeyFrom)
			if err == nil {
				if !bytes.Equal(addressFrom.Bytes(), addressFromPk.Bytes()) {
					err = errors.Errorf("expected account address %s but got %s from the private key", addressFrom.String(), addressFromPk.String())
					return emptyCosmosAddress, nil, err
				}
			} else {
				// use it as a name then
				keyName = *cosmosKeyFrom
			}
		}

		if len(keyName) == 0 {
			keyName = defaultKeyringKeyName
		}

		// wrap a PK into a Keyring
		kb, err := KeyringForPrivKey(keyName, cosmosAccPk)
		return addressFromPk, kb, err

	case len(*cosmosKeyFrom) > 0:
		var fromIsAddress bool
		addressFrom, err := cosmtypes.AccAddressFromBech32(*cosmosKeyFrom)
		if err == nil {
			fromIsAddress = true
		}

		var passReader io.Reader = os.Stdin
		if len(*cosmosKeyPassphrase) > 0 {
			passReader = newPassReader(*cosmosKeyPassphrase)
		}

		var absoluteKeyringDir string
		if filepath.IsAbs(*cosmosKeyringDir) {
			absoluteKeyringDir = *cosmosKeyringDir
		} else {
			absoluteKeyringDir, _ = filepath.Abs(*cosmosKeyringDir)
		}

		kb, err := keyring.New(
			*cosmosKeyringAppName,
			*cosmosKeyringBackend,
			absoluteKeyringDir,
			passReader,
			hd.EthSecp256k1Option(),
		)
		if err != nil {
			err = errors.Wrap(err, "failed to init keyring")
			return emptyCosmosAddress, nil, err
		}

		var keyInfo keyring.Info
		if fromIsAddress {
			if keyInfo, err = kb.KeyByAddress(addressFrom); err != nil {
				err = errors.Wrapf(err, "couldn't find an entry for the key %s in keybase", addressFrom.String())
				return emptyCosmosAddress, nil, err
			}
		} else {
			if keyInfo, err = kb.Key(*cosmosKeyFrom); err != nil {
				err = errors.Wrapf(err, "could not find an entry for the key '%s' in keybase", *cosmosKeyFrom)
				return emptyCosmosAddress, nil, err
			}
		}

		switch keyType := keyInfo.GetType(); keyType {
		case keyring.TypeLocal:
			// kb has a key and it's totally usable
			return keyInfo.GetAddress(), kb, nil
		case keyring.TypeLedger:
			// the kb stores references to ledger keys, so we must explicitly
			// check that. kb doesn't know how to scan HD keys - they must be added manually before
			if *cosmosUseLedger {
				return keyInfo.GetAddress(), kb, nil
			}
			err := errors.Errorf("'%s' key is a ledger reference, enable ledger option", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		case keyring.TypeOffline:
			err := errors.Errorf("'%s' key is an offline key, not supported yet", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		case keyring.TypeMulti:
			err := errors.Errorf("'%s' key is an multisig key, not supported yet", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		default:
			err := errors.Errorf("'%s' key  has unsupported type: %s", keyInfo.GetName(), keyType)
			return emptyCosmosAddress, nil, err
		}

	default:
		err := errors.New("insufficient cosmos key details provided")
		return emptyCosmosAddress, nil, err
	}
}

func initOCRKey(
	ocrKeyringDir *string,
	ocrKeyID *string,
	ocrKeyPassphrase *string,
	ocrPrivKey *string,
) (keyID string, key ocrkey.KeyV2, err error) {
	switch {
	case len(*ocrPrivKey) > 0:
		pkBytes, err := hexToBytes(*ocrPrivKey)
		if err != nil {
			err = errors.Wrap(err, "failed to hex-decode OCR privkey")
			return "", ocrkey.KeyV2{}, err
		}

		key = ocrkey.Raw(pkBytes).Key()
		keyID = key.ID()

		// check that if OCR Key ID was specified separately, it must match the provided privkey
		if len(*ocrKeyID) > 0 {
			specKeyID := strings.ToLower(*ocrKeyID)
			specKeyID = strings.TrimPrefix(specKeyID, "0x")

			if keyID != specKeyID {
				err = errors.Errorf("expected OCR Key ID %s but got %s from the private key", specKeyID, keyID)
				return "", ocrkey.KeyV2{}, err
			}
		}

		return keyID, key, nil

	case len(*ocrKeyID) > 0:
		specKeyID := strings.ToLower(*ocrKeyID)
		specKeyID = strings.TrimPrefix(specKeyID, "0x")
		if _, err := hex.DecodeString(specKeyID); err != nil {
			return "", ocrkey.KeyV2{}, err
		}

		ocrFileEnc, err := ioutil.ReadFile(filepath.Join(*ocrKeyringDir, specKeyID+"_ocr.json"))
		if err != nil {
			return "", ocrkey.KeyV2{}, err
		}

		key, err = ocrkey.FromEncryptedJSON(ocrFileEnc, *ocrKeyPassphrase)
		if err != nil {
			return "", ocrkey.KeyV2{}, err
		}

		keyID = key.ID()
		if keyID != specKeyID {
			err = errors.Errorf("expected OCR Key ID %s but got %s from the encrypted OCR key file", specKeyID, keyID)
			return "", ocrkey.KeyV2{}, err
		}

		return keyID, key, nil

	default:
		err := errors.New("insufficient OCR key details provided")
		return "", ocrkey.KeyV2{}, err
	}
}

func initP2PKey(
	p2pKeyringDir *string,
	p2pPeerID *string,
	p2pKeyPassphrase *string,
	p2pPrivKey *string,
) (peerID p2pkey.PeerID, key p2pkey.Key, err error) {
	switch {
	case len(*p2pPrivKey) > 0:
		key = p2pkey.Base64ToPrivKey(*p2pPrivKey)
		peerID = key.MustGetPeerID()

		// check that if P2P Key ID was specified separately, it must match the provided privkey
		if len(*p2pPeerID) > 0 {
			specPeerID := strings.ToLower(*p2pPeerID)
			specPeerID = strings.TrimPrefix(specPeerID, "0x")

			if peerID != p2pkey.PeerID(specPeerID) {
				err = errors.Errorf("expected P2P Key ID %s but got %s from the private key", specPeerID, peerID)
				return "", p2pkey.Key{}, err
			}
		}

		return peerID, key, nil

	case len(*p2pPeerID) > 0:
		specPeerID := strings.ToLower(*p2pPeerID)
		specPeerID = strings.TrimPrefix(specPeerID, "0x")
		if _, err := hex.DecodeString(specPeerID); err != nil {
			return "", p2pkey.Key{}, err
		}

		ocrFileEnc, err := ioutil.ReadFile(filepath.Join(*p2pKeyringDir, specPeerID+"_peer.json"))
		if err != nil {
			return "", p2pkey.Key{}, err
		}

		key, err = p2pkey.FromEncryptedJSON(ocrFileEnc, *p2pKeyPassphrase)
		if err != nil {
			return "", p2pkey.Key{}, err
		}

		peerID = key.MustGetPeerID()
		if peerID != p2pkey.PeerID(specPeerID) {
			err = errors.Errorf("expected P2P Peer ID %s but got %s from the encrypted P2P key file", specPeerID, peerID)
			return "", p2pkey.Key{}, err
		}

		return peerID, key, nil

	default:
		err := errors.New("insufficient P2P key details provided")
		return "", p2pkey.Key{}, err
	}
}

func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

type passReader struct {
	pass string
	buf  *bytes.Buffer
}

var _ io.Reader = &passReader{}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return
}

// KeyringForPrivKey creates a temporary in-mem keyring for a PrivKey.
// Allows to init Context when the key has been provided in plaintext and parsed.
func KeyringForPrivKey(name string, privKey cryptotypes.PrivKey) (keyring.Keyring, error) {
	kb := keyring.NewInMemory(hd.EthSecp256k1Option())
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
	orFatal(err)

	return string(buf)
}

func orFatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func keystorePrefix() string {
	home, _ := os.UserHomeDir()

	if len(home) == 0 {
		return "keystore"
	}

	return filepath.Join(home, ".ocr2", "keystore")
}

func ensureDir(path string) {
	err := os.MkdirAll(path, 0700)
	orFatal(err)
}

func keysCmd(cmd *cli.Cmd) {
	cmd.Command("ocr", "Manage local OCR keys", func(sub *cli.Cmd) {
		sub.Command("add", "Generate new OCR key", ocrKeysAdd)
		sub.Command("delete", "Delete OCR key from local keystore", ocrKeysDelete)
		sub.Command("view", "Get and view OCR key by its Key ID", ocrKeysView)
		sub.Command("unsafe-export-pk", "Get and export OCR private key by its Key ID", ocrKeysExport)
		sub.Command("list", "List all OCR keys from the local keystore", ocrKeysList)
	})

	cmd.Command("p2p", "Manage local P2P keys", func(sub *cli.Cmd) {
		sub.Command("add", "Generate new P2P key", p2pKeysAdd)
		sub.Command("delete", "Delete P2P key from local keystore", p2pKeysDelete)
		sub.Command("view", "Get and view P2P key by its Key ID", p2pKeysView)
		sub.Command("unsafe-export-pk", "Get and export P2P private key by its Peer ID", p2pKeysExport)
		sub.Command("list", "List all P2P keys from the local keystore", p2pKeysList)
	})
}

func ocrKeysAdd(c *cli.Cmd) {
	ocrKeyringDir := c.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	c.Before = func() {
		ensureDir(*ocrKeyringDir)
	}

	c.Action = func() {
		key, err := ocrkey.NewV2()
		orFatal(err)

		keyFileName := fmt.Sprintf("%s_ocr.json", key.ID())
		keyFilePath := filepath.Join(*ocrKeyringDir, keyFileName)

		fmt.Println("Passphrase for the new OCR key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		data, err := key.ToEncryptedJSON(keyPassphrase, utils.DefaultScryptParams)
		err = ioutil.WriteFile(keyFilePath, data, 0600)
		orFatal(err)

		log.Infoln("Generated", key.String())
		log.Infoln("Key successfully saved to", keyFilePath)
	}
}

func ocrKeysDelete(c *cli.Cmd) {
	ocrKeyringDir := c.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	ocrKeyID := c.StringArg("OCR_KEY_ID", "", "Specify the OCR Key ID to locate and delete")

	c.Before = func() {
		ensureDir(*ocrKeyringDir)
	}

	c.Action = func() {
		specKeyID := strings.ToLower(*ocrKeyID)
		specKeyID = strings.TrimPrefix(specKeyID, "0x")
		_, err := hex.DecodeString(specKeyID)
		orFatal(errors.Wrap(err, "failed to decode key ID - must be a valid hex"))

		keyFileName := fmt.Sprintf("%s_ocr.json", specKeyID)
		keyFilePath := filepath.Join(*ocrKeyringDir, keyFileName)

		if info, err := os.Stat(keyFilePath); os.IsNotExist(err) || info.IsDir() {
			orFatal(errors.Errorf("Key file for %s not found", specKeyID))
		}

		ok := stdinConfirm(fmt.Sprintf("Erase key %s from the keystore? This cannot be undone. Continue? [y/N]", specKeyID))
		if !ok {
			log.Infoln("Canceled")
			return
		}

		err = os.Remove(keyFilePath)
		orFatal(err)

		log.Infoln("Deleted key file at", keyFilePath)
	}
}

func ocrKeysView(c *cli.Cmd) {
	ocrKeyringDir := c.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	ocrKeyID := c.StringArg("OCR_KEY_ID", "", "Specify the OCR Key ID to view the key")

	c.Before = func() {
		ensureDir(*ocrKeyringDir)
	}

	c.Action = func() {
		specKeyID := strings.ToLower(*ocrKeyID)
		specKeyID = strings.TrimPrefix(specKeyID, "0x")
		_, err := hex.DecodeString(specKeyID)
		orFatal(errors.Wrap(err, "failed to decode key ID - must be a valid hex"))

		keyFileName := fmt.Sprintf("%s_ocr.json", specKeyID)
		keyFilePath := filepath.Join(*ocrKeyringDir, keyFileName)

		data, err := os.ReadFile(keyFilePath)
		orFatal(err)

		fmt.Println("Passphrase to unlock the key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		key, err := ocrkey.FromEncryptedJSON(data, keyPassphrase)
		orFatal(err)

		log.Infoln("Loaded", key.String())

		v, _ := json.MarshalIndent(ocrkey.EncryptedOCRKeyExport{
			KeyType:           "OCR",
			ID:                key.ID(),
			OffChainPublicKey: key.OffChainSigning.PublicKey(),
			ConfigPublicKey:   key.PublicKeyConfig(),
		}, "", "\t")
		fmt.Println(string(v))
	}
}

func ocrKeysExport(c *cli.Cmd) {
	ocrKeyringDir := c.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	ocrKeyID := c.StringArg("OCR_KEY_ID", "", "Specify the OCR Key ID to export the private key")

	c.Before = func() {
		ensureDir(*ocrKeyringDir)
	}

	c.Action = func() {
		specKeyID := strings.ToLower(*ocrKeyID)
		specKeyID = strings.TrimPrefix(specKeyID, "0x")
		_, err := hex.DecodeString(specKeyID)
		orFatal(errors.Wrap(err, "failed to decode key ID - must be a valid hex"))

		keyFileName := fmt.Sprintf("%s_ocr.json", specKeyID)
		keyFilePath := filepath.Join(*ocrKeyringDir, keyFileName)

		data, err := os.ReadFile(keyFilePath)
		orFatal(err)

		fmt.Println("Passphrase to unlock the key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		key, err := ocrkey.FromEncryptedJSON(data, keyPassphrase)
		orFatal(err)

		log.Infoln("Loaded", key.String())

		ok := stdinConfirm("Showing private key reveals sensitive information. Continue? [y/N]")
		if !ok {
			log.Infoln("Canceled")
			return
		}

		fmt.Printf("%02X\n", []byte(key.Raw()))
	}
}

func ocrKeysList(c *cli.Cmd) {
	ocrKeyringDir := c.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	c.Before = func() {
		ensureDir(*ocrKeyringDir)
	}

	c.Action = func() {
		err := filepath.Walk(*ocrKeyringDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), "_ocr.json") {
				return nil
			}

			data, err := os.ReadFile(path)
			orFatal(err)

			var k ocrkey.EncryptedOCRKeyExport
			err = json.Unmarshal(data, &k)
			orFatal(err)

			fmt.Printf("* %s\n", k.ID)

			return nil
		})

		orFatal(err)
	}
}

func p2pKeysAdd(c *cli.Cmd) {
	p2pKeyringDir := c.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	c.Before = func() {
		ensureDir(*p2pKeyringDir)
	}

	c.Action = func() {
		key, err := p2pkey.CreateKey()
		orFatal(err)

		peerID := peer.ID(key.MustGetPeerID())

		keyFileName := fmt.Sprintf("%s_p2p.json", peerID.Pretty())
		keyFilePath := filepath.Join(*p2pKeyringDir, keyFileName)

		fmt.Println("Passphrase for the new P2P key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		data, err := key.ToEncryptedExport(keyPassphrase, utils.DefaultScryptParams)
		err = ioutil.WriteFile(keyFilePath, data, 0600)
		orFatal(err)

		log.Infof("Generated %s (%s)", peerID.Pretty(), peerID.ShortString())
		log.Infoln("Key successfully saved to", keyFilePath)
	}
}

func p2pKeysDelete(c *cli.Cmd) {
	p2pKeyringDir := c.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	p2pPeerID := c.StringArg("P2P_PEER_ID", "", "Specify the P2P Peer ID to locate and delete")

	c.Before = func() {
		ensureDir(*p2pKeyringDir)
	}

	c.Action = func() {
		peerID, err := peer.Decode(*p2pPeerID)
		orFatal(errors.Wrap(err, "failed to decode peer ID - must be a valid CID of a key or a raw multihash"))

		keyFileName := fmt.Sprintf("%s_p2p.json", peerID.Pretty())
		keyFilePath := filepath.Join(*p2pKeyringDir, keyFileName)

		if info, err := os.Stat(keyFilePath); os.IsNotExist(err) || info.IsDir() {
			orFatal(errors.Errorf("Key file for %s not found", peerID.Pretty()))
		}

		ok := stdinConfirm(fmt.Sprintf("Erase key %s from the keystore? This cannot be undone. Continue? [y/N]", peerID.Pretty()))
		if !ok {
			log.Infoln("Canceled")
			return
		}

		err = os.Remove(keyFilePath)
		orFatal(err)

		log.Infoln("Deleted key file at", keyFilePath)
	}
}

func p2pKeysView(c *cli.Cmd) {
	p2pKeyringDir := c.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	p2pPeerID := c.StringArg("P2P_PEER_ID", "", "Specify the P2P Peer ID to view the key")

	c.Before = func() {
		ensureDir(*p2pKeyringDir)
	}

	c.Action = func() {
		peerID, err := peer.Decode(*p2pPeerID)
		orFatal(errors.Wrap(err, "failed to decode peer ID - must be a valid CID of a key or a raw multihash"))

		keyFileName := fmt.Sprintf("%s_p2p.json", peerID.Pretty())
		keyFilePath := filepath.Join(*p2pKeyringDir, keyFileName)

		data, err := os.ReadFile(keyFilePath)
		orFatal(err)

		fmt.Println("Passphrase to unlock the key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		key, err := p2pkey.FromEncryptedJSON(data, keyPassphrase)
		orFatal(err)

		log.Infoln("Loaded", peerID.ShortString())
		bz, _ := key.GetPublic().Raw()

		v, _ := json.MarshalIndent(p2pkey.EncryptedP2PKeyExport{
			KeyType:   "P2P",
			PeerID:    key.MustGetPeerID(),
			PublicKey: p2pkey.PublicKeyBytes(bz),
		}, "", "\t")
		fmt.Println(string(v))
	}
}

func p2pKeysExport(c *cli.Cmd) {
	p2pKeyringDir := c.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	p2pPeerID := c.StringArg("P2P_PEER_ID", "", "Specify the P2P Peer ID to export the private key")

	c.Before = func() {
		ensureDir(*p2pKeyringDir)
	}

	c.Action = func() {
		peerID, err := peer.Decode(*p2pPeerID)
		orFatal(errors.Wrap(err, "failed to decode peer ID - must be a valid CID of a key or a raw multihash"))

		keyFileName := fmt.Sprintf("%s_p2p.json", peerID.Pretty())
		keyFilePath := filepath.Join(*p2pKeyringDir, keyFileName)

		data, err := os.ReadFile(keyFilePath)
		orFatal(err)

		fmt.Println("Passphrase to unlock the key: ")
		keyPassphrase, err := keyPassphraseFromStdin()
		orFatal(err)

		key, err := p2pkey.FromEncryptedJSON(data, keyPassphrase)
		orFatal(err)

		log.Infoln("Loaded", peerID.ShortString())

		ok := stdinConfirm("Showing private key reveals sensitive information. Continue? [y/N]")
		if !ok {
			log.Infoln("Canceled")
			return
		}

		fmt.Println(key.PrivKeyToBase64())
	}
}

func p2pKeysList(c *cli.Cmd) {
	p2pKeyringDir := c.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  keystorePrefix(),
	})

	c.Before = func() {
		ensureDir(*p2pKeyringDir)
	}

	c.Action = func() {
		err := filepath.Walk(*p2pKeyringDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			} else if info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(info.Name(), "_p2p.json") {
				return nil
			}

			data, err := os.ReadFile(path)
			orFatal(err)

			var k p2pkey.EncryptedP2PKeyExport
			err = json.Unmarshal(data, &k)
			orFatal(err)

			fmt.Printf("* %s (%s)\n", peer.ID(k.PeerID).Pretty(), peer.ID(k.PeerID).ShortString())

			return nil
		})

		orFatal(err)
	}
}
