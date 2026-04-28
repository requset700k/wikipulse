// EC2Orchestrator: AWS SDK로 세션당 EC2 인스턴스를 생성/삭제한다.
// Launch Template + cloud-init user-data로 SSH 서버를 자동 설치.
// KubeVirt 풀 소진 시 HybridOrchestrator가 폴백으로 호출.
package vm

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Orchestrator provisions VMs on AWS EC2 using a Launch Template.
// Used as overflow when KubeVirt pool is full.
type EC2Orchestrator struct {
	client         *ec2.Client
	launchTemplate string
	instanceType   ec2types.InstanceType
}

func NewEC2(client *ec2.Client, launchTemplate, instanceType string) *EC2Orchestrator {
	return &EC2Orchestrator{
		client:         client,
		launchTemplate: launchTemplate,
		instanceType:   ec2types.InstanceType(instanceType),
	}
}

func (e *EC2Orchestrator) Create(ctx context.Context, req CreateRequest) (*VMInfo, error) {
	userDataScript := fmt.Sprintf(`#!/bin/bash
SESSION_ID=%s
LAB_ID=%s
apt-get update -qq
apt-get install -y openssh-server
systemctl enable ssh && systemctl start ssh
`, req.SessionID, req.LabID)

	out, err := e.client.RunInstances(ctx, &ec2.RunInstancesInput{
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		InstanceType: e.instanceType,
		LaunchTemplate: &ec2types.LaunchTemplateSpecification{
			LaunchTemplateName: aws.String(e.launchTemplate),
		},
		UserData: aws.String(userDataScript),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags: []ec2types.Tag{
					{Key: aws.String("session-id"), Value: aws.String(req.SessionID)},
					{Key: aws.String("lab-id"), Value: aws.String(req.LabID)},
					{Key: aws.String("managed-by"), Value: aws.String("cledyu")},
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("RunInstances: %w", err)
	}

	instance := out.Instances[0]
	return &VMInfo{
		ID:       aws.ToString(instance.InstanceId),
		Provider: ProviderEC2,
		Port:     22,
	}, nil
}

func (e *EC2Orchestrator) Status(ctx context.Context, vmID string, _ Provider) (*VMStatus, error) {
	out, err := e.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{vmID},
	})
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances: %w", err)
	}

	if len(out.Reservations) == 0 || len(out.Reservations[0].Instances) == 0 {
		return &VMStatus{Ready: false, Phase: "not-found"}, nil
	}

	inst := out.Reservations[0].Instances[0]
	state := string(inst.State.Name)
	ready := state == "running" && aws.ToString(inst.PublicIpAddress) != ""

	return &VMStatus{Ready: ready, Phase: state}, nil
}

func (e *EC2Orchestrator) Delete(ctx context.Context, vmID string, _ Provider) error {
	_, err := e.client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{vmID},
	})
	return err
}
