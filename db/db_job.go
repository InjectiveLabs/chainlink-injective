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

func (d *dbService) UpsertJob(
	ctx context.Context,
	job *model.Job,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId": job.JobID,
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	upd := bson.M{
		"$set": job,
	}

	_, err := d.jobCollection().UpdateOne(dbCtx, filter, upd, opts)
	if err != nil {
		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to upsert a document")
		return err
	}

	return nil
}

func (d *dbService) DeleteJob(
	ctx context.Context,
	jobID model.ID,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"jobId": jobID,
	}

	opts := &options.DeleteOptions{}
	_, err := d.jobCollection().DeleteOne(dbCtx, q, opts)
	if err != nil {
		err = errors.Wrap(err, "failed to delete documents")
		return err
	}

	return nil
}

func (d *dbService) ListJobs(
	ctx context.Context,
	cursor *model.Cursor,
) ([]*model.Job, error) {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"isActive": true,
	}

	opts := newFindOptionsWithCursor(cursor)
	opts.SetSort(bson.M{
		"createdAt": 1,
	})

	cur, err := d.jobCollection().Find(dbCtx, q, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []*model.Job{}, nil
		}

		err = errors.Wrap(err, "failed to query documents")
		return nil, err
	}

	var jobs []*model.Job
	if err := cur.All(dbCtx, &jobs); err != nil {

		err = errors.Wrap(err, "failed to decode documents")
		return nil, err
	}

	return jobs, nil
}
