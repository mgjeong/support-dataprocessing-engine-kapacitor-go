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
	"os"
	"log"
	"encoding/json"
)

type FileSink struct {
	targetFile *os.File
}

func (f *FileSink) AddSink(address, topic string) error {
	log.Println("Add a file sink", address)
	destination, err := os.Create(address)
	f.targetFile = destination
	return err
}

func (f *FileSink) Flush(record *map[string]interface{}) error {
	jsonBytes, err := json.Marshal(record)
	log.Println("Writing: ", string(jsonBytes))
	f.targetFile.Write(jsonBytes)
	f.targetFile.WriteString("\n")
	return err
}

func (f *FileSink) Close() {
	if f.targetFile != nil {
		f.targetFile.Close()
	}
}
