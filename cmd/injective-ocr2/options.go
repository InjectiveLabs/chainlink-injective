package main

import cli "github.com/jawher/mow.cli"

// initGlobalOptions defines some global CLI options, that are useful for most parts of the app.
// Before adding option to there, consider moving it into the actual Cmd.
func initGlobalOptions(
	envName **string,
	appLogLevel **string,
	svcWaitTimeout **string,
) {
	*envName = app.String(cli.StringOpt{
		Name:   "e env",
		Desc:   "The environment name this app runs in. Used for metrics and error reporting.",
		EnvVar: "ORACLE_ENV",
		Value:  "local",
	})

	*appLogLevel = app.String(cli.StringOpt{
		Name:   "l log-level",
		Desc:   "Available levels: error, warn, info, debug.",
		EnvVar: "ORACLE_LOG_LEVEL",
		Value:  "info",
	})

	*svcWaitTimeout = app.String(cli.StringOpt{
		Name:   "svc-wait-timeout",
		Desc:   "Standard wait timeout for external services (e.g. Cosmos daemon GRPC connection)",
		EnvVar: "ORACLE_SERVICE_WAIT_TIMEOUT",
		Value:  "1m",
	})
}

func initChainlinkOptions(
	cmd *cli.Cmd,
	eiChainlinkURL **string,
	eiAccessKeyIC **string,
	eiSecretIC **string,
	eiAccessKeyCI **string,
	eiSecretCI **string,
	eiListenAddrCI **string,
) {
	*eiChainlinkURL = cmd.String(cli.StringOpt{
		Name:   "ei-chainlink-url",
		Desc:   "Chainlink Node URL to use for the external initiator",
		EnvVar: "EI_CHAINLINKURL",
		Value:  "http://localhost:6688",
	})

	*eiAccessKeyIC = cmd.String(cli.StringOpt{
		Name:   "ei-ic-accesskey",
		Desc:   "External Initiator access key for Initiator->Chainlink calls.",
		EnvVar: "EI_IC_ACCESSKEY",
		Value:  "",
	})

	*eiSecretIC = cmd.String(cli.StringOpt{
		Name:   "ei-ic-secret",
		Desc:   "External Initiator secret for Initiator->Chainlink calls.",
		EnvVar: "EI_IC_SECRET",
		Value:  "",
	})

	*eiAccessKeyCI = cmd.String(cli.StringOpt{
		Name:   "ei-ci-accesskey",
		Desc:   "External Initiator access key for Chainlink->Initiator calls.",
		EnvVar: "EI_CI_ACCESSKEY",
		Value:  "",
	})

	*eiListenAddrCI = cmd.String(cli.StringOpt{
		Name:   "ei-ci-listen",
		Desc:   "External Initiator listen address for Chainlink->Initiator calls.",
		EnvVar: "EI_CI_LISTEN_ADDR",
		Value:  "localhost:8866", // ha!
	})
}

