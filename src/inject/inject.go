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

package main

import (
	"github.com/influxdata/kapacitor/udf/agent"
	"log"
	"os"
	"inject/injectHandler"
)

func main() {
	thisAgent := agent.New(os.Stdin, os.Stdout)
	thisHandler := injectHandler.NewInjectHandler(thisAgent)
	thisAgent.Handler = thisHandler

	log.Println("Starting injecting agent: PID", os.Getpid())
	thisAgent.Start()
	err := thisAgent.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Finishing injecting agent: PID", os.Getpid())
}
