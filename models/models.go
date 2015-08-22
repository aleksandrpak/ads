package models

import "gopkg.in/mgo.v2/bson"

type ID struct {
	ID bson.ObjectId `bson:"_id,omitempty"`
}
