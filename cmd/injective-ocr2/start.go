package main

import (
	"context"
	"os"
	"time"

	cli "github.com/jawher/mow.cli"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/libocr/commontypes"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	"github.com/xlab/closer"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chainlink-injective/api"
	"github.com/InjectiveLabs/chainlink-injective/chainlink"
	"github.com/InjectiveLabs/chainlink-injective/db"
	"github.com/InjectiveLabs/chainlink-injective/db/dbconn"
	ocrtypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/ocr2"
	"github.com/InjectiveLabs/chainlink-injective/p2p"
	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
)

// startCmd action runs the service
//
// $ injective-ocr2 start
func startCmd(cmd *cli.Cmd) {
	var (
		// Cosmos params
		cosmosChainID   *string
		cosmosGRPC      *string
		tendermintRPC   *string
		cosmosGasPrices *string

		// Cosmos Key Management
		cosmosKeyringDir     *string
		cosmosKeyringAppName *string
		cosmosKeyringBackend *string
		cosmosKeyFrom        *string
		cosmosKeyPassphrase  *string
		cosmosPrivKey        *string
		cosmosUseLedger      *bool

		ocrKeyringDir    *string
		ocrKeyID         *string
		ocrKeyPassphrase *string
		ocrPrivKey       *string

		p2pKeyringDir    *string
		p2pPeerID        *string
		p2pKeyPassphrase *string
		p2pPrivKey       *string

		p2pDHTLookupInterval         *string
		p2pIncomingMessageBufferSize *int
		p2pOutgoingMessageBufferSize *int
		p2pNewStreamTimeout          *string
		p2pBootstrapCheckInterval    *string
		p2pTraceLogging              *bool
		p2pV2AnnounceAddresses       *[]string
		p2pV2Bootstrappers           *[]string
		p2pV2DeltaDial               *string
		p2pV2DeltaReconcile          *string
		p2pV2ListenAddresses         *[]string

		dbMongoConnection *string
		dbMongoDBName     *string

		eiChainlinkURL *string
		eiAccessKeyIC  *string
		eiSecretIC     *string
		eiAccessKeyCI  *string
		eiSecretCI     *string
		eiListenAddrCI *string

		// Metrics
		statsdPrefix   *string
		statsdAddr     *string
		statsdStuckDur *string
		statsdMocking  *string
		statsdDisabled *string
	)

	initCosmosOptions(
		cmd,
		&cosmosChainID,
		&cosmosGRPC,
		&tendermintRPC,
		&cosmosGasPrices,
	)

	initCosmosKeyOptions(
		cmd,
		&cosmosKeyringDir,
		&cosmosKeyringAppName,
		&cosmosKeyringBackend,
		&cosmosKeyFrom,
		&cosmosKeyPassphrase,
		&cosmosPrivKey,
		&cosmosUseLedger,
	)

	initOCRKeyOptions(
		cmd,
		&ocrKeyringDir,
		&ocrKeyID,
		&ocrKeyPassphrase,
		&ocrPrivKey,
	)

	initP2PKeyOptions(
		cmd,
		&p2pKeyringDir,
		&p2pPeerID,
		&p2pKeyPassphrase,
		&p2pPrivKey,
	)

	initP2PNetworkOptions(
		cmd,
		&p2pDHTLookupInterval,
		&p2pIncomingMessageBufferSize,
		&p2pOutgoingMessageBufferSize,
		&p2pNewStreamTimeout,
		&p2pBootstrapCheckInterval,
		&p2pTraceLogging,
		&p2pV2AnnounceAddresses,
		&p2pV2Bootstrappers,
		&p2pV2DeltaDial,
		&p2pV2DeltaReconcile,
		&p2pV2ListenAddresses,
	)

	initDBOptions(
		cmd,
		&dbMongoConnection,
		&dbMongoDBName,
	)

	initChainlinkOptions(
		cmd,
		&eiChainlinkURL,
		&eiAccessKeyIC,
		&eiSecretIC,
		&eiAccessKeyCI,
		&eiSecretCI,
		&eiListenAddrCI,
	)

	initStatsdOptions(
		cmd,
		&statsdPrefix,
		&statsdAddr,
		&statsdStuckDur,
		&statsdMocking,
		&statsdDisabled,
	)

	cmd.Action = func() {
		// ensure a clean exit
		defer closer.Close()

		startMetricsGathering(
			statsdPrefix,
			statsdAddr,
			statsdStuckDur,
			statsdMocking,
			statsdDisabled,
		)

		if *cosmosUseLedger {
			log.Fatalln("cannot really use Ledger for oracle service loop, since signatures msut be realtime")
		}

		senderAddress, cosmosKeyring, err := initCosmosKeyring(
			cosmosKeyringDir,
			cosmosKeyringAppName,
			cosmosKeyringBackend,
			cosmosKeyFrom,
			cosmosKeyPassphrase,
			cosmosPrivKey,
			cosmosUseLedger,
		)
		if err != nil {
			log.WithError(err).Fatalln("failed to init Cosmos keyring")
		}

		log.Infoln("Using Cosmos Sender", senderAddress.String())

		clientCtx, err := chainclient.NewClientContext(*cosmosChainID, senderAddress.String(), cosmosKeyring)
		if err != nil {
			log.WithError(err).Fatalln("failed to initialize cosmos client context")
		}
		clientCtx = clientCtx.WithNodeURI(*tendermintRPC)
		tmRPC, err := rpchttp.New(*tendermintRPC, "/websocket")
		if err != nil {
			log.WithError(err).Fatalln("failed to connect to tendermint RPC")
		}
		clientCtx = clientCtx.WithClient(tmRPC)

		cosmosClient, err := chainclient.NewCosmosClient(clientCtx, *cosmosGRPC, chainclient.OptionGasPrices(*cosmosGasPrices))
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"endpoint": *cosmosGRPC,
			}).Fatalln("failed to connect to daemon, is injectived running?")
		}
		closer.Bind(func() {
			cosmosClient.Close()
		})

		log.Infoln("Waiting for GRPC services")
		time.Sleep(1 * time.Second)

		daemonWaitCtx, cancelWait := context.WithTimeout(context.Background(), time.Minute)
		daemonConn := cosmosClient.QueryClient()
		waitForService(daemonWaitCtx, daemonConn)
		cancelWait()

		// Setup MongoDB database connection
		//

		dbConnContext, cancelFn := context.WithTimeout(context.Background(), 20*time.Second)
		dbConn, err := dbconn.NewMongoConn(dbConnContext, &dbconn.MongoConfig{
			Connection: *dbMongoConnection,
			Database:   *dbMongoDBName,
		})
		if err != nil {
			log.WithError(err).Fatalln("failed to init MongoDB client")
		} else if err := dbConn.TestConn(dbConnContext); err != nil {
			log.Fatalln(err)
		}
		cancelFn()
		closer.Bind(func() {
			if err := dbConn.Close(); err != nil {
				log.WithError(err).Warningln("failed to close MongoDB connection")
			}
		})

		dbSvc, err := db.NewDBService(dbConn)
		if err != nil {
			err = errors.Wrap(err, "failed to init DB service")
			log.Fatalln(err)
		}
		closer.Bind(func() {
			dbSvc.Close()
		})

		// Init Chainlink Node Webhook client
		//

		webhookClient := chainlink.NewWebhookClient(
			*eiChainlinkURL,
			*eiAccessKeyIC,
			*eiSecretIC,
		)

		// Parse P2P Network options and identity
		//

		p2pNetworkConfig, err := parseP2PNetworkOptions(
			p2pDHTLookupInterval,
			p2pIncomingMessageBufferSize,
			p2pOutgoingMessageBufferSize,
			p2pNewStreamTimeout,
			p2pBootstrapCheckInterval,
			p2pTraceLogging,
			p2pV2AnnounceAddresses,
			p2pV2Bootstrappers,
			p2pV2DeltaDial,
			p2pV2DeltaReconcile,
			p2pV2ListenAddresses,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to parse P2P Networking options")
			log.Fatalln(err)
		}

		peerID, peerKey, err := initP2PKey(
			p2pKeyringDir,
			p2pPeerID,
			p2pKeyPassphrase,
			p2pPrivKey,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to load P2P Peer key")
			log.Fatalln(err)
		}

		log.Infoln("Using PeerID %s for P2P identity", peer.ID(peerID).Pretty())

		// Load OCR2 key fron the keystore
		//

		ocrKeyID, ocrKey, err := initOCRKey(
			ocrKeyringDir,
			ocrKeyID,
			ocrKeyPassphrase,
			ocrPrivKey,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to load P2P Peer key")
			log.Fatalln(err)
		}

		log.Infoln("Using OCR2 key ID", ocrKeyID)

		// Start the Job service (the main OCR2 jobs dispatcher)
		//

		jobSvc := ocr2.NewJobService(
			dbSvc,
			webhookClient,
			peerKey,
			p2pNetworkConfig,
			ocrKey,
			*cosmosChainID,
			ocrtypes.NewQueryClient(daemonConn),
			cosmosClient,
			senderAddress,
			cosmosKeyring,
		)
		closer.Bind(func() {
			jobSvc.Close()
		})

		apiCredentials := api.AuthCredentials{
			AccessKey: *eiAccessKeyCI,
			Secret:    *eiSecretCI,
		}

		apiSrv, err := api.NewServer(
			apiCredentials,
			jobSvc,
		)

		go func() {
			if err := apiSrv.ListenAndServe(*eiListenAddrCI); err != nil {
				log.Errorln(err)

				// signal there that the app has failed
				os.Exit(1)
			}
		}()

		closer.Hold()
	}
}

