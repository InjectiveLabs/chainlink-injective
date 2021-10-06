package db

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/metrics"
)

func (d *jobDBService) SetPersistentState(
	ctx context.Context,
	state *model.JobPersistentState,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId":        d.jobID,
		"configDigest": state.ConfigDigest,
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	upd := bson.M{
		"$set": state,
	}

	_, err := d.persistentStateCollection().UpdateOne(dbCtx, filter, upd, opts)
	if err != nil {
		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to upsert a document")
		return err
	}

	return nil
}

func (d *jobDBService) GetPersistentState(
	ctx context.Context,
	configDigest model.ID,
) (*model.JobPersistentState, error) {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId":        d.jobID,
		"configDigest": configDigest,
	}

	var state model.JobPersistentState

	opts := &options.FindOneOptions{}
	err := d.persistentStateCollection().FindOne(dbCtx, filter, opts).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			metrics.ReportFuncError(d.svcTags)
			return nil, ErrNotFound
		}

		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to query document")
		return nil, err
	}

	return &state, err
}
