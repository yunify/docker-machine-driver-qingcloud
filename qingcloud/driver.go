package qingcloud

import (
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/yunify/qingcloud-sdk-go/config"
)

const (
	defaultSSHPort   = 22
	defaultSSHUser   = "root"
	defaultImage     = "trustysrvx64h"
	defaultZone      = "pek3a"
	defaultCPU       = 1
	defaultMemory    = 1024
	defaultOpTimeout = 180 //second
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
	Vxnet           string
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

	client, err := NewClient(d)
	if err != nil {
		return err
	}
	instanceID, err := client.RunInstance(d)
	if err != nil {
		return err
	}
	log.Infof("Created Instance %s",
		instanceID)

	return nil
}

func (d *Driver) createSSHKey() (*SSHKeyPair, error) {
	//TODO
	return nil, nil
}

func (d *Driver) GetIP() (string, error) {
	return "", nil
}

// GetMachineName returns the name of the machine
func (d *Driver) GetMachineName() string {
	return ""
}

// GetSSHKeyPath returns key path for use with ssh
func (d *Driver) GetSSHKeyPath() string {
	return ""
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	return defaultSSHPort, nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	return defaultSSHUser
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	return "", nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return state.None, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return nil
}

// Start a host
func (d *Driver) Start() error {
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return nil
}
