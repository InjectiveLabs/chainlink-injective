# chainlink-injective

OCR2 median reporting plugin integration demo with libocr2 and Cosmos module, tailored for Injective Chain.

There is the instruction on running this oracle node.

## Prepare the environment

### Step 1

Clone https://github.com/InjectiveLabs/injective-core on branch `f/ocr` and build the chain node.
It will have the OCR chain module that is ready to accept transmissions. Installing the node is as simple as running `make install` in the corresponding repo.

### Step 2

Finally, clone https://github.com/InjectiveLabs/chainlink-injective (this repo on `master`) and let's run it. Make sure `injectived` executable is available on your system. Also [jq](https://stedolan.github.io/jq/) tool is required.

### Step 3

Install Docker. Run Chainlink Node and its Web UI

```bash
> docker-compose -f test/docker-compose.yml up -d

# check the logs of the node
> docker logs -f test_chainlink-node_1
```

### Step 4

Add external adapter bridge (or using Web UI / CLI):

```bash
> ./test/scripts/cl_add_bridge.sh injective-ea http://host.docker.internal:8866

Adding Bridge 'injective-ea' (http://host.docker.internal:8866) to Chainlink node
Bridge has been added to Chainlink node
Done adding Bridge 'injective-ea'
```

### Step 5

According to [Adding External Initiators to Nodes](https://docs.chain.link/docs/external-initiators-in-nodes/), let's add our OCR2 oracle as an external initiator to the Chainlink Node.
This can be achieved using a CLI tool, but there is a quicker script:

```bash
> ./test/scripts/cl_add_ei.sh injective http://host.docker.internal:8866

Adding External Initiator 'injective' (http://host.docker.internal:8866) to Chainlink node...
EI has been added to Chainlink node
Done adding EI 'injective'
```

Don't forget to copy the generated credentials to `.env` of chainlink-injective (see next steps):
```bash
> cat external_initiator_injective.env

EI_CI_ACCESSKEY=OFBujjXMeF1+uBTuUilZC4SHOExRoR6lEORa4bwusSHzp4jDGWFuzTj4ocC5GkXG
EI_CI_SECRET=TU4EAMy8AirTWPHaYa/KBN6Vao3XrJPsq+5FC79RZ2/TU7L1W3laNr5UbRQMIeQR
EI_IC_ACCESSKEY=670f81fa23b344e980993531c75474e4
EI_IC_SECRET=yM1ZMawqm8VcHv788Cin4QN10Sz4flPMlldyr9HROYpBBNRQWqxKSRogWo+E2TWM
```

Copy `.env.example` to `.env` and fill the Chainlink Node CI and IC secrets! Make sure `EI_CI_LISTEN_ADDR` is available and `EI_CHAINLINKURL` points to your Chainlink Node correctly.

```bash
> cat external_initiator_injective.env >> .env
```

### Running a network with 3-node consensus

Start a mock network of 3 injectived nodes with proper consensus, governance and some pre-baked accounts with balances.

```bash
> CLEANUP=1 ./test/e2e_multinode.sh injectived
```

Omit `CLEANUP=1` when running this command second time, it restarts the chain (all 3 nodes). If you include it again, it will wipe state, then restart. The last argument is the path to the node binary, can be arbitrary to `$GOHOME/bin/injectived` or where it was installed.

```bash
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

```bash
> make test
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

### Verifying on-chain state

After the chain consensus has been started and migrations completed, the module must contain 2 feeds ready to receive transmissions. To query the chain state of the OCR module from CLI, use the following commands:

```bash
> injectived --home ./var/data/injective-1/n0 q ocr

Usage:
  injectived query ocr [flags]
  injectived query ocr [command]

Available Commands:
  feed-config         Gets ocr feed config
  feed-config-info    Gets ocr feed config info
  latest-round        Gets ocr latest round by feed id.
  latest-transmission Gets ocr latest transmission details by feed id.
  module-state        Gets ocr module state.
  owed-amount         Gets owed amount by transmitter address.
  params              Gets ocr params
```

For example, get the feed config for `LINK/USDT` feed:

```bash
> injectived --home ./var/data/injective-1/n0 q ocr feed-config LINK/USDT
feed_config:
  f: 0
  offchain_config: e30=
  offchain_config_version: "2"
  onchain_config:
    billing_admin: ""
    chain_id: ""
    description: LINK/USDT Feed
    feed_admin: ""
    feed_id: LINK/USDT
    is_testing: true
    link_denom: peggy0x514910771AF9Ca656af840dff83E8264EcF986CA
    link_per_observation: "10"
    link_per_transmission: "69"
    max_answer: "99999999999999999.000000000000000000"
    min_answer: "0.000000000000000001"
    unique_reports: false
  signers:
  - inj128jwakuw3wrq6ye7m4p64wrzc5rfl8tvwzc6s8
  transmitters:
  - inj128jwakuw3wrq6ye7m4p64wrzc5rfl8tvwzc6s8
feed_config_info:
  config_count: "1"
  f: 0
  latest_config_block_number: "19"
  latest_config_digest: AAKuNguQkMLWL9HfNGV+InUggwbvg/tvTHwM1cS0gqI=
  "n": 1
```

## Running OCR2 oracle

Install the binary by running `make install`, it will make `injective-ocr2` available on your system, or at least in Go home bin. 

```bash
Usage: injective-ocr2 [OPTIONS] COMMAND [arg...]

Injective OCR2 compatible oracle and external adapter for Chainlink Node.

Options:
  -e, --env                The environment name this app runs in. Used for metrics and error reporting. (env $ORACLE_ENV) (default "local")
  -l, --log-level          Available levels: error, warn, info, debug. (env $ORACLE_LOG_LEVEL) (default "info")
      --svc-wait-timeout   Standard wait timeout for external services (e.g. Cosmos daemon GRPC connection) (env $ORACLE_SERVICE_WAIT_TIMEOUT) (default "1m")

Commands:
  start                    Starts the OCR2 service.
  keys                     Keys management.
  version                  Print the version information and exit.
```

Start a local MongoDB instance:

```bash
> make mongo
# to stop: make mongo-stop
```

Run the service itself:

```bash
> injective-ocr2 start
INFO[0000] Using Cosmos Sender inj128jwakuw3wrq6ye7m4p64wrzc5rfl8tvwzc6s8
INFO[0000] Waiting for GRPC services
INFO[0001] Using PeerID 12D3KooWPaHvunmPm3qjhsffgZBd2rQS4tdCgSYWEeRiX6hDsrdq for P2P identity
INFO[0001] Using OCR2 key ID f7b80d092a4c328ef52508d2cef17f4f31d16293729e19c62f9ad6cb59a961a0

[GIN-debug] GET    /health                   --> github.com/InjectiveLabs/chainlink-injective/api.handleShowHealth.func1 (3 handlers)
[GIN-debug] POST   /runs                     --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobRun.func1 (3 handlers)
[GIN-debug] POST   /jobs                     --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobCreate.func1 (4 handlers)
[GIN-debug] DELETE /jobs/:jobid              --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobStop.func1 (4 handlers)
```

Double check that the external initiator is actually registered within Chainlink Node, as they are not currently not being displayed under Bridges tab in Web UI.

```bash
> docker exec -it test_chainlink-node_1 /bin/bash
#
# Inside test_chainlink-node_1 container instance:
> chainlink admin login --file /run/secrets/apicredentials
> chainlink initiators list

╔ External Initiators:
╬════╬═══════════╬═══════════════════════════════════════╬═════════════════════════╬════════════════════════════════╬════════════════════════════╬
║ ID ║   NAME    ║                  URL                  ║        ACCESSKEY        ║          OUTGOINGTOKEN         ║           CREATEDAT        ║
╬════╬═══════════╬═══════════════════════════════════════╬═════════════════════════╬════════════════════════════════╬════════════════════════════╬
║  1 ║ injective ║ http://host.docker.internal:8866/jobs ║ 415401b........38988d85 ║ AJWIiIMB......utdYGFRSFOO9VmFd ║ 2021-10-07 17:13:50.021709 ║
║    ║           ║                                       ║                         ║                                ║ +0000 UTC                  ║
╬════╬═══════════╬═══════════════════════════════════════╬═════════════════════════╬════════════════════════════════╬════════════════════════════╬
```

Time to schedule a new Job. Go to Web UI at http://localhost:6688/jobs (u: `test@test` / p: `test_test`)

```toml
type            = "webhook"
schemaVersion   = 1
externalInitiators = [
  { name = "injective", spec = "{\"feedId\": \"LINK/USDT\",\"p2pBootstrapPeers\": [\"16Uiu2HAm58SP7UL8zsnpeuwHfytLocaqgnyaYKP8wu7qRdrixLju@chain.link:1234\"],\"isBootstrapPeer\": false,\"keyID\": \"f7b80d092a4c328ef52508d2cef17f4f31d16293729e19c62f9ad6cb59a961a0\",\"observationTimeout\": \"10s\",\"blockchainTimeout\": \"10s\",\"contractConfigConfirmations\": 1}" }
]
observationSource   = """
   ticker [type=http method=GET url="https://api.binance.com/api/v3/ticker/price?symbol=LINKUSDT"];
   parse_rice [type="jsonparse" path="price"]
   multiply_decimals [type="multiply" times=1000000]
   send_to_bridge [type=bridge name="injective-ea" requestData=<{"jobID":$(jobSpec.externalJobID), "result":$(multiplyDecimals)}>]

   ticker -> parse_rice -> multiply_decimals -> send_to_bridge
"""
```