package p2p

import (
	"context"
	"time"

	"github.com/InjectiveLabs/chainlink-injective/db"
	"github.com/InjectiveLabs/chainlink-injective/db/model"
)

var _ DiscovererDatabase = &announceDBWrapper{}

type announceDBWrapper struct {
	svc db.JobDBService
}

func NewAnnounceDBWrapper(dbSvc db.JobDBService) DiscovererDatabase {
	return &announceDBWrapper{
		svc: dbSvc,
	}
}

func (j *announceDBWrapper) StoreAnnouncement(ctx context.Context, peerID string, ann []byte) error {
	return j.svc.UpsertAnnouncement(ctx, &model.JobPeerAnnouncement{
		JobID:     j.svc.JobID(),
		PeerID:    model.ID(peerID),
		Announce:  ann,
		CreatedAt: time.Now().UTC(),
	})
}

func (j *announceDBWrapper) ReadAnnouncements(ctx context.Context, peerIDs []string) (map[string][]byte, error) {
	announcements, err := j.svc.ListAnnouncements(ctx, peerIDs, &model.Cursor{
		Limit: 10000,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[string][]byte, len(announcements))
	for _, ann := range announcements {
		result[string(ann.PeerID)] = ann.Announce
	}

	return result, nil
}