func initP2PNetworkOptions(
	cmd *cli.Cmd,
	p2pDHTLookupInterval **string,
	p2pIncomingMessageBufferSize **int,
	p2pOutgoingMessageBufferSize **int,
	p2pNewStreamTimeout **string,
	p2pBootstrapCheckInterval **string,
	p2pTraceLogging **bool,
	p2pV2AnnounceAddresses **[]string,
	p2pV2Bootstrappers **[]string,
	p2pV2DeltaDial **string,
	p2pV2DeltaReconcile **string,
	p2pV2ListenAddresses **[]string,
) {
	*p2pDHTLookupInterval = cmd.String(cli.StringOpt{
		Name:   "p2p-dht-lookup-interval",
		Desc:   "Specify P2P DHT Lookup interval.",
		EnvVar: "P2P_DHT_LOOKUP_INTERVAL",
		Value:  "10s",
	})

	*p2pIncomingMessageBufferSize = cmd.Int(cli.IntOpt{
		Name:   "p2p-incoming-message-buffer-size",
		Desc:   "Specify P2P Incoming Message buffer size in bytes.",
		EnvVar: "P2P_INCOMING_MESSAGE_BUFFER_SIZE",
		Value:  10,
	})

	*p2pOutgoingMessageBufferSize = cmd.Int(cli.IntOpt{
		Name:   "p2p-outgoing-message-buffer-size",
		Desc:   "Specify P2P Outgoing Message buffer size in bytes.",
		EnvVar: "P2P_OUTGOING_MESSAGE_BUFFER_SIZE",
		Value:  10,
	})

	*p2pNewStreamTimeout = cmd.String(cli.StringOpt{
		Name:   "p2p-new-stream-timeout",
		Desc:   "Specify P2P New Stream timeout.",
		EnvVar: "P2P_NEW_STREAM_TIMEOUT",
		Value:  "10s",
	})

	*p2pBootstrapCheckInterval = cmd.String(cli.StringOpt{
		Name:   "p2p-bootstrap-check-interval",
		Desc:   "Specify P2P Bootstrap Check interval.",
		EnvVar: "P2P_BOOTSTRAP_CHECK_INTERVAL",
		Value:  "20s",
	})

	*p2pTraceLogging = cmd.Bool(cli.BoolOpt{
		Name:   "p2p-trace-logging",
		Desc:   "Specify P2P Trace Logging.",
		EnvVar: "P2P_TRACE_LOGGING",
		Value:  false,
	})

	*p2pV2AnnounceAddresses = cmd.Strings(cli.StringsOpt{
		Name:   "p2p-v2-announce-addresses",
		Desc:   "Specify P2P V2 Announce Addresses.",
		EnvVar: "P2P_V2_ANNOUNCE_ADDRESSES",
		Value:  []string{},
	})

	*p2pV2Bootstrappers = cmd.Strings(cli.StringsOpt{
		Name:   "p2p-v2-bootstrappers",
		Desc:   "Specify P2P V2 Bootstrappers.",
		EnvVar: "P2P_V2_BOOTSTRAPPERS",
		Value:  []string{},
	})

	*p2pV2DeltaDial = cmd.String(cli.StringOpt{
		Name:   "p2p-v2-delta-dial",
		Desc:   "Specify P2P V2 Delta Dial duration.",
		EnvVar: "P2P_V2_DELTA_DIAL",
		Value:  "15s",
	})

	*p2pV2DeltaReconcile = cmd.String(cli.StringOpt{
		Name:   "p2p-v2-delta-reconcile",
		Desc:   "Specify P2P V2 Delta Reconcile duration.",
		EnvVar: "P2P_V2_DELTA_RECONCILE",
		Value:  "1m",
	})

	*p2pV2ListenAddresses = cmd.Strings(cli.StringsOpt{
		Name:   "p2p-v2-listen-addresses",
		Desc:   "Specify P2P V2 Listen Addresses.",
		EnvVar: "P2P_V2_LISTEN_ADDRESSES",
		Value:  []string{},
	})
}

func initCosmosOptions(
	cmd *cli.Cmd,
	cosmosChainID **string,
	cosmosGRPC **string,
	tendermintRPC **string,
	cosmosGasPrices **string,
) {
	*cosmosChainID = cmd.String(cli.StringOpt{
		Name:   "cosmos-chain-id",
		Desc:   "Specify Chain ID of the Cosmos network.",
		EnvVar: "ORACLE_COSMOS_CHAIN_ID",
		Value:  "injective-1",
	})

	*cosmosGRPC = cmd.String(cli.StringOpt{
		Name:   "cosmos-grpc",
		Desc:   "Cosmos GRPC querying endpoint",
		EnvVar: "ORACLE_COSMOS_GRPC",
		Value:  "tcp://localhost:9900",
	})

	*tendermintRPC = cmd.String(cli.StringOpt{
		Name:   "tendermint-rpc",
		Desc:   "Tendermint RPC endpoint",
		EnvVar: "ORACLE_TENDERMINT_RPC",
		Value:  "http://localhost:26657",
	})

	*cosmosGasPrices = cmd.String(cli.StringOpt{
		Name:   "cosmos-gas-prices",
		Desc:   "Specify Cosmos chain transaction fees as sdk.Coins gas prices",
		EnvVar: "ORACLE_COSMOS_GAS_PRICES",
		Value:  "", // example: 500000000inj
	})
}

