package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/InjectiveLabs/chainlink-injective/db/dbconn"
	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/metrics"
)

type DBService interface {
	JobCollection

	DBName() string
	Client() *mongo.Client
	Connection() dbconn.Conn
	Close()
}

var _ DBService = &dbService{}

type JobCollection interface {
	UpsertJob(
		ctx context.Context,
		spec *model.Job,
	) error

	DeleteJob(
		ctx context.Context,
		jobID model.ID,
	) error

	ListJobs(
		ctx context.Context,
		cursor *model.Cursor,
	) ([]*model.Job, error)
}

type JobDBService interface {
	PersistentStateCollection
	ContractConfigCollection
	PendingTransmissionCollection
	PeerAnnouncementCollection

	JobID() model.ID
	DBName() string
	Client() *mongo.Client
	Close()
}

var _ JobDBService = &jobDBService{}

type PersistentStateCollection interface {
	SetPersistentState(
		ctx context.Context,
		state *model.JobPersistentState,
	) error

	GetPersistentState(
		ctx context.Context,
		configDigest model.ID,
	) (*model.JobPersistentState, error)
}

type ContractConfigCollection interface {
	SetContractConfig(
		ctx context.Context,
		config *model.JobContractConfig,
	) error

	GetContractConfig(
		ctx context.Context,
	) (*model.JobContractConfig, error)
}

type PendingTransmissionCollection interface {
	InsertPendingTranmission(
		ctx context.Context,
		pendingTx *model.JobPendingTransmission,
	) error

	ListPendingTransmissions(
		ctx context.Context,
		configDigest model.ID,
		cursor *model.Cursor,
	) ([]*model.JobPendingTransmission, error)

	DeletePendingTransmission(
		ctx context.Context,
		reportTimestamp model.ReportTimestamp,
	) error

	DeletePendingTransmissionsOlderThan(
		ctx context.Context,
		timestamp time.Time,
	) error
}

type PeerAnnouncementCollection interface {
	UpsertAnnouncement(
		ctx context.Context,
		ann *model.JobPeerAnnouncement,
	) error

	ListAnnouncements(
		ctx context.Context,
		peerIDs []string,
		cursor *model.Cursor,
	) ([]*model.JobPeerAnnouncement, error)
}

func NewDBService(
	conn dbconn.Conn,
) (DBService, error) {
	d := &dbService{
		conn: conn,
		db:   conn.Backend().(*mongo.Client),

		svcTags: metrics.Tags{
			"svc": "db",
		},
	}

	d.ensureIndex()

	return d, nil
}

type dbService struct {
	conn dbconn.Conn
	db   *mongo.Client

	svcTags metrics.Tags
}

func (d *dbService) DBName() string {
	return d.conn.DatabaseName()
}

func (d *dbService) Client() *mongo.Client {
	return d.db
}

func (d *dbService) Connection() dbconn.Conn {
	return d.conn
}

func (d *dbService) Close() {
	return
}

func (d *dbService) jobCollection() *mongo.Collection {
	return d.db.Database(d.conn.DatabaseName()).Collection("jobs")
}

func (d *dbService) ensureIndex() {
	_, _ = d.jobCollection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		dbconn.MakeIndex(true, bson.D{{"jobId", 1}}),
	})
}

func NewJobDBService(
	conn dbconn.Conn,
	jobID string,
) (JobDBService, error) {
	d := &jobDBService{
		conn:  conn,
		db:    conn.Backend().(*mongo.Client),
		jobID: jobID,

		svcTags: metrics.Tags{
			"svc": "job_db",
			"job": jobID,
		},
	}

	d.ensureIndex()

	return d, nil
}

var ErrNotFound = errors.New("object not found")

const (
	defaultListLimit           = 100
	defaultQueryTimeout        = 5 * time.Minute
	defaultSlowClientTolerance = 500 * time.Millisecond
)

type jobDBService struct {
	conn  dbconn.Conn
	db    *mongo.Client
	jobID string

	svcTags metrics.Tags
}

func (d *jobDBService) JobID() model.ID {
	return model.ID(d.jobID)
}

func (d *jobDBService) DBName() string {
	return d.conn.DatabaseName()
}

func (d *jobDBService) Client() *mongo.Client {
	return d.db
}

func (d *jobDBService) Close() {
	return
}

func (d *jobDBService) persistentStateCollection() *mongo.Collection {
	return d.db.Database(d.conn.DatabaseName()).Collection("job_persistent_states")
}

func (d *jobDBService) contractConfigCollection() *mongo.Collection {
	return d.db.Database(d.conn.DatabaseName()).Collection("job_contract_configs")
}

func (d *jobDBService) pendingTransmissionCollection() *mongo.Collection {
	return d.db.Database(d.conn.DatabaseName()).Collection("job_pending_transmissions")
}

func (d *jobDBService) peerAnnouncementCollection() *mongo.Collection {
	return d.db.Database(d.conn.DatabaseName()).Collection("job_peer_announcements")
}

func (d *jobDBService) ensureIndex() {
	_, _ = d.persistentStateCollection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		dbconn.MakeIndex(true, bson.D{{"jobId", 1}}),
		dbconn.MakeIndex(false, bson.D{{"configDigest", 1}}),
		dbconn.MakeIndex(false, bson.D{{"isActive", 1}}),
	})

	_, _ = d.contractConfigCollection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		dbconn.MakeIndex(true, bson.D{{"jobId", 1}}),
		dbconn.MakeIndex(true, bson.D{{"jobId", 1}, {"configDigest", 1}}),
	})

	_, _ = d.pendingTransmissionCollection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		dbconn.MakeIndex(false, bson.D{{"jobId", 1}}),
		dbconn.MakeIndex(false, bson.D{{"configDigest", 1}}),
		dbconn.MakeIndex(true, bson.D{{"reportTimestamp.epoch", 1}, {"reportTimestamp.round", 1}}),
		dbconn.MakeIndex(false, bson.D{{"tx.createdAt", 1}}),
	})

	_, _ = d.peerAnnouncementCollection().Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		dbconn.MakeIndex(false, bson.D{{"jobId", 1}}),
		dbconn.MakeIndex(true, bson.D{{"jobId", 1}, {"peerId", 1}}),
	})
}
