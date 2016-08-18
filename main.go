/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	ecc "github.com/ernestio/ernest-config-client"
	"github.com/nats-io/nats"
)

var nc *nats.Conn
var natsErr error

func eventHandler(m *nats.Msg) {
	var n Event

	err := n.Process(m.Data)
	if err != nil {
		nc.Publish("nat.delete.aws.error", m.Data)
		return
	}

	if err = n.Validate(); err != nil {
		n.Error(err)
		return
	}

	err = deleteNat(&n)
	if err != nil {
		n.Error(err)
		return
	}

	n.Complete()
}

func natGatewayByID(svc *ec2.EC2, id string) (*ec2.NatGateway, error) {
	req := ec2.DescribeNatGatewaysInput{
		NatGatewayIds: []*string{aws.String(id)},
	}
	resp, err := svc.DescribeNatGateways(&req)
	if err != nil {
		return nil, err
	}

	if len(resp.NatGateways) != 1 {
		return nil, errors.New("Could not find nat gateway")
	}

	return resp.NatGateways[0], nil
}

func isNatGatewayDeleted(svc *ec2.EC2, id string) bool {
	gw, _ := natGatewayByID(svc, id)
	if *gw.State == ec2.NatGatewayStateDeleted {
		return true
	}

	return false
}

func deleteNat(ev *Event) error {
	creds := credentials.NewStaticCredentials(ev.DatacenterAccessKey, ev.DatacenterAccessToken, "")
	svc := ec2.New(session.New(), &aws.Config{
		Region:      aws.String(ev.DatacenterRegion),
		Credentials: creds,
	})

	req := ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(ev.NatGatewayAWSID),
	}

	_, err := svc.DeleteNatGateway(&req)
	if err != nil {
		return err
	}

	for isNatGatewayDeleted(svc, ev.NatGatewayAWSID) {
		time.Sleep(time.Second * 3)
	}

	return nil
}

func main() {
	nc = ecc.NewConfig(os.Getenv("NATS_URI")).Nats()

	fmt.Println("listening for nat.delete.aws")
	nc.Subscribe("nat.delete.aws", eventHandler)

	runtime.Goexit()
}
