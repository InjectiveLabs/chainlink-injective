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

func (d *jobDBService) SetContractConfig(
	ctx context.Context,
	state *model.JobContractConfig,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId": d.jobID,
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	upd := bson.M{
		"$set": state,
	}

	_, err := d.contractConfigCollection().UpdateOne(dbCtx, filter, upd, opts)
	if err != nil {
		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to update a document")
		return err
	}

	return nil
}

func (d *jobDBService) GetContractConfig(
	ctx context.Context,
) (*model.JobContractConfig, error) {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId": d.jobID,
	}

	var state model.JobContractConfig

	opts := &options.FindOneOptions{}
	err := d.contractConfigCollection().FindOne(dbCtx, filter, opts).Decode(&state)
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
