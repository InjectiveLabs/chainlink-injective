package ocr2

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	chainclient "github.com/InjectiveLabs/sdk-go/chain/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/InjectiveLabs/chainlink-injective/chainlink"
	"github.com/InjectiveLabs/chainlink-injective/db"
	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/injective"
	"github.com/InjectiveLabs/chainlink-injective/injective/tmclient"
	chaintypes "github.com/InjectiveLabs/chainlink-injective/injective/types"
	"github.com/InjectiveLabs/chainlink-injective/keys/ocrkey"
	"github.com/InjectiveLabs/chainlink-injective/keys/p2pkey"
	"github.com/InjectiveLabs/chainlink-injective/p2p"
)

type JobService interface {
	StartJob(jobID string, spec *model.JobSpec) error
	RunJob(jobID, result string) error
	StopJob(jobID string) error
	Close() error
}

var _ JobService = &jobService{}

type jobService struct {
	dbSvc  db.DBService
	client chainlink.WebhookClient

	peerKey              p2pkey.Key
	peerNetworkingConfig p2p.NetworkingConfig
	ocrKey               ocrkey.KeyV2
	ocrConfig            Config

	chainID          string
	chainQueryClient chaintypes.QueryClient
	cosmosClient     chainclient.CosmosClient
	tmClient         tmclient.TendermintClient
	onchainSigner    sdk.AccAddress
	cosmosKeyring    keyring.Keyring

	activeJobsMux *sync.RWMutex
	activeJobs    map[string]Job

	runningMux *sync.RWMutex
	running    bool
	onceStart  sync.Once
	onceStop   sync.Once

	logger log.Logger
}

func NewJobService(
	dbSvc db.DBService,
	client chainlink.WebhookClient,
	peerKey p2pkey.Key,
	peerNetworkingConfig p2p.NetworkingConfig,
	ocrKey ocrkey.KeyV2,
	chainID string,
	chainQueryClient chaintypes.QueryClient,
	cosmosClient chainclient.CosmosClient,
	tmClient tmclient.TendermintClient,
	onchainSigner sdk.AccAddress,
	cosmosKeyring keyring.Keyring,
) JobService {
	j := &jobService{
		dbSvc:  dbSvc,
		client: client,

		peerKey:              peerKey,
		peerNetworkingConfig: peerNetworkingConfig,
		ocrKey:               ocrKey,
		// hardcoded defaults for now
		ocrConfig: Config{
			ContractPollInterval:               15 * time.Second,
			ContractTransmitterTransmitTimeout: 10 * time.Second,
			DatabaseTimeout:                    10 * time.Second,
		},

		chainID:          chainID,
		chainQueryClient: chainQueryClient,
		cosmosClient:     cosmosClient,
		tmClient:         tmClient,
		onchainSigner:    onchainSigner,
		cosmosKeyring:    cosmosKeyring,

		activeJobsMux: new(sync.RWMutex),
		activeJobs:    make(map[string]Job),
		runningMux:    new(sync.RWMutex),

		logger: log.WithFields(log.Fields{
			"svc": "ocr2_job_svc",
		}),
	}

	if err := j.restartExistingJobs(); err != nil {
		j.logger.WithError(err).Warningln("⚠️  failed to restart existing jobs")
	}

	return j
}

type Config struct {
	ContractPollInterval               time.Duration
	ContractTransmitterTransmitTimeout time.Duration
	DatabaseTimeout                    time.Duration
}

// restartExistingJobs revives jobs upon service start. Not thread
func (j *jobService) restartExistingJobs() error {
	dbCtx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()

	jobs, err := j.dbSvc.ListJobs(dbCtx, &model.Cursor{
		Limit: 10000,
	})
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if !job.IsActive {
			// be super-sure about that
			continue
		}

		if err := j.ocrStartForJob(string(job.JobID), job.Spec); err != nil {
			j.logger.WithError(err).WithField("jobID", job.JobID).Warningln("failed to start OCR for Job")
		}
	}

	if len(j.activeJobs) != len(jobs) {
		j.logger.Warningln("⚠️  not all jobs recovered successfully")
	} else if len(j.activeJobs) > 0 {
		j.logger.WithField("jobs", len(j.activeJobs)).Infoln("✅ all jobs recovered successfully")
	}

	return nil
}