func initCosmosKeyOptions(
	cmd *cli.Cmd,
	cosmosKeyringDir **string,
	cosmosKeyringAppName **string,
	cosmosKeyringBackend **string,
	cosmosKeyFrom **string,
	cosmosKeyPassphrase **string,
	cosmosPrivKey **string,
	cosmosUseLedger **bool,
) {
	*cosmosKeyringBackend = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring",
		Desc:   "Specify Cosmos keyring backend (os|file|kwallet|pass|test)",
		EnvVar: "ORACLE_COSMOS_KEYRING",
		Value:  "file",
	})

	*cosmosKeyringDir = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring-dir",
		Desc:   "Specify Cosmos keyring dir, if using file keyring.",
		EnvVar: "ORACLE_COSMOS_KEYRING_DIR",
		Value:  "",
	})

	*cosmosKeyringAppName = cmd.String(cli.StringOpt{
		Name:   "cosmos-keyring-app",
		Desc:   "Specify Cosmos keyring app name.",
		EnvVar: "ORACLE_COSMOS_KEYRING_APP",
		Value:  "injectived",
	})

	*cosmosKeyFrom = cmd.String(cli.StringOpt{
		Name:   "cosmos-from",
		Desc:   "Specify the Cosmos validator key name or address. If specified, must exist in keyring, ledger or match the privkey.",
		EnvVar: "ORACLE_COSMOS_FROM",
	})

	*cosmosKeyPassphrase = cmd.String(cli.StringOpt{
		Name:   "cosmos-from-passphrase",
		Desc:   "Specify keyring passphrase, otherwise Stdin will be used.",
		EnvVar: "ORACLE_COSMOS_FROM_PASSPHRASE",
	})

	*cosmosPrivKey = cmd.String(cli.StringOpt{
		Name:   "cosmos-pk",
		Desc:   "Provide a raw Cosmos account private key of the validator in hex. USE FOR TESTING ONLY!",
		EnvVar: "ORACLE_COSMOS_PK",
	})

	*cosmosUseLedger = cmd.Bool(cli.BoolOpt{
		Name:   "cosmos-use-ledger",
		Desc:   "Use the Cosmos app on hardware ledger to sign transactions.",
		EnvVar: "ORACLE_COSMOS_USE_LEDGER",
		Value:  false,
	})
}

func initOCRKeyOptions(
	cmd *cli.Cmd,
	ocrKeyringDir **string,
	ocrKeyID **string,
	ocrKeyPassphrase **string,
	ocrPrivKey **string,
) {
	*ocrKeyringDir = cmd.String(cli.StringOpt{
		Name:   "ocr-keyring-dir",
		Desc:   "Specify OCR keyring dir to search for keys.",
		EnvVar: "ORACLE_OCR_KEYRING_DIR",
		Value:  "",
	})

	*ocrKeyID = cmd.String(cli.StringOpt{
		Name:   "ocr-key-id",
		Desc:   "Specify the OCR Key ID. If specified, must exist in keyring, or match the privkey.",
		EnvVar: "ORACLE_OCR_KEY_ID",
	})

	*ocrKeyPassphrase = cmd.String(cli.StringOpt{
		Name:   "ocr-key-passphrase",
		Desc:   "Specify OCR key passphrase.",
		EnvVar: "ORACLE_OCR_KEY_PASSPHRASE",
	})

	*ocrPrivKey = cmd.String(cli.StringOpt{
		Name:   "ocr-pk",
		Desc:   "Provide a raw OCR private key (Ed25519) in hex. USE FOR TESTING ONLY!",
		EnvVar: "ORACLE_OCR_PK",
	})
}

