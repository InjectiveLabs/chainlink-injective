package ocr2

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	"github.com/smartcontractkit/libocr/commontypes"
	ocr2 "github.com/smartcontractkit/libocr/offchainreporting2"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/InjectiveLabs/chainlink-injective/chainlink"
	"github.com/InjectiveLabs/chainlink-injective/db"
	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/injective/median_report"
	"github.com/InjectiveLabs/chainlink-injective/keys/ocrkey"
	"github.com/InjectiveLabs/chainlink-injective/keys/p2pkey"
	"github.com/InjectiveLabs/chainlink-injective/logging"
	"github.com/InjectiveLabs/chainlink-injective/p2p"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
)

type Job interface {
	median.DataSource

	Start() error
	Run(data string) error
	Stop() error
}

var _ Job = &job{}

type job struct {
	jobID   string
	jobSpec *model.JobSpec

	stateDB                JobStateDB
	client                 chainlink.WebhookClient
	transmitter            ocrtypes.ContractTransmitter
	medianReporter         median.MedianContract
	onchainKeyring         ocrtypes.OnchainKeyring
	configTracker          ocrtypes.ContractConfigTracker
	offchainConfigDigester ocrtypes.OffchainConfigDigester

	svc    ocr2Service
	p2pSvc p2pService

	runData chan *big.Int

	runningMux *sync.RWMutex
	running    bool
	onceStart  sync.Once
	onceStop   sync.Once

	logger log.Logger
}

type ocr2Service interface {
	Start() error
	Close() error
}

type p2pService interface {
	Start() error
	Close() error
}

func (s *jobService) newJob(
	jobID string,
	jobSpec *model.JobSpec,
	stateDB db.JobDBService,
	transmitter ocrtypes.ContractTransmitter,
	medianReporter median.MedianContract,
	onchainKeyring ocrtypes.OnchainKeyring,
	configTracker ocrtypes.ContractConfigTracker,
	offchainConfigDigester ocrtypes.OffchainConfigDigester,
) (Job, error) {
	j := &job{
		jobID:   jobID,
		jobSpec: jobSpec,

		stateDB: &jobDBWrapper{
			svc: stateDB,
		},
		client:                 s.client,
		transmitter:            transmitter,
		medianReporter:         medianReporter,
		onchainKeyring:         onchainKeyring,
		configTracker:          configTracker,
		offchainConfigDigester: offchainConfigDigester,

		runData: make(chan *big.Int),

		runningMux: new(sync.RWMutex),
		logger: log.WithFields(log.Fields{
			"svc":   "ocr2_job",
			"jobID": jobID,
		}),
	}

	if err := j.initOracleService(
		s.peerKey,
		s.peerNetworkingConfig,
		s.ocrKey,
		s.ocrConfig,
	); err != nil {
		return nil, err
	}

	return j, nil
}

func (j *job) initOracleService(
	peerKey p2pkey.Key,
	peerNetworkingConfig p2p.NetworkingConfig,
	ocrKey ocrkey.KeyV2,
	ocrConfig Config,
) error {
	if j.jobSpec.KeyID != model.ID(ocrKey.GetID()) {
		return errors.New("refusing to start Job with unexpected OCR2 Key")
	}

	p2pService, err := p2p.NewService(
		peerKey,
		peerNetworkingConfig,
		j.stateDB,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to init P2P Service")
		return err
	}
	j.p2pSvc = p2pService

	ocrLogger := logging.WrapCommonLogger(logging.NewSuplog(log.InfoLevel, false).WithField("svc", "ocr2_node"))

	chainTimeout, err := time.ParseDuration(j.jobSpec.BlockchainTimeout)
	if err != nil {
		return errors.Wrap(err, "incorrect job spec: missing blockchainTimeout")
	}

	configConfirmations := uint16(j.jobSpec.ContractConfigConfirmations)
	if configConfirmations == 0 {
		configConfirmations = 1
	}

	localConfig := ocrtypes.LocalConfig{
		BlockchainTimeout:                  chainTimeout,
		ContractConfigConfirmations:        configConfirmations,
		SkipContractConfigConfirmations:    false,
		ContractConfigTrackerPollInterval:  ocrConfig.ContractPollInterval,
		ContractTransmitterTransmitTimeout: ocrConfig.ContractTransmitterTransmitTimeout,
		DatabaseTimeout:                    ocrConfig.DatabaseTimeout,

		// Set this to EnableDangerousDevelopmentMode to turn on dev mode.
		DevelopmentMode: "",
	}

	var v2BootstrapPeers []commontypes.BootstrapperLocator
	for _, peer := range j.jobSpec.P2PBootstrapPeers {
		var bootstrapPeer commontypes.BootstrapperLocator
		if err := bootstrapPeer.UnmarshalText([]byte(peer)); err != nil {
			err = errors.Wrap(err, "failed to parse P2P bootstrap peers")
			return err
		}

		v2BootstrapPeers = append(v2BootstrapPeers, bootstrapPeer)
	}

	if err := p2pService.Start(); err != nil {
		err = errors.Wrap(err, "failed to start P2P service")
		return err
	}

	if j.jobSpec.IsBootstrapPeer {
		bootstrapArgs := ocr2.BootstrapperArgs{
			BootstrapperFactory:    p2pService.Peer(),
			V2Bootstrappers:        v2BootstrapPeers,
			ContractConfigTracker:  j.configTracker,
			Database:               j.stateDB,
			LocalConfig:            localConfig,
			Logger:                 ocrLogger,
			MonitoringEndpoint:     NewMonitor(),
			OffchainConfigDigester: j.offchainConfigDigester,
		}

		bootstrapNode, err := ocr2.NewBootstrapper(bootstrapArgs)
		if err != nil {
			err = errors.Wrap(err, "failed to init OCR2 bootstrap node")
			return err
		}

		j.svc = bootstrapNode
		return nil
	}

	numericalMedianFactory := median.NumericalMedianFactory{
		ContractTransmitter:   j.medianReporter,
		DataSource:            j, // reads from Observe() of this job
		JuelsPerEthDataSource: &dsZero{},
		Logger:                ocrLogger,
		ReportCodec:           median_report.ReportCodec{},
	}

	ocrArgs := ocr2.OracleArgs{
		BinaryNetworkEndpointFactory: p2pService.Peer(),
		V2Bootstrappers:              v2BootstrapPeers,
		ContractTransmitter:          j.transmitter,
		ContractConfigTracker:        j.configTracker,
		Database:                     j.stateDB,
		LocalConfig:                  localConfig,
		Logger:                       ocrLogger,
		MonitoringEndpoint:           NewMonitor(),
		OffchainConfigDigester:       j.offchainConfigDigester,
		OffchainKeyring:              ocrkey.NewOCR2KeyWrapper(ocrKey),
		OnchainKeyring:               j.onchainKeyring,
		ReportingPluginFactory:       numericalMedianFactory,
	}

	oracleNode, err := ocr2.NewOracle(ocrArgs)
	if err != nil {
		err = errors.Wrap(err, "failed to init OCR2 oracle node")
		return err
	}

	j.svc = oracleNode
	return nil
}

