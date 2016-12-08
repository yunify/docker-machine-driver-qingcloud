package qingcloud

import (
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/yunify/qingcloud-sdk-go/config"
	qcservice "github.com/yunify/qingcloud-sdk-go/service"
	"time"
)

const (
	INSTANCE_STATUS_PENDING    = "pending"
	INSTANCE_STATUS_RUNNING    = "running"
	INSTANCE_STATUS_STOPPED    = "stopped"
	INSTANCE_STATUS_SUSPENDED  = "suspended"
	INSTANCE_STATUS_TERMINATED = "terminated"
	INSTANCE_STATUS_CEASED     = "ceased"
)

type Client interface {
	RunInstance(arg *RunInstanceArg) (*qcservice.Instance, error)
	DescribeInstance(instanceID string) (*qcservice.Instance, error)
	StartInstance(instanceID string) error
	StopInstance(instanceID string, force bool) error
	RestartInstance(instanceID string) error
	TerminateInstance(instanceID string) error
	WaitInstanceStatus(instanceID string, status string) error
	CreateKeyPair(keyPairName string, publicKey string) (string, error)
	DescribeKeyPair(keyPairID string) (*qcservice.KeyPair, error)
	DeleteKeyPair(keyPairID string) error
}

func NewClient(config *config.Config, zone string) (Client, error) {
	qcService, err := qcservice.Init(config)
	if err != nil {
		return nil, err
	}
	instanceService, err := qcService.Instance(zone)
	if err != nil {
		return nil, err
	}
	jobService, err := qcService.Job(zone)
	if err != nil {
		return nil, err
	}
	keypairService, err := qcService.KeyPair(zone)
	if err != nil {
		return nil, err
	}
	c := &client{
		instanceService: instanceService,
		jobService:      jobService,
		keypairService:  keypairService,
		opTimeout:       defaultOpTimeout,
		zone:            zone,
	}
	return c, nil
}

type client struct {
	instanceService *qcservice.InstanceService
	jobService      *qcservice.JobService
	keypairService  *qcservice.KeyPairService
	opTimeout       int
	zone            string
}

type RunInstanceArg struct {
	CPU          int
	Memory       int
	ImageID      string
	LoginKeyPair string
	VxNet        string
	InstanceName string
}

func (c *client) RunInstance(arg *RunInstanceArg) (*qcservice.Instance, error) {
	input := &qcservice.RunInstancesInput{
		CPU:          arg.CPU,
		Count:        1,
		ImageID:      arg.ImageID,
		Memory:       arg.Memory,
		LoginKeyPair: arg.LoginKeyPair,
		LoginMode:    "keypair",
		InstanceName: arg.InstanceName,
		//SecurityGroup string   `json:"security_group" name:"security_group" location:"requestParams"`
		VxNets: []string{arg.VxNet},
		//Volumes       []string `json:"volumes" name:"volumes" location:"requestParams"`
	}

	output, err := c.instanceService.RunInstances(input)
	if err != nil {
		return nil, err
	}
	if len(output.Instances) == 0 {
		return nil, errors.New("Create instance response error.")
	}
	jobID := output.JobID
	jobErr := c.waitJob(jobID)
	if jobErr != nil {
		return nil, jobErr
	}
	instanceID := output.Instances[0]
	waitErr := c.WaitInstanceStatus(instanceID, INSTANCE_STATUS_RUNNING)
	if waitErr != nil {
		return nil, waitErr
	}
	ins, waitErr := c.waitInstanceNetwork(instanceID)
	if waitErr != nil {
		return nil, waitErr
	}
	return ins, nil
}
func (c *client) DescribeInstance(instanceID string) (*qcservice.Instance, error) {
	input := &qcservice.DescribeInstancesInput{Instances: []string{instanceID}}
	output, err := c.instanceService.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	if len(output.InstanceSet) == 0 {
		return nil, fmt.Errorf("Instance with id [%s] not exist.", instanceID)
	}
	return output.InstanceSet[0], nil
}

func (c *client) StartInstance(instanceID string) error {
	input := &qcservice.StartInstancesInput{Instances: []string{instanceID}}
	output, err := c.instanceService.StartInstances(input)
	if err != nil {
		return err
	}
	jobID := output.JobID
	waitErr := c.waitJob(jobID)
	if waitErr != nil {
		return waitErr
	}
	return c.WaitInstanceStatus(instanceID, INSTANCE_STATUS_RUNNING)
}

