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

func (d *jobDBService) UpsertAnnouncement(
	ctx context.Context,
	ann *model.JobPeerAnnouncement,
) error {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	filter := bson.M{
		"jobId":  d.jobID,
		"peerId": ann.PeerID,
	}

	opts := &options.UpdateOptions{}
	opts.SetUpsert(true)
	upd := bson.M{
		"$set": ann,
	}

	_, err := d.peerAnnouncementCollection().UpdateOne(dbCtx, filter, upd, opts)
	if err != nil {
		metrics.ReportFuncError(d.svcTags)
		err = errors.Wrap(err, "failed to upsert a document")
		return err
	}

	return nil
}

func (d *jobDBService) ListAnnouncements(
	ctx context.Context,
	peerIDs []string,
	cursor *model.Cursor,
) ([]*model.JobPeerAnnouncement, error) {
	metrics.ReportFuncCall(d.svcTags)
	doneFn := metrics.ReportFuncTiming(d.svcTags)
	defer doneFn()

	dbCtx, cancelFn := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancelFn()

	q := bson.M{
		"jobId": d.jobID,
		"peerId": bson.M{
			"$in": peerIDs,
		},
	}

	opts := newFindOptionsWithCursor(cursor)
	opts.SetSort(bson.M{
		"createdAt": 1,
	})

	cur, err := d.peerAnnouncementCollection().Find(dbCtx, q, opts)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []*model.JobPeerAnnouncement{}, nil
		}

		err = errors.Wrap(err, "failed to query documents")
		return nil, err
	}

	var peerAnnouncements []*model.JobPeerAnnouncement
	if err := cur.All(dbCtx, &peerAnnouncements); err != nil {

		err = errors.Wrap(err, "failed to decode documents")
		return nil, err
	}

	return peerAnnouncements, nil
}