func (j *job) Start() (err error) {
	j.onceStart.Do(func() {
		j.logger.Infoln("Starting OCR2 Job")

		j.runningMux.Lock()
		defer j.runningMux.Unlock()
		j.running = true

		if err := j.svc.Start(); err != nil {
			err = errors.Wrap(err, "failed to start OCR2 service")
		}
	})

	return err
}

func (j *job) Run(data string) error {
	j.logger.Infoln("Run OCR2 Job")

	j.runningMux.RLock()
	defer j.runningMux.RUnlock()
	if !j.running {
		err := errors.New("job is not running")
		j.logger.WithError(err).Warningln("failed to run job")
		return err
	}

	observedValue, ok := new(big.Int).SetString(data, 10)
	if !ok {
		err := errors.Errorf("failed to parse job input %s as big.Int", data)
		j.logger.WithError(err).Warningln("failed to run job")
		return err
	}

	select {
	case j.runData <- observedValue:
	default:
	}

	return nil
}

func (j *job) Stop() (err error) {
	j.onceStart.Do(func() {
		j.logger.Infoln("Stopping OCR2 Job")

		j.runningMux.Lock()
		defer j.runningMux.Unlock()

		if !j.running {
			return
		}
		j.running = false

		if err := j.p2pSvc.Close(); err != nil {
			j.logger.WithError(err).Warningln("failed to stop P2P service")
		}

		if err := j.svc.Close(); err != nil {
			err = errors.Wrap(err, "failed to stop OCR2 service")
		}
	})

	return err
}

func (j *job) StateDB() JobStateDB {
	return j.stateDB
}

var (
	ErrJobStopped     = errors.New("job stopped")
	ErrObserveTimeout = errors.New("observation timed out")
)

// Observe queries the data source. Returns a value or an error. Once the
// context is expires, Observe may still do cheap computations and return a
// result, but should return as quickly as possible.
//
// More details: In the current implementation, the context passed to
// Observe will time out after LocalConfig.DataSourceTimeout. However,
// Observe should *not* make any assumptions about context timeout behavior.
// Once the context times out, Observe should prioritize returning as
// quickly as possible, but may still perform fast computations to return a
// result rather than error. For example, if Observe medianizes a number
// of data sources, some of which already returned a result to Observe prior
// to the context's expiry, Observe might still compute their median, and
// return it instead of an error.
//
// Important: Observe should not perform any potentially time-consuming
// actions like database access, once the context passed has expired.
func (j *job) Observe(ctx context.Context) (*big.Int, error) {
	j.logger.Infoln("Observe triggered")
	ts := time.Now()

	go func() {
		if err := j.client.TriggerJob(j.jobID); err != nil {
			j.logger.WithError(err).Errorln("failed to trigger Job on the Chainlink node")
		}
	}()

	select {
	case result, ok := <-j.runData:
		if !ok {
			j.logger.Warningln("Observe exits due to Job shutdown")
			return nil, ErrJobStopped
		}

		j.logger.WithField("data", result.String()).Infoln("Observation received in", time.Since(ts))
		return result, nil

	case <-ctx.Done():
		j.logger.WithError(ctx.Err()).Warningln("Observation timed out in", time.Since(ts))
		return nil, ErrObserveTimeout
	}
}