func parseP2PNetworkOptions(
	p2pDHTLookupInterval *string,
	p2pIncomingMessageBufferSize *int,
	p2pOutgoingMessageBufferSize *int,
	p2pNewStreamTimeout *string,
	p2pBootstrapCheckInterval *string,
	p2pTraceLogging *bool,
	p2pV2AnnounceAddresses *[]string,
	p2pV2Bootstrappers *[]string,
	p2pV2DeltaDial *string,
	p2pV2DeltaReconcile *string,
	p2pV2ListenAddresses *[]string,
) (cfg p2p.NetworkingConfig, err error) {
	if len(*p2pDHTLookupInterval) > 0 {
		var interval time.Duration
		interval, err = time.ParseDuration(*p2pDHTLookupInterval)
		if err != nil {
			err = errors.Wrap(err, "failed to parse duration p2pDHTLookupInterval")
			return
		}

		cfg.DHTLookupInterval = int(interval / time.Second)
		if cfg.DHTLookupInterval == 0 {
			cfg.DHTLookupInterval = 10
		}
	}

	if *p2pIncomingMessageBufferSize > 0 {
		cfg.IncomingMessageBufferSize = *p2pIncomingMessageBufferSize
	}

	if *p2pOutgoingMessageBufferSize > 0 {
		cfg.OutgoingMessageBufferSize = *p2pOutgoingMessageBufferSize
	}

	if len(*p2pNewStreamTimeout) > 0 {
		cfg.NewStreamTimeout, err = time.ParseDuration(*p2pNewStreamTimeout)
		if err != nil {
			err = errors.Wrap(err, "failed to parse duration p2pNewStreamTimeout")
			return
		}
	}

	if len(*p2pBootstrapCheckInterval) > 0 {
		cfg.BootstrapCheckInterval, err = time.ParseDuration(*p2pBootstrapCheckInterval)
		if err != nil {
			err = errors.Wrap(err, "failed to parse duration p2pBootstrapCheckInterval")
			return
		}
	}

	if *p2pTraceLogging {
		cfg.TraceLogging = true
	}

	if len(*p2pV2AnnounceAddresses) > 0 {
		for _, addr := range *p2pV2AnnounceAddresses {
			cfg.P2PV2AnnounceAddresses = append(cfg.P2PV2AnnounceAddresses, addr)
		}
	}

	if len(*p2pV2Bootstrappers) > 0 {
		for idx, bootstrapperSpec := range *p2pV2Bootstrappers {
			var locator commontypes.BootstrapperLocator

			if err = locator.UnmarshalText([]byte(bootstrapperSpec)); err != nil {
				err = errors.Wrapf(err, "failed to parse %d-th elem of p2pV2Bootstrappers", idx)
				return
			}

			cfg.P2PV2Bootstrappers = append(cfg.P2PV2Bootstrappers, locator)
		}
	}

	if len(*p2pV2DeltaDial) > 0 {
		cfg.P2PV2DeltaDial, err = time.ParseDuration(*p2pV2DeltaDial)
		if err != nil {
			err = errors.Wrap(err, "failed to parse duration p2pV2DeltaDial")
			return
		}
	}

	if len(*p2pV2DeltaReconcile) > 0 {
		cfg.P2PV2DeltaReconcile, err = time.ParseDuration(*p2pV2DeltaReconcile)
		if err != nil {
			err = errors.Wrap(err, "failed to parse duration p2pV2DeltaReconcile")
			return
		}
	}

	if len(*p2pV2ListenAddresses) > 0 {
		for _, addr := range *p2pV2ListenAddresses {
			cfg.P2PV2ListenAddresses = append(cfg.P2PV2ListenAddresses, addr)
		}
	}

	return cfg, err
}
