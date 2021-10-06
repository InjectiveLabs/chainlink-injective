package dbconn

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Conn interface {
	Backend() interface{}
	DatabaseName() string
	TestConn(ctx context.Context) error
	Close() error
}

// NewMongoConn creates a new MongoDB connection and initialzied client.
func NewMongoConn(ctx context.Context, cfg *MongoConfig) (Conn, error) {
	cfg = checkMongoConfig(cfg)
	opt := options.Client()
	opt.ApplyURI(cfg.Connection)
	opt.SetAppName(cfg.AppName)
	opt.SetReadPreference(readpref.Primary())

	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		return nil, err
	}

	conn := &mongoConn{
		client: client,
		cfg:    cfg,
	}

	return conn, nil
}

type MongoConfig struct {
	AppName    string
	Connection string
	Database   string
	Debug      bool
}

func checkMongoConfig(cfg *MongoConfig) *MongoConfig {
	if cfg == nil {
		cfg = &MongoConfig{}
	}
	if len(cfg.AppName) == 0 {
		cfg.AppName = "go.mongodb.org/mongo-driver"
	}
	if len(cfg.Connection) == 0 {
		cfg.Connection = "mongodb://127.0.0.1:27017"
	}
	if len(cfg.Database) == 0 {
		cfg.Database = "default"
	}

	return cfg
}

type mongoConn struct {
	client *mongo.Client
	cfg    *MongoConfig
}

func (c *mongoConn) Backend() interface{} {
	return c.client
}

func (c *mongoConn) DatabaseName() string {
	return c.cfg.Database
}

func (c *mongoConn) TestConn(ctx context.Context) error {
	err := c.client.Ping(ctx, nil)
	if err != nil {
		err = errors.Wrap(err, "MongoDB connection test failed")
		return err
	}

	return nil
}

func (c *mongoConn) Close() error {
	return c.client.Disconnect(context.Background())
}

func MakeIndex(unique bool, keys interface{}) mongo.IndexModel {
	idx := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index(),
	}
	if unique {
		idx.Options = idx.Options.SetUnique(true)
	}

	return idx
}

func MakeIndexWithExpiry(ttl int, unique bool, keys interface{}) mongo.IndexModel {
	idx := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index(),
	}
	if ttl > 0 {
		idx.Options.SetExpireAfterSeconds(int32(ttl))
	}
	if unique {
		idx.Options = idx.Options.SetUnique(true)
	}

	return idx
}
