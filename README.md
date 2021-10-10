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

The credentials for Chainlink Web UI are following:
```
username: test@test
password: test_test
```

### Step 4

Add external adapters bridges (or using Web UI / CLI) for all 4 oracles:

```bash
> ./test/test_init_bridges.sh

Adding Bridge 'injective-ea0' (http://host.docker.internal:8866) to Chainlink node
Bridge has been added to Chainlink node
Done adding Bridge 'injective-ea0'

Adding Bridge 'injective-ea1' (http://host.docker.internal:8867) to Chainlink node
Bridge has been added to Chainlink node
Done adding Bridge 'injective-ea1'

Adding Bridge 'injective-ea2' (http://host.docker.internal:8868) to Chainlink node
Bridge has been added to Chainlink node
Done adding Bridge 'injective-ea2'

Adding Bridge 'injective-ea3' (http://host.docker.internal:8869) to Chainlink node
Bridge has been added to Chainlink node
Done adding Bridge 'injective-ea3'
```

### Step 5

According to [Adding External Initiators to Nodes](https://docs.chain.link/docs/external-initiators-in-nodes/), let's add our OCR2 oracle as an external initiator to the Chainlink Node.
This can be achieved using a CLI tool, but there is a quicker script that uses the API. 

```bash
> ./test/test_init_eis.sh

Adding External Initiator 'injective-ei0' (http://host.docker.internal:8866) to Chainlink node...
EI has been added to Chainlink node
Done adding EI 'injective-ei0'

Adding External Initiator 'injective-ei1' (http://host.docker.internal:8867) to Chainlink node...
EI has been added to Chainlink node
Done adding EI 'injective-ei1'

Adding External Initiator 'injective-ei2' (http://host.docker.internal:8868) to Chainlink node...
EI has been added to Chainlink node
Done adding EI 'injective-ei2'

Adding External Initiator 'injective-ei3' (http://host.docker.internal:8869) to Chainlink node...
EI has been added to Chainlink node
Done adding EI 'injective-ei3'
```

It will save 4 generated files with CI/IC credentials into
```
./test/oracles/oracle0/external_initiator_injective-ei0.env
./test/oracles/oracle0/external_initiator_injective-ei1.env
./test/oracles/oracle0/external_initiator_injective-ei2.env
./test/oracles/oracle0/external_initiator_injective-ei3.env
```

Later they will be loaded by `test_oracles_start.sh` script.

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

Running Suite: Injective/Cosmos OCR module E2E Test Suite
=========================================================
Random Seed: 1633702647
Will run 1 of 1 specs

OCR Feed Configs Proposals to set configs
  Submits Governance Proposals and Funds Feed Reward Pool
  /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:81

• [SLOW TEST:18.322 seconds]
OCR Feed Configs
/Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:20
  Proposals to set configs
  /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:21
    Submits Governance Proposals and Funds Feed Reward Pool
    /Users/xlab/Documents/dev/InjectiveLabs/chainlink-injective/test/e2e/ocr_configs.go:81
------------------------------

Ran 1 of 1 Specs in 18.327 seconds
SUCCESS! -- 1 Passed | 0 Failed | 0 Pending | 0 Skipped
PASS
```

The created proposals will set a config for both feeds (LINK/USDC, INJ/USDC) and authrize our test oracles to sign and transmit reports onchain.

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

For example, get the feed config for `LINK/USDC` feed:

```bash
> injectived --home ./var/data/injective-1/n0 q ocr feed-config LINK/USDC
feed_config:
  f: 1
  offchain_config: CICg2eYdEIDkl9ASGIDkl9ASIIC8wZYLKIDkl9ASMP4BOgQBAQEBQiA1xYd9JqzdrbTZFe37XGakJ6PO2DKCkhWa/JCYDRRcXEIgyGOsc7xyDHmzTLBT2Bqb3yxwlPcxT/MubKbqdRnaIgpCILcyCNCyP4LCCxDvZZv/zqcTfkBM4xz0TPfmZWsGxuvSQiBqEloiNpBcFmFZd7bQsFmxmFfV+hDSUuhfn1gHigJHCko0MTJEM0tvb1dFb3k0S3JQM3V3ZDR1Wm1ERkJmS3VyMkY1elNOVFZNU3d5bVE5aU5DRnQ3Wko0MTJEM0tvb1dIZ29La3phTkdLWUszOVBNanlIM3RQQngxaURIbUVIenJCQ211S2huNEM4Rko0MTJEM0tvb1dKTFJYN04xYVAxWFNTN3ZIemlyZWVCY3M3bTlLdjMyMUZxWENDUGN3QjJQMko0MTJEM0tvb1dUMm1QYTVvbnFYR2tpY3ZhUVVIU1c2ZDZBVldjNUNMcXhNU1FUZlFDRGdjcVILCICt4gQQgMivoCVYgLzBlgtggLzBlgtogLzBlgtwgLzBlgt4gLzBlguCAYwBCiAkGQYBTVZaTi2WWvUFIFCFRck2LMFDifZbvwovn2wHVBIgGhv3U8JQXatSBLKE8f79fcTrg/YwRmTJwvQ8v0JWZmAaEA2f8S7UptzfYrOh2c0aHjMaEOOzcc00J0fGuSyqK01HlvYaECYmNaST/OH0LmJuDSfKD4IaEEDhjXNXatUM4pVXXUXVqZo=
  offchain_config_version: "2"
  onchain_config:
    billing_admin: ""
    chain_id: injective-1
    description: LINK/USDC Feed
    feed_admin: ""
    feed_id: LINK/USDC
    link_denom: peggy0x514910771AF9Ca656af840dff83E8264EcF986CA
    link_per_observation: "10"
    link_per_transmission: "69"
    max_answer: "99999999999999999.000000000000000000"
    min_answer: "0.000000000000000001"
    unique_reports: false
  signers:
  - inj1s4d8ygx4ej9k5wkge00uhcmdzd44udmfx98g78
  - inj1zm0y9tdptfxtkc86f3hsuhk74fx2j2sylyd57d
  - inj1u34x223x5y9fr3d09kyupycuqya8mlms2j5kua
  - inj1555f842w0jfdns23n0z466jtjdlhj6xv3c267k
  transmitters:
  - inj1s4d8ygx4ej9k5wkge00uhcmdzd44udmfx98g78
  - inj1zm0y9tdptfxtkc86f3hsuhk74fx2j2sylyd57d
  - inj1u34x223x5y9fr3d09kyupycuqya8mlms2j5kua
  - inj1555f842w0jfdns23n0z466jtjdlhj6xv3c267k
feed_config_info:
  config_count: "1"
  f: 1
  latest_config_block_number: "32"
  latest_config_digest: AAKhiy5LLi3Jon/OhLszF3M8RHltWiNgm8cSA0IQFac=
  "n": 4
```

We use 4 oracles there because `N=4 > F*3`.

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

Run ALL 4 oracle instances at once:

```bash
> ./test/test_oracles_start.sh

[start] running 4 oracles
[post-start]
Logs:
  * tail -f ./var/oracles/oracle0.log
  * tail -f ./var/oracles/oracle1.log
  * tail -f ./var/oracles/oracle2.log
  * tail -f ./var/oracles/oracle3.log

Stopping:
  * ./test/test_stop_oracles.sh
```

Monitor the logs of the first oracle (the bridge that will be used in Job spec):

```bash
> tail -f ./var/oracles/oracle0.log

time="2021-10-08T17:28:26+03:00" level=info msg="Using Cosmos Sender inj1s4d8ygx4ej9k5wkge00uhcmdzd44udmfx98g78"
time="2021-10-08T17:28:26+03:00" level=info msg="Waiting for GRPC services"
time="2021-10-08T17:28:28+03:00" level=info msg="Using PeerID 12D3KooWEoy4KrP3uwd4uZmDFBfKur2F5zSNTVMSwymQ9iNCFt7Z for P2P identity"
time="2021-10-08T17:28:28+03:00" level=info msg="Using OCR2 key ID 013208ee22ef424aa5d3a5abc3784459d8d72f6d602bbd19a94b626f8c9d932b"

[GIN-debug] GET    /health                   --> github.com/InjectiveLabs/chainlink-injective/api.handleShowHealth.func1 (3 handlers)
[GIN-debug] POST   /runs                     --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobRun.func1 (3 handlers)
[GIN-debug] POST   /jobs                     --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobCreate.func1 (4 handlers)
[GIN-debug] DELETE /jobs/:jobid              --> github.com/InjectiveLabs/chainlink-injective/api.(*httpServer).handleJobStop.func1 (4 handlers)
```

Double check that all 4 external initiators (`injective-ei*`) are actually registered within Chainlink Node, as they are not being displayed under Bridges tab in Web UI.

```bash
> docker exec -it test_chainlink-node_1 /bin/bash
#
# Inside test_chainlink-node_1 container instance:
> chainlink admin login --file /run/secrets/apicredentials
> chainlink initiators list

╔ External Initiators:
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬
║ ID ║     NAME      ║                  URL                  ║            ACCESSKEY             ║                          OUTGOINGTOKEN                           ║           CREATEDAT            ║           UPDATEDAT            ║
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬
║  1 ║ injective-ei0 ║ http://host.docker.internal:8866/jobs ║ 90c2142bfa0c4573bcff7955c32c8b17 ║ Y2HGB4rkqxMQvRoJbvkTVkkM1N2ni+4rS7JUEhP6ha9g1a0X6FCKH+4sV0TS1drP ║ 2021-10-08 12:25:10.89611      ║ 2021-10-08 12:25:10.89611      ║
║    ║               ║                                       ║                                  ║                                                                  ║ +0000 UTC                      ║ +0000 UTC                      ║
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬
║  2 ║ injective-ei1 ║ http://host.docker.internal:8867/jobs ║ 6161f044fa644d968752079b03111daa ║ I2zfNdIBfdxxraCX1LU+RV+uE5xCvgEE5/q/X1imbpzJp8192FGPyq/hbEJIa9Mk ║ 2021-10-08 12:25:11.131688     ║ 2021-10-08 12:25:11.131688     ║
║    ║               ║                                       ║                                  ║                                                                  ║ +0000 UTC                      ║ +0000 UTC                      ║
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬
║  3 ║ injective-ei2 ║ http://host.docker.internal:8868/jobs ║ 437fa2ca57364d2eb4f2790b8984f4db ║ MczBesfIBdGp/0UkjRjjtQ0Q3IdzUTxXRouYznPuFwPrqOizH4ErQ+gVcnd6ogRe ║ 2021-10-08 12:25:11.443282     ║ 2021-10-08 12:25:11.443282     ║
║    ║               ║                                       ║                                  ║                                                                  ║ +0000 UTC                      ║ +0000 UTC                      ║
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬
║  4 ║ injective-ei3 ║ http://host.docker.internal:8869/jobs ║ 8b80a95cb2f94c60a580bf287b8edde3 ║ FEkxJEDsGPoVy1gWJFpShZOdJIyI+SDSX1T9vYkITRKI3yeOJL2DbUuogSetOMUk ║ 2021-10-08 12:25:11.773275     ║ 2021-10-08 12:25:11.773275     ║
║    ║               ║                                       ║                                  ║                                                                  ║ +0000 UTC                      ║ +0000 UTC                      ║
╬════╬═══════════════╬═══════════════════════════════════════╬══════════════════════════════════╬══════════════════════════════════════════════════════════════════╬════════════════════════════════╬════════════════════════════════╬

> chainlink bridges list
╔ Bridges
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
║     NAME      ║                  URL                  ║ CONFIRMATIONS ║
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
║ injective-ea0 ║ http://host.docker.internal:8866/runs ║             0 ║
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
║ injective-ea1 ║ http://host.docker.internal:8867/runs ║             0 ║
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
║ injective-ea2 ║ http://host.docker.internal:8868/runs ║             0 ║
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
║ injective-ea3 ║ http://host.docker.internal:8869/runs ║             0 ║
╬═══════════════╬═══════════════════════════════════════╬═══════════════╬
```

It's time to schedule a new Job. There are multiple ways to create Jobs in Chainlink node, either via Web UI or API. We will use the following Job spec but adapted for each oracle, so there will be 4 jobs for `injective-ei{1-4}`/`injective-ea{1-4}` correspondingly:

```toml
type            = "webhook"
schemaVersion   = 1

externalInitiators = [
  { name = "injective-ei0", spec = "{\"feedId\": \"LINK/USDC\",\"p2pBootstrapPeers\": [\"12D3KooWEoy4KrP3uwd4uZmDFBfKur2F5zSNTVMSwymQ9iNCFt7Z@127.0.0.1:4466\"],\"isBootstrapPeer\": false,\"keyID\": \"013208ee22ef424aa5d3a5abc3784459d8d72f6d602bbd19a94b626f8c9d932b\",\"observationTimeout\": \"10s\",\"blockchainTimeout\": \"10s\",\"contractConfigConfirmations\": 1}" },
]

observationSource   = """
   ticker [type=http method=GET url="https://api.binance.com/api/v3/ticker/price?symbol=LINKUSDC"];
   parsePrice [type="jsonparse" path="price"]
   multiplyDecimals [type="multiply" times=1000000]
   sendToBridge [type=bridge name="injective-ea0" requestData=<{"jobID":$(jobSpec.externalJobID), "result":$(multiplyDecimals)}>]

   ticker -> parsePrice -> multiplyDecimals -> sendToBridge
"""
```

To add all 4 Jobs automatically:
```bash
> ./test/test_jobs_start.sh

Starting job 'job_linkusdc_ei0' (./test/jobs/job_linkusdc_ei0.toml) via Chainlink node
Job has been added via Chainlink node
Done adding Job 'job_linkusdc_ei0'

Starting job 'job_linkusdc_ei1' (./test/jobs/job_linkusdc_ei1.toml) via Chainlink node
Job has been added via Chainlink node
Done adding Job 'job_linkusdc_ei1'

Starting job 'job_linkusdc_ei2' (./test/jobs/job_linkusdc_ei2.toml) via Chainlink node
Job has been added via Chainlink node
Done adding Job 'job_linkusdc_ei2'

Starting job 'job_linkusdc_ei3' (./test/jobs/job_linkusdc_ei3.toml) via Chainlink node
Job has been added via Chainlink node
Done adding Job 'job_linkusdc_ei3'
```

Each oracle node will spawn its own OCR2 peer! They will be comunicating and the elected transmitters will send the median result to the chain. Make sure to check jobs logs:

```bash
Logs:
  * tail -f ./var/oracles/oracle0.log
  * tail -f ./var/oracles/oracle1.log
  * tail -f ./var/oracles/oracle2.log
  * tail -f ./var/oracles/oracle3.log
```

And you can query onchain state using on of the CLI subcommands on the chain client:

```bash
> injectived --home ./var/data/injective-1/n0 q ocr latest-transmission LINK/USDC
config_digest: AAKocozoxdRepNyjfzFr9pBpJoTzQ0IT3Wl1efeXa7E=
data:
  answer: "27480000.000000000000000000"
  observations_timestamp: "1633819742"
  transmission_timestamp: "1633819745"
epoch_and_round:
  epoch: "45"
  round: "1"

> injectived --home ./var/data/injective-1/n0 q ocr latest-transmission LINK/USDC
config_digest: AAKocozoxdRepNyjfzFr9pBpJoTzQ0IT3Wl1efeXa7E=
data:
  answer: "27550000.000000000000000000"
  observations_timestamp: "1633820296"
  transmission_timestamp: "1633820299"
epoch_and_round:
  epoch: "93"
  round: "1"
```

It works! 🎉