func initP2PKeyOptions(
	cmd *cli.Cmd,
	p2pKeyringDir **string,
	p2pPeerID **string,
	p2pKeyPassphrase **string,
	p2pPrivKey **string,
) {
	*p2pKeyringDir = cmd.String(cli.StringOpt{
		Name:   "p2p-keyring-dir",
		Desc:   "Specify P2P keyring dir to search for keys.",
		EnvVar: "ORACLE_P2P_KEYRING_DIR",
		Value:  "",
	})

	*p2pPeerID = cmd.String(cli.StringOpt{
		Name:   "p2p-peer-id",
		Desc:   "Specify the P2P Peer ID. If specified, must exist in keyring, or match the privkey.",
		EnvVar: "ORACLE_P2P_PEER_ID",
	})

	*p2pKeyPassphrase = cmd.String(cli.StringOpt{
		Name:   "p2p-key-passphrase",
		Desc:   "Specify P2P key passphrase.",
		EnvVar: "ORACLE_P2P_KEY_PASSPHRASE",
	})

	*p2pPrivKey = cmd.String(cli.StringOpt{
		Name:   "p2p-pk",
		Desc:   "Provide a raw P2P private key (libp2p Ed25519) in hex. USE FOR TESTING ONLY!",
		EnvVar: "ORACLE_P2P_PK",
	})
}

func initDBOptions(
	c *cli.Cmd,
	dbMongoConnection **string,
	dbMongoDBName **string,
) {
	*dbMongoConnection = c.String(cli.StringOpt{
		Name:   "M mongo-connection",
		Desc:   "Specify MongoDB connection string.",
		EnvVar: "ORACLE_DB_MONGO_CONNECTION",
		Value:  "mongodb://127.0.0.1:27017",
	})

	*dbMongoDBName = c.String(cli.StringOpt{
		Name:   "mongo-db-name",
		Desc:   "Specify MongoDB database name.",
		EnvVar: "ORACLE_DB_MONGO_DBNAME",
		Value:  "ocr2",
	})
}

// initStatsdOptions sets options for StatsD metrics.
func initStatsdOptions(
	cmd *cli.Cmd,
	statsdPrefix **string,
	statsdAddr **string,
	statsdStuckDur **string,
	statsdMocking **string,
	statsdDisabled **string,
) {
	*statsdPrefix = cmd.String(cli.StringOpt{
		Name:   "statsd-prefix",
		Desc:   "Specify StatsD compatible metrics prefix.",
		EnvVar: "ORACLE_STATSD_PREFIX",
		Value:  "oracle",
	})

	*statsdAddr = cmd.String(cli.StringOpt{
		Name:   "statsd-addr",
		Desc:   "UDP address of a StatsD compatible metrics aggregator.",
		EnvVar: "ORACLE_STATSD_ADDR",
		Value:  "localhost:8125",
	})

	*statsdStuckDur = cmd.String(cli.StringOpt{
		Name:   "statsd-stuck-func",
		Desc:   "Sets a duration to consider a function to be stuck (e.g. in deadlock).",
		EnvVar: "ORACLE_STATSD_STUCK_DUR",
		Value:  "5m",
	})

	*statsdMocking = cmd.String(cli.StringOpt{
		Name:   "statsd-mocking",
		Desc:   "If enabled replaces statsd client with a mock one that simply logs values.",
		EnvVar: "ORACLE_STATSD_MOCKING",
		Value:  "false",
	})

	*statsdDisabled = cmd.String(cli.StringOpt{
		Name:   "statsd-disabled",
		Desc:   "Force disabling statsd reporting completely.",
		EnvVar: "ORACLE_STATSD_DISABLED",
		Value:  "true",
	})
}
