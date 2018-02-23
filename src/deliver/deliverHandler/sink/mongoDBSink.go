// Copyright 2018 Samsung Electronics All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and 
// limitations under the License.

package sink

import (
	"log"
	"gopkg.in/mgo.v2"
	"time"
	"errors"
	"strings"
)

type MongoDBSink struct {
	session    *mgo.Session
	collection *mgo.Collection
}

func (m *MongoDBSink) AddSink(address, topic string) error {
	log.Println("DPRuntime mongoDB sink")
	var err error
	// Kapacitor starts to receive data regardless of mongoDB connected
	// If session fails, this Kapacitor task will be stopped
	// This should be fixed not to start and reported to Runtime in the first place
	m.session, err = mgo.DialWithTimeout(address, 5*time.Second)
	if err != nil {
		return errors.New("error: failed to connect MongoDB")
	}

	if topic == "" {
		return errors.New("error: DB and collection names must be provided")
	}

	dbSplits := strings.Split(topic, ":")
	if len(dbSplits) != 2 {
		return errors.New("error: DB and collection must be specified as DB:COLLECTION")
	}
	m.collection = m.session.DB(dbSplits[0]).C(dbSplits[1])
	return nil
}

func (m *MongoDBSink) Flush(record *map[string]interface{}) error {
	log.Println("Writing into MongoDB")
	err := m.collection.Insert(record)
	return err
}

func (m *MongoDBSink) Close() {
	if m.session != nil {
		m.session.Close()
	}
}
