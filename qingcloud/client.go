package qingcloud

import (
	"errors"
	"fmt"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/yunify/qingcloud-sdk-go/service/instance"
	"github.com/yunify/qingcloud-sdk-go/service/job"
	"time"
)

type Client interface {
	RunInstance(d *Driver) (string, error)
	//GetInstanceState(d *Driver) (string, error)
	//StartInstance(d *Driver) error
	//StopInstance(d *Driver) error
	//RestartInstance(d *Driver) error
	//DeleteInstance(d *Driver) error
	//WaitForInstanceStatus(d *Driver, status string) error
	//WaitForInstanceNetwork(d *Driver)
	//GetInstanceIPAddresses(d *Driver) ([]IPAddress, error)
	//GetPublicKey(keyPairName string) ([]byte, error)
	//CreateKeyPair(d *Driver, name string, publicKey string) error
	//DeleteKeyPair(d *Driver, name string) error
	//GetNetworkID(d *Driver) (string, error)
	//GetImageID(d *Driver) (string, error)
}

func NewClient(d *Driver) (Client, error) {
	config := d.Config()
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
	}
	return c, nil
}

type client struct {
	instanceService *instance.InstanceService
	jobService      *job.JobService
	opTimeout       int
}

func (c *client) RunInstance(d *Driver) (string, error) {
	input := &instance.RunInstancesInput{
		CPU:          d.CPU,
		Count:        1,
		ImageID:      d.Image,
		Memory:       d.Memory,
		LoginKeyPair: d.SSHKeyID,
		LoginMode:    "keypair",
		//SecurityGroup string   `json:"security_group" name:"security_group" location:"requestParams"`
		VxNets: []string{d.Vxnet},
		//Volumes       []string `json:"volumes" name:"volumes" location:"requestParams"`
		Zone: d.Zone,
	}

	output, err := c.instanceService.RunInstances(input)
	if err != nil {
		return "", err
	}
	if len(output.Instances) == 0 {
		return "", errors.New("create instance response error.")
	}
	jobID := output.JobID
	jobErr := c.WaitJob(jobID)
	if jobErr != nil {
		return "", jobErr
	}
	instanceID := output.Instances[0]
	waitErr := c.WaitInstanceNetwork(instanceID)
	if waitErr != nil {
		return "", waitErr
	}
	return instanceID, nil
}

func (c *client) WaitJob(jobID string) error {
	log.Infof("Waiting for Job [%s] finished", jobID)
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		input := &job.DescribeJobsInput{Jobs: []string{jobID}}
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

func (c *client) WaitInstanceNetwork(instanceID string) error {
	log.Infof("Waiting for IP address to be assigned to Instance [%s]", instanceID)
	return mcnutils.WaitForSpecificOrError(func() (bool, error) {
		input := &instance.DescribeInstancesInput{Instances: []string{instanceID}}
		output, err := c.instanceService.DescribeInstances(input)
		if err != nil {
			return false, err
		}
		if len(output.InstanceSet) == 0 {
			return false, errors.New("describe instance response error.")
		}
		i := output.InstanceSet[0]
		if len(i.VxNets) == 0 || i.VxNets[0].PrivateIP == "" {
			return false, nil
		}
		return true, nil
	}, (c.opTimeout / 5), 5*time.Second)
}
