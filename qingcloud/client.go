package qingcloud

import (
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/yunify/qingcloud-sdk-go/config"
	"github.com/yunify/qingcloud-sdk-go/service/instance"
	"github.com/yunify/qingcloud-sdk-go/service/job"
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
	RunInstance(arg *RunInstanceArg) (*instance.Instance, error)
	DescribeInstance(instanceID string) (*instance.Instance, error)
	StartInstance(instanceID string) error
	StopInstance(instanceID string, force bool) error
	RestartInstance(instanceID string) error
	TerminateInstance(instanceID string) error
	WaitInstanceStatus(instanceID string, status string) error
}

func NewClient(config *config.Config, zone string) (Client, error) {
	instanceService, err := instance.Init(config)
	if err != nil {
		return nil, err
	}
	jobService, err := job.Init(config)
	if err != nil {
		return nil, err
	}
	c := &client{
		instanceService: instanceService,
		jobService:      jobService,
		opTimeout:       defaultOpTimeout,
		zone:            zone,
	}
	return c, nil
}

type client struct {
	instanceService *instance.InstanceService
	jobService      *job.JobService
	opTimeout       int
	zone            string
}

type RunInstanceArg struct {
	CPU          int
	Memory       int
	ImageID      string
	LoginKeyPair string
	VxNet        string
}

func (c *client) RunInstance(arg *RunInstanceArg) (*instance.Instance, error) {
	input := &instance.RunInstancesInput{
		CPU:          arg.CPU,
		Count:        1,
		ImageID:      arg.ImageID,
		Memory:       arg.Memory,
		LoginKeyPair: arg.LoginKeyPair,
		LoginMode:    "keypair",
		//SecurityGroup string   `json:"security_group" name:"security_group" location:"requestParams"`
		VxNets: []string{arg.VxNet},
		//Volumes       []string `json:"volumes" name:"volumes" location:"requestParams"`
		Zone: c.zone,
	}

	output, err := c.instanceService.RunInstances(input)
	if err != nil {
		return nil, err
	}
	if len(output.Instances) == 0 {
		return nil, errors.New("create instance response error.")
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
func (c *client) DescribeInstance(instanceID string) (*instance.Instance, error) {
	input := &instance.DescribeInstancesInput{Instances: []string{instanceID}, Zone: c.zone}
	output, err := c.instanceService.DescribeInstances(input)
	if err != nil {
		return nil, err
	}
	if len(output.InstanceSet) == 0 {
		return nil, fmt.Errorf("instance with id [%s] not exist.", instanceID)
	}
	return output.InstanceSet[0], nil
}

func (c *client) StartInstance(instanceID string) error {
	input := &instance.StartInstancesInput{Instances: []string{instanceID}, Zone: c.zone}
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
	input := &instance.StopInstancesInput{Instances: []string{instanceID}, Force: forceParam, Zone: c.zone}
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
	input := &instance.RestartInstancesInput{Instances: []string{instanceID}, Zone: c.zone}
	_, err := c.instanceService.RestartInstances(input)
	if err != nil {
		return err
	}
	//TODO wait jobid
	//jobID := output.
	//return c.WaitJob(jobID)
	return nil
}

func (c *client) TerminateInstance(instanceID string) error {
	input := &instance.TerminateInstancesInput{Instances: []string{instanceID}, Zone: c.zone}
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

func (c *client) waitJob(jobID string) error {
	log.Debugf("Waiting for Job [%s] finished", jobID)
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		input := &job.DescribeJobsInput{Jobs: []string{jobID}, Zone: c.zone}
		output, err := c.jobService.DescribeJobs(input)
		if err != nil {
			return false, err
		}
		if len(output.JobSet) == 0 {
			return false, fmt.Errorf("can not find job [%s]", jobID)
		}
		j := output.JobSet[0]
		if j.Status == "failed" {
			return false, fmt.Errorf("job [%s] failed", jobID)
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

func (c *client) waitInstanceNetwork(instanceID string) (*instance.Instance, error) {
	log.Debugf("Waiting for IP address to be assigned to Instance [%s]", instanceID)
	var ins *instance.Instance
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
