/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"encoding/json"
	"errors"
	"log"
)

var (
	ErrDatacenterIDInvalid          = errors.New("Datacenter VPC ID invalid")
	ErrDatacenterRegionInvalid      = errors.New("Datacenter Region invalid")
	ErrDatacenterCredentialsInvalid = errors.New("Datacenter credentials invalid")
	ErrNetworkIDInvalid             = errors.New("Network id invalid")
	ErrNatGatewayIDInvalid          = errors.New("Nat Gateway aws id invalid")
)

// Event stores the nat gateway data
type Event struct {
	UUID                   string `json:"_uuid"`
	BatchID                string `json:"_batch_id"`
	ProviderType           string `json:"_type"`
	VPCID        string `json:"vpc_id"`
	DatacenterRegion       string `json:"datacenter_region"`
	DatacenterAccessKey    string `json:"datacenter_access_key"`
	DatacenterAccessToken  string `json:"datacenter_access_token"`
	NetworkAWSID           string `json:"network_aws_id"`
	NatGatewayAWSID        string `json:"nat_gateway_aws_id"`
	NatGatewayAllocationID string `json:"nat_gateway_allocation_id"`
	NatGatewayAllocationIP string `json:"nat_gateway_allocation_ip"`
	ErrorMessage           string `json:"error,omitempty"`
}

// Validate checks if all criteria are met
func (ev *Event) Validate() error {
	if ev.VPCID == "" {
		return ErrDatacenterIDInvalid
	}

	if ev.DatacenterRegion == "" {
		return ErrDatacenterRegionInvalid
	}

	if ev.DatacenterAccessKey == "" || ev.DatacenterAccessToken == "" {
		return ErrDatacenterCredentialsInvalid
	}

	if ev.NatGatewayAWSID == "" {
		return ErrNatGatewayIDInvalid
	}

	return nil
}

// Process the raw event
func (ev *Event) Process(data []byte) error {
	err := json.Unmarshal(data, &ev)
	if err != nil {
		nc.Publish("nat.delete.aws.error", data)
	}
	return err
}

// Error the request
func (ev *Event) Error(err error) {
	log.Printf("Error: %s", err.Error())
	ev.ErrorMessage = err.Error()

	data, err := json.Marshal(ev)
	if err != nil {
		log.Panic(err)
	}
	nc.Publish("nat.delete.aws.error", data)
}

// Complete the request
func (ev *Event) Complete() {
	data, err := json.Marshal(ev)
	if err != nil {
		ev.Error(err)
	}
	nc.Publish("nat.delete.aws.done", data)
}