func (c *client) StopInstance(instanceID string, force bool) error {
	var forceParam int
	if force {
		forceParam = 1
	} else {
		forceParam = 0
	}
	input := &qcservice.StopInstancesInput{Instances: []string{instanceID}, Force: forceParam}
	output, err := c.instanceService.StopInstances(input)
	if err != nil {
		return err
	}
	jobID := output.JobID
	waitErr := c.waitJob(jobID)
	if waitErr != nil {
		return waitErr
	}
	return c.WaitInstanceStatus(instanceID, INSTANCE_STATUS_STOPPED)
}

func (c *client) RestartInstance(instanceID string) error {
	input := &qcservice.RestartInstancesInput{Instances: []string{instanceID}}
	output, err := c.instanceService.RestartInstances(input)
	if err != nil {
		return err
	}
	jobID := output.JobID
	waitErr := c.waitJob(jobID)
	if waitErr != nil {
		return waitErr
	}
	return c.WaitInstanceStatus(instanceID, INSTANCE_STATUS_RUNNING)
}

func (c *client) TerminateInstance(instanceID string) error {
	input := &qcservice.TerminateInstancesInput{Instances: []string{instanceID}}
	output, err := c.instanceService.TerminateInstances(input)
	if err != nil {
		return err
	}
	jobID := output.JobID
	waitErr := c.waitJob(jobID)
	if waitErr != nil {
		return waitErr
	}
	return c.WaitInstanceStatus(instanceID, INSTANCE_STATUS_TERMINATED)
}

func (c client) CreateKeyPair(keyPairName string, publicKey string) (string, error) {
	log.Debugf("Create KeyPair name: [%s], publicKey: [%s]", keyPairName, publicKey)
	input := &qcservice.CreateKeyPairInput{Mode: "user", KeyPairName: keyPairName, PublicKey: publicKey}
	output, err := c.keypairService.CreateKeyPair(input)
	if err != nil {
		return "", err
	}
	return output.KeyPairID, nil
}

func (c client) DescribeKeyPair(keyPairID string) (*qcservice.KeyPair, error) {
	input := &qcservice.DescribeKeyPairsInput{KeyPairs: []string{keyPairID}}
	output, err := c.keypairService.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	if len(output.KeyPairSet) == 0 {
		return nil, fmt.Errorf("KeyPair with id [%s] not exist.", keyPairID)
	}
	return output.KeyPairSet[0], nil
}

func (c client) DeleteKeyPair(keyPairID string) error {
	input := &qcservice.DeleteKeyPairsInput{KeyPairs: []string{keyPairID}}
	_, err := c.keypairService.DeleteKeyPairs(input)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) waitJob(jobID string) error {
	log.Debugf("Waiting for Job [%s] finished", jobID)
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		input := &qcservice.DescribeJobsInput{Jobs: []string{jobID}}
		output, err := c.jobService.DescribeJobs(input)
		if err != nil {
			return false, err
		}
		if len(output.JobSet) == 0 {
			return false, fmt.Errorf("Can not find job [%s]", jobID)
		}
		j := output.JobSet[0]
		if j.Status == "failed" {
			return false, fmt.Errorf("Job [%s] failed", jobID)
		}
		return true, nil
	}, (c.opTimeout / 5), 5*time.Second)
}

func (c *client) WaitInstanceStatus(instanceID string, status string) error {
	log.Debugf("Waiting for Instance [%s] status [%s] ", instanceID, status)
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		i, err := c.DescribeInstance(instanceID)
		if err != nil {
			return false, err
		}
		if i.Status == status {
			if i.TransitionStatus != "" {
				//wait transition to finished
				return false, nil
			}
			log.Debugf("Instance [%s] status is [%s] ", instanceID, i.Status)
			return true, nil
		}
		return false, nil
	}, (c.opTimeout / 5), 5*time.Second)
}

func (c *client) waitInstanceNetwork(instanceID string) (*qcservice.Instance, error) {
	log.Debugf("Waiting for IP address to be assigned to Instance [%s]", instanceID)
	var ins *qcservice.Instance
	err := mcnutils.WaitForSpecificOrError(func() (bool, error) {
		i, err := c.DescribeInstance(instanceID)
		if err != nil {
			return false, err
		}
		if len(i.VxNets) == 0 || i.VxNets[0].PrivateIP == "" {
			return false, nil
		}
		ins = i
		log.Debugf("Instance [%s] get IP address [%s]", instanceID, ins.VxNets[0].PrivateIP)
		return true, nil
	}, (c.opTimeout / 5), 5*time.Second)
	return ins, err
}
