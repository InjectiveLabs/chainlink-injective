package db

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/InjectiveLabs/chainlink-injective/db/model"
)

func newFindOptionsWithCursor(cursor *model.Cursor) *options.FindOptions {
	opts := &options.FindOptions{}
	opts.SetHint(bson.D{{"_id", 1}})

	if cursor == nil {
		cursor = &model.Cursor{
			Limit: defaultListLimit,
		}
	}
	if !cursor.From.IsZero() {
		opts.SetMin(bson.D{{"_id", cursor.From}})
	}
	if cursor.To != nil {
		opts.SetMax(bson.D{{"_id", cursor.To}})
	}
	if cursor.Limit >= 0 {
		opts.SetLimit(int64(cursor.Limit))
	}

	return opts
}
