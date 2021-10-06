# chainlink-injective

Price feed OCR integration demo with libocr2 and Cosmos module, tailored for Injective Chain.

There is the instruction on running it.

## Step 1

Clone https://github.com/InjectiveLabs/injective-core on branch `f/ocr` and build the chain node. It will have the OCR chain module that is ready to accept transmissions. This handler https://github.com/InjectiveLabs/injective-core/blob/bd2d7adc2578ea9be956d60ba9f706bec2716728/injective-chain/modules/ocr/keeper/msg_server.go#L41 will take care of any incoming `MsgTransmit` from chain clients.

Installing the node is as simple as running `make install` in the corresponding repo.

## Step 2

Clone https://github.com/InjectiveLabs/libocr-internal on branch `ocr2-inj` and remember its location. This lib has been enhanced with a chain client, that is ready to work with Injective Chain out of the box, and with slight changes will be able to handle vanilla Cosmos chains as well. The chain adapter is located there https://github.com/InjectiveLabs/libocr-internal/tree/ocr2-inj/offchainreporting2/chains/injective

The package itself doesn't contain executables.

## Step 3

Finally, clone https://github.com/InjectiveLabs/chainlink-injective (this repo on `master`) and let's run it. Make sure `injectived` executable is available on your system. Also [jq](https://stedolan.github.io/jq/) tool is required.

### Preps

1) Adjust `go.mod` in this repo, specifically this line:

```
replace github.com/smartcontractkit/libocr => /Users/xlab/Documents/dev/InjectiveLabs/libocr-internal
```

2) Copy `.env.example` to `.env` and you can adjust it too, if needed.

### Running a 3-node consensus

Start a mock network of 3 injectived nodes with proper consensus, governance and some pre-baked accounts with balances.

```
$ CLEANUP=1 ./test/e2e_multinode.sh injectived
```

Omit `CLEANUP=1` when running this command second time, it restarts the chain (all 3 nodes). If you include it again, it will wipe state, then restart. The last argument is the path to the node binary, can be arbitrary to `$GOHOME/bin/injectived` or where it was installed.

```
Starting nodes...
Waiting for chains to start...

Logs:
  * tail -f ./var/data/injective-1.n0.log
  * tail -f ./var/data/injective-1.n1.log
  * tail -f ./var/data/injective-1.n2.log

Env for easy access:
export H1='--home ./var/data/injective-1/n0/'
export H2='--home ./var/data/injective-1/n1/'
export H3='--home ./var/data/injective-1/n2/'

Command Line Access:
  * injectived --home ./var/data/injective-1/n0 status
  * injectived --home ./var/data/injective-1/n1 status
  * injectived --home ./var/data/injective-1/n2 status
```

You can use `tail -f ./var/data/injective-1.n0.log` to tail the logs of the first node, that listens on `:9900` (main GRPC interface for clients) and `:26657` (Tendermint Core low-level RPC). Both will be used for our `ocr-pricefeed` example and are specified in the env config.

### Running migrations

Since everything has been pre-baked already, even the Oracle's private key, we have nothing to migrate except make some SetConfig proposals. Voting time on the mock chain is set to 5 seconds, so the voting will pass quickly.

Run migrations with:

```
$ make test
```

According to this source, the proposal will set a config for both feeds (`LINK/USDT`, `INJ/USDT`) and authrize our test oracle to sign and transmit later:
https://github.com/InjectiveLabs/chainlink-injective/blob/9cc92dc9a4196348c56f10720170c405fe6b082d/test/e2e/ocr_configs.go#L31-L48

```
Running Suite: Injective/Cosmos OCR module E2E Test Suite
=========================================================
Random Seed: 1631772922
Will run 1 of 1 specs

OCR Feed Configs Proposals to set configs
  Submits Governance Proposals
  /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:50

• [SLOW TEST:6.410 seconds]
OCR Feed Configs
/Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:14
  Proposals to set configs
  /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:15
    Submits Governance Proposals
    /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:50
------------------------------

Ran 1 of 1 Specs in 6.414 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS

Ginkgo ran 1 suite in 15.404257292s
Test Suite Passed
```

### Running OCR2 oracle

Install the binary by running `make install`, it will make `injective-ocr2` available on your system, or at least in Go home bin. 

```
TODO: running OCR2 with node.
```
