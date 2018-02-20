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
	"testing"
	"io/ioutil"
	"os"
	"fmt"
)

const testFilename = "./testfilesink.txt"
const key = "id"
const value = "testdata"
var expectedResult = fmt.Sprintf("{\"%s\":\"%s\"}\n", key, value)

func TestFileSink_AddSink(t *testing.T) {
	testSink := new(FileSink)
	err := testSink.AddSink(testFilename, "")
	if err != nil {
		t.Error(err)
	}
	testMap := make(map[string]interface{})
	testMap[key] = value
	testSink.Flush(&testMap)
	if err != nil {
		t.Error(err)
	}
	testSink.Close()

	var data []byte
	data, err = ioutil.ReadFile(testFilename)
	if err != nil {
		t.Error(err)
	}
	if string(data) != expectedResult {
		t.Log("Written context:", string(data), ", but expected:", expectedResult)
		t.Error("Wrong context was written")
	}
	err = os.Remove(testFilename)
	if err != nil {
		t.Error("cannot remove testfile")
	}
}
