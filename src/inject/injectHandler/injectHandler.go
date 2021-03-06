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

package injectHandler

import (
	"log"
	"errors"
	"net"
	"os"
	"strings"
	"strconv"
	"fmt"
	"github.com/influxdata/kapacitor/udf/agent"
	"github.com/mgjeong/protocol-ezmq-go/ezmq"
	"encoding/json"
)

type injectHandler struct {
	source  string
	address string
	topic   string

	ezmqSub *ezmq.EZMQSubscriber

	agent *agent.Agent
}

var conn *net.UDPConn
var table string

func NewInjectHandler(agent *agent.Agent) *injectHandler {
	return &injectHandler{agent: agent}
}

func (p *injectHandler) Info() (*agent.InfoResponse, error) {
	info := &agent.InfoResponse{
		Wants:    agent.EdgeType_STREAM,
		Provides: agent.EdgeType_STREAM,
		Options: map[string]*agent.OptionInfo{
			"source":  {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
			"address": {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
			"topic":   {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
			"into":    {ValueTypes: []agent.ValueType{agent.ValueType_STRING}},
		},
	}
	return info, nil
}

func (p *injectHandler) Init(r *agent.InitRequest) (*agent.InitResponse, error) {
	init := &agent.InitResponse{
		Success: true,
		Error:   "",
	}

	for _, opt := range r.Options {
		switch opt.Name {
		case "source":
			p.source = opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue
		case "address":
			p.address = opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue
		case "topic":
			p.topic = opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue
		case "into":
			table = opt.Values[0].Value.(*agent.OptionValue_StringValue).StringValue
		}
	}

	if p.address == "" {
		return nil, errors.New("address must be specified")
	}

	if p.source == "" {
		return nil, errors.New("source must be specified")
	}

	if table == "" {
		return nil, errors.New("target table must be specified")
	}

	log.Println("Waiting to make source in PID ", os.Getpid())

	const udpPort = "9100"
	udpAddr, initError := net.ResolveUDPAddr("udp", "localhost:" + udpPort)
	if initError != nil {
		return nil, initError
	}
	conn, initError = net.DialUDP("udp", nil, udpAddr)
	if initError != nil {
		return nil, initError
	}

	if p.initializeEZMQ() == nil {
		return nil, initError
	}

	p.ezmqSub, initError = p.addSource()
	if initError != nil {
		return nil, initError
	}

	log.Println("Ready to inject from", p.address)
	return init, initError
}

func (p *injectHandler) initializeEZMQ() *ezmq.EZMQAPI {
	instance := ezmq.GetInstance()
	result := instance.Initialize()
	log.Println("Initializing EZMQ, error code: ", result)
	return instance
}

func (p *injectHandler) addSource() (*ezmq.EZMQSubscriber, error) {
	log.Println("Start to make source [", p.address, "] in PID ", os.Getpid())
	target := strings.Split(p.address, ":")
	port, err := strconv.Atoi(target[1])
	if err != nil {
		return nil, errors.New("invalid port number")
	}

	subCB := func(event ezmq.Event) { eventHandler(event) }
	subTopicCB := func(topic string, event ezmq.Event) { eventHandler(event) }

	subscriber := ezmq.GetEZMQSubscriber(target[0], port, subCB, subTopicCB)
	result := subscriber.Start()
	if result != ezmq.EZMQ_OK {
		return nil, errors.New("failed to subscription")
	}

	if p.topic != "" {
		result = subscriber.SubscribeForTopic(p.topic)
	} else {
		result = subscriber.Subscribe()
	}

	log.Println("subscriber is working with error ", result)
	return subscriber, nil
}

func eventHandler(event ezmq.Event) {
	var msg string

	msg = table + " "
	readings := event.GetReading()
	timeStamp := ""
	for i := 0; i < len(readings); i++ {
		body, timeStamped := jsonIntoInfluxBody(readings[i].GetValue())
		msg += body
		if timeStamped != "" {
			timeStamp = timeStamped
		}
	}
	msg = msg[:len(msg)-1]

	if timeStamp != "" {
		msg += " " + timeStamp
	}

	log.Println(os.Getpid(), "message: ", msg)

	forwardEventToKapacitor(msg)
}

func jsonIntoInfluxBody(msg string) (string, string) {
	var body string
	data := make(map[string]interface{})
	decoder := json.NewDecoder(strings.NewReader(msg))
	decoder.UseNumber()
	decoder.Decode(&data)
	var timeStamp = ""
	for key, value := range data {
		var stringValue string
		switch value.(type) {
		case string:
			stringValue = value.(string)
			body += key + "=" + fmt.Sprintf("\"%s\",", value.(string))
		case json.Number:
			stringValue = value.(json.Number).String()
			body += key + "=" + stringValue + ","
		}

		// Custom conditional statements for timestamp
		if key == "sTime" || key == "timestamp" {
			timeStamp = stringValue
		}
	}
	return body, timeStamp
}

func forwardEventToKapacitor(msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
                log.Println("Failed to forward message via UDP")
	}
}

func (p *injectHandler) Snapshot() (*agent.SnapshotResponse, error) {
	return &agent.SnapshotResponse{}, nil
}

func (p *injectHandler) Restore(req *agent.RestoreRequest) (*agent.RestoreResponse, error) {
	// Currently, all the information necessary is set when Init() is called
	// Therefore, bypass this function
	return &agent.RestoreResponse{
		Success: true,
	}, nil
}

func (p *injectHandler) BeginBatch(batch *agent.BeginBatch) error {
	return errors.New("batching is not supported")
}

func (p *injectHandler) Point(point *agent.Point) error {
	p.agent.Responses <- &agent.Response{
		Message: &agent.Response_Point{
			Point: point,
		},
	}
	return nil
}

func (p *injectHandler) EndBatch(batch *agent.EndBatch) error {
	return nil
}

func (p *injectHandler) Stop() {
	log.Println("Stopping UDF: PID", os.Getpid())
	if p.ezmqSub != nil {
		p.ezmqSub.Stop()
	}
	if conn != nil {
		conn.Close()
	}
	close(p.agent.Responses)
}
