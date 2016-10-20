package qingcloud

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/yunify/qingcloud-sdk-go/config"
	"github.com/yunify/qingcloud-sdk-go/service/instance"
)

const (
	defaultImage     = "trustysrvx64h"
	defaultZone      = "pek3a"
	defaultCPU       = 1
	defaultMemory    = 1024
	defaultOpTimeout = 180 //second
	dockerPort       = 2376
	swarmPort        = 3376
)

type Driver struct {
	*drivers.BaseDriver
	AccessKeyID     string
	SecretAccessKey string
	Zone            string
	Image           string
	CPU             int
	Memory          int
	SSHKeyID        string
	VxNet           string
	InstanceID      string
	client          Client
}

type SSHKeyPair struct {
	ID string
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "QINGCLOUD_ACCESS_KEY_ID",
			Name:   "qy_access_key_id",
			Usage:  "qingcloud access key id",
		},
		mcnflag.StringFlag{
			EnvVar: "QINGCLOUD_SECRET_ACCESS_KEY",
			Name:   "qy_secret_access_key",
			Usage:  "qingcloud secret access key",
		},
	}
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		Image:  defaultImage,
		CPU:    defaultCPU,
		Memory: defaultMemory,
		Zone:   defaultZone,
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "qingcloud"
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessKeyID = flags.String("qy_access_key_id")
	d.SecretAccessKey = flags.String("qy_secret_access_key")
	return nil
}

func (d *Driver) Config() *config.Config {
	config := config.New(d.AccessKeyID, d.SecretAccessKey)
	return config
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	if d.SSHKeyID == "" {
		key, err := d.createSSHKey()
		if err != nil {
			return err
		}
		d.SSHKeyID = key.ID
	}

	log.Infof("Creating Qingcloud Instance...")

	client := d.GetClient()
	arg := &RunInstanceArg{
		CPU:          d.CPU,
		Memory:       d.Memory,
		ImageID:      d.Image,
		VxNet:        d.VxNet,
		LoginKeyPair: d.SSHKeyID,
	}
	ins, err := client.RunInstance(arg)
	if err != nil {
		return err
	}
	d.InstanceID = ins.InstanceID
	d.IPAddress = ins.VxNets[0].PrivateIP
	d.MachineName = ins.InstanceID

	log.Infof("Created Instance %s",
		ins.InstanceID)

	return nil
}

func (d *Driver) GetClient() Client {
	if d.client == nil {
		client, err := NewClient(d.Config(), d.Zone)
		if err != nil {
			panic(fmt.Sprintf("init client error: %s", err.Error()))
		}
		d.client = client
	}
	return d.client
}

func (d *Driver) getInstance() (*instance.Instance, error) {
	return d.GetClient().DescribeInstance(d.InstanceID)
}

func (d *Driver) createSSHKey() (*SSHKeyPair, error) {
	//TODO
	return nil, nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	i, err := d.getInstance()
	if err != nil {
		return state.None, err
	}
	switch i.Status {
	case "pending":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "stopped":
		return state.Stopped, nil
	case "suspended", "terminated", "ceased":
		return state.Error, nil
	}
	return state.Error, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return d.GetClient().StopInstance(d.InstanceID, true)
}

// Remove a host
func (d *Driver) Remove() error {
	return d.GetClient().TerminateInstance(d.InstanceID)
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return d.GetClient().RestartInstance(d.InstanceID)
}

// Start a host
func (d *Driver) Start() error {
	return d.GetClient().StartInstance(d.InstanceID)
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return d.GetClient().StopInstance(d.InstanceID, false)
}