func (j *jobService) StartJob(jobID string, jobSpec *model.JobSpec) error {
	j.activeJobsMux.Lock()
	defer j.activeJobsMux.Unlock()

	if _, ok := j.activeJobs[jobID]; ok {
		return errors.New("job with the same ID already running")
	}

	dbCtx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFn()
	if err := j.dbSvc.UpsertJob(dbCtx, &model.Job{
		JobID:     model.ID(jobID),
		Spec:      jobSpec,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		j.logger.WithError(err).Warningln("failed to store Job in DB")
		return ErrInternal
	}

	return j.ocrStartForJob(jobID, jobSpec)
}

func (j *jobService) ocrStartForJob(jobID string, jobSpec *model.JobSpec) error {
	stateDB, err := db.NewJobDBService(j.dbSvc.Connection(), jobID)
	if err != nil {
		err = errors.Wrap(err, "failed to init Job DB service")
		return err
	}

	transmitter := &injective.CosmosModuleTransmitter{
		FeedId:       string(jobSpec.FeedID),
		QueryClient:  j.chainQueryClient,
		CosmosClient: j.cosmosClient,
	}

	onchainKeyring := &injective.InjectiveModuleOnchainKeyring{
		Signer:  j.onchainSigner,
		Keyring: j.cosmosKeyring,
	}

	medianReporter := &injective.CosmosMedianReporter{
		FeedId:      string(jobSpec.FeedID),
		QueryClient: j.chainQueryClient,
	}

	configTracker := &injective.CosmosModuleConfigTracker{
		FeedId:           string(jobSpec.FeedID),
		QueryClient:      j.chainQueryClient,
		TendermintClient: j.tmClient,
	}

	offchainConfigDigester := &injective.CosmosOffchainConfigDigester{}

	job, err := j.newJob(
		jobID,
		jobSpec,
		stateDB,
		transmitter,
		medianReporter,
		onchainKeyring,
		configTracker,
		offchainConfigDigester,
	)
	if err != nil {
		return err
	}

	j.activeJobs[jobID] = job
	go job.Start()

	return nil
}

var (
	ErrJobNotFound = errors.New("job not found")
	ErrInternal    = errors.New("internal error")
)

func (j *jobService) RunJob(jobID, result string) error {
	j.activeJobsMux.RLock()
	defer j.activeJobsMux.RUnlock()

	activeJob, ok := j.activeJobs[jobID]
	if !ok {
		j.logger.WithField("jobID", jobID).Warningln(ErrJobNotFound)
		return ErrJobNotFound
	}

	return activeJob.Run(result)
}

func (j *jobService) StopJob(jobID string) error {
	j.activeJobsMux.Lock()
	defer j.activeJobsMux.Unlock()

	activeJob, ok := j.activeJobs[jobID]
	if !ok {
		return nil
	}

	defer func() {
		delete(j.activeJobs, jobID)

		dbCtx, cancelFn := context.WithTimeout(context.Background(), 30*time.Second)
		if err := j.dbSvc.DeleteJob(dbCtx, model.ID(jobID)); err != nil {
			j.logger.WithError(err).Warningln("failed to delete Job from DB")
		}
		cancelFn()
	}()

	return activeJob.Stop()
}

func (j *jobService) Close() (err error) {
	j.onceStop.Do(func() {
		j.runningMux.Lock()
		defer j.runningMux.Unlock()

		defer func() {
			j.running = false
			if len(j.activeJobs) > 0 {
				j.logger.Warningln("dirty exit")
			}
		}()

		for jobID, job := range j.activeJobs {
			if err := job.Stop(); err != nil {
				j.logger.WithField("jobID", jobID).WithError(err).Warningln("failed to shutdown the job")
				continue
			}

			delete(j.activeJobs, jobID)
		}
	})

	return err
}
