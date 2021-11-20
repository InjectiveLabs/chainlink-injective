package db

import (
	"context"
	"database/sql"
	"net/url"

	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/pkg/errors"
	postgres_models "github.com/smartcontractkit/chainlink-relay/core/store/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ExternalGorm interface {
	CreateJob(ctx context.Context, job *model.Job) error
	LoadJobs(ctx context.Context) ([]*model.Job, error)
	DeleteJob(ctx context.Context, jobID string) error

	Client() *gorm.DB
	Connection() (*sql.DB, error)
	String() string
}

type externalGorm struct {
	db *gorm.DB
}

func NewExternalPostgres(u *url.URL) (ExternalGorm, error) {
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{})
	if err != nil {
		err = errors.Wrap(err, "gorm failed to connect to PostgreSQL DB")
		return nil, err
	}

	e := &externalGorm{
		db: db,
	}

	if err := e.migrate(); err != nil {
		err = errors.Wrap(err, "gorm migration failed")
		return nil, err
	}

	return e, nil
}

func (e *externalGorm) Client() *gorm.DB {
	return e.db
}

func (e *externalGorm) Connection() (*sql.DB, error) {
	return e.db.DB()
}

func (e *externalGorm) String() string {
	return "DB Driver: gorm SQL"
}

func (e *externalGorm) migrate() error {
	if err := e.db.AutoMigrate(&postgres_models.Job{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.EncryptedKeyRings{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.EthKeyStates{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.P2pPeers{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.Offchainreporting2ContractConfigs{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.Offchainreporting2DiscovererAnnouncements{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.Offchainreporting2PersistentStates{}); err != nil {
		return err
	}
	if err := e.db.AutoMigrate(&postgres_models.Offchainreporting2PendingTransmissions{}); err != nil {
		return err
	}

	return nil
}

// CreateJob saves the job data in the DB
func (e *externalGorm) CreateJob(ctx context.Context, job *model.Job) error {
	ormJob := jobToOrm(job)

	if len(ormJob.JobID) == 0 {
		return errors.New("JobID cannot be empty")
	}

	return e.db.WithContext(ctx).Create(ormJob).Error
}

// LoadJobs retrieves all jobs from the DB
func (e *externalGorm) LoadJobs(ctx context.Context) ([]*model.Job, error) {
	var ormJobs []postgres_models.Job

	err := e.db.WithContext(ctx).Find(&ormJobs).Error
	if err != nil {
		err = errors.Wrapf(err, "failed to query jobs")
		return nil, err
	}

	jobs := make([]*model.Job, 0, len(ormJobs))
	for i := range ormJobs {
		jobs = append(jobs, ormToJob(&ormJobs[i]))
	}

	return jobs, err
}

// loadJob retrieves a specific job from the DB
func (e *externalGorm) loadJob(ctx context.Context, jobID string, discard bool) (*model.Job, error) {
	var ormJob postgres_models.Job

	if err := e.db.WithContext(ctx).Where("job_id = ?", jobID).First(&ormJob).Error; err != nil {
		return nil, err
	}

	if discard {
		return nil, nil
	}

	return ormToJob(&ormJob), nil
}

// DeleteJob removes the job data from the DB
func (e *externalGorm) DeleteJob(ctx context.Context, jobID string) error {
	if len(jobID) == 0 {
		return errors.New("JobID cannot be empty")
	}

	job, err := e.loadJob(ctx, jobID, true)
	if err != nil {
		err = errors.Wrapf(err, "failed to load job data for %s", jobID)
		return err
	}

	return e.db.WithContext(ctx).Delete(job).Error
}

func jobToOrm(job *model.Job) *postgres_models.Job {
	ormJob := &postgres_models.Job{
		JobID:           string(job.JobID),
		IsBootstrapPeer: job.Spec.IsBootstrapPeer,
		ContractAddress: string(job.Spec.FeedID),
		KeyBundleID:     string(job.Spec.KeyID),

		ContractConfigConfirmations:            uint16(job.Spec.ContractConfigConfirmations),
		ContractConfigTrackerSubscribeInterval: job.Spec.ContractConfigTrackerSubscribeInterval,
		ObservationTimeout:                     job.Spec.ObservationTimeout,
		BlockchainTimeout:                      job.Spec.BlockchainTimeout,
	}

	for _, peer := range job.Spec.P2PBootstrapPeers {
		ormJob.P2PBootstrapPeers = append(ormJob.P2PBootstrapPeers, peer)
	}

	return ormJob
}

func ormToJob(ormJob *postgres_models.Job) *model.Job {
	job := &model.Job{
		JobID: model.ID(ormJob.JobID),
		Spec: &model.JobSpec{
			IsBootstrapPeer: ormJob.IsBootstrapPeer,
			FeedID:          model.ID(ormJob.ContractAddress),
			KeyID:           model.ID(ormJob.KeyBundleID),

			ContractConfigConfirmations:            int(ormJob.ContractConfigConfirmations),
			ContractConfigTrackerSubscribeInterval: ormJob.ContractConfigTrackerSubscribeInterval,
			ObservationTimeout:                     ormJob.ObservationTimeout,
			BlockchainTimeout:                      ormJob.BlockchainTimeout,
		},
		IsActive: true,
	}

	for _, peer := range ormJob.P2PBootstrapPeers {
		job.Spec.P2PBootstrapPeers = append(job.Spec.P2PBootstrapPeers, peer)
	}

	return job
}
