// Copyright (C) 2015 Nippon Telegraph and Telephone Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/cihub/seelog"
	. "github.com/osrg/earthquake/earthquake/signal"
)

// idempotent (RFC 7231)
//
// client can be http.DefaultClient in most cases
func DeleteAction(client *http.Client, ocURL string, action Action) error {
	url := ocURL + "/actions/" + action.EntityID() + "/" + action.ID()
	log.Debugf("REST deleting action of %s", action, url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	// respStr must be an empty json
	log.Debugf("REST deleted action of %s, resp=%d(%s)", action, url, resp.StatusCode, respBody)
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response %#v while deleting action %s to %s", resp, action, url)
	}
	return nil
}

// client can be http.DefaultClient in most cases
func SendEvent(client *http.Client, ocURL string, event Event) error {
	jsonStr, err := json.Marshal(event.JSONMap())
	if err != nil {
		return err
	}
	url := ocURL + "/events/" + event.EntityID() + "/" + event.ID()
	log.Debugf("REST sending event %s to %s", event, url)
	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonStr))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	// respStr must be an empty json
	log.Debugf("REST sent event %s to %s, resp=%d(%s)", event, url, resp.StatusCode, respBody)
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response %#v while sending event %s to %s", resp, event, url)
	}
	return nil
}

// client can be http.DefaultClient in most cases
func GetAction(client *http.Client, ocURL string, entityID string) (Action, error) {
	url := ocURL + "/actions/" + entityID
	log.Debugf("REST getting action %s", url)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	log.Debugf("REST got action %s, resp=%d(%s)", url, resp.StatusCode, respBody)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response %#v while getting action %s", resp, url)
	}
	signal, err := NewSignalFromJSONString(string(respBody), time.Now())
	action, ok := signal.(Action)
	if !ok {
		return nil, fmt.Errorf("cannot convert %s to Action", signal)
	}
	return action, nil
}

// util for SendEvent/GetAction/DeleteAction
type Transceiver struct {
	OrchestratorURL string
	EntityID        string
	Client          *http.Client
	m               map[string]chan Action // key: event id
}

func NewTransceiver(orchestratorURL string, entityID string) (*Transceiver, error) {
	t := Transceiver{
		OrchestratorURL: orchestratorURL,
		EntityID:        entityID,
		Client:          http.DefaultClient,
		m:               make(map[string]chan Action),
	}
	return &t, nil
}

func (this *Transceiver) SendEvent(event Event) (chan Action, error) {
	if event.EntityID() != this.EntityID {
		return nil, fmt.Errorf("bad entity id for event %s (want %s)", event, this.EntityID)
	}
	err := SendEvent(this.Client, this.OrchestratorURL, event)
	if err != nil {
		return nil, err
	}
	this.m[event.ID()] = make(chan Action)
	return this.m[event.ID()], nil
}

func (this *Transceiver) onAction(action Action) error {
	event := action.Event()
	if event == nil {
		return fmt.Errorf("No event found for action %s", action)
	}
	// NOTE: this event can be dummy nop event
	actionChan, ok := this.m[event.ID()]
	if !ok {
		return fmt.Errorf("No channel found for action %s (event id=%s)", action, event.ID())
	}
	delete(this.m, event.ID())
	go func() {
		actionChan <- action
	}()
	return nil
}

func (this *Transceiver) routine() {
	errors := 0
	onHTTPError := func(err error) {
		log.Error(err)
		errors += 1
		time.Sleep(time.Duration(errors) * time.Second)
	}
	onActionError := func(err error) {
		log.Error(err)
	}
	for {
		action, err := GetAction(this.Client, this.OrchestratorURL, this.EntityID)
		if err != nil {
			onHTTPError(err)
			continue
		}
		err = DeleteAction(this.Client, this.OrchestratorURL, action)
		if err != nil {
			onHTTPError(err)
			continue
		}
		err = this.onAction(action)
		if err != nil {
			onActionError(err)
			continue
		}
		errors = 0
	}
}

func (this *Transceiver) Start() {
	go this.routine()
}
