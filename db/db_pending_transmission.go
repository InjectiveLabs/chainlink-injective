package db

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/metrics"
)

func (d *jobDBService) InsertPendingTranmission(
	ctx context.Context,
	pendingTx *model.JobPendingTransmission,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	// ensure pending Tx is saved under correct Job ID
	pendingTx.JobID = model.ID(d.jobID)

	opts := &options.InsertOneOptions{}
	_, err := d.pendingTransmissionCollection().InsertOne(dbCtx, pendingTx, opts)
	if err != nil {
		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to upsert a document")
		return err
	}

	return nil
}

func (d *jobDBService) ListPendingTransmissions(
	ctx context.Context,
	configDigest model.ID,
	cursor *model.Cursor,
) ([]*model.JobPendingTransmission, error) {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"jobId":        d.jobID,
		"configDigest": configDigest,
	}

	opts := newFindOptionsWithCursor(cursor)
	opts.SetSort(bson.M{
		"createdAt": 1,
	})

	cur, err := d.pendingTransmissionCollection().Find(dbCtx, q, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []*model.JobPendingTransmission{}, nil
		}

		err = errors.Wrap(err, "failed to query documents")
		return nil, err
	}

	var pendingTransmissions []*model.JobPendingTransmission
	if err := cur.All(dbCtx, &pendingTransmissions); err != nil {

		err = errors.Wrap(err, "failed to decode documents")
		return nil, err
	}

	return pendingTransmissions, nil
}

func (d *jobDBService) DeletePendingTransmission(
	ctx context.Context,
	reportTimestamp model.ReportTimestamp,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"jobId":           d.jobID,
		"reportTimestamp": reportTimestamp,
	}

	opts := &options.DeleteOptions{}
	_, err := d.pendingTransmissionCollection().DeleteOne(dbCtx, q, opts)
	if err != nil {
		err = errors.Wrap(err, "failed to delete documents")
		return err
	}

	return nil
}

func (d *jobDBService) DeletePendingTransmissionsOlderThan(
	ctx context.Context,
	timestamp time.Time,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"jobId": d.jobID,
		"tx.createdAt": bson.M{
			"$lt": primitive.NewDateTimeFromTime(timestamp),
		},
	}

	opts := &options.DeleteOptions{}
	_, err := d.pendingTransmissionCollection().DeleteMany(dbCtx, q, opts)
	if err != nil {
		err = errors.Wrap(err, "failed to delete documents")
		return err
	}

	return nil
}
