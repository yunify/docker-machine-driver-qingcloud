package qingcloud

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/machine/libmachine/log"
	"github.com/yunify/qingcloud-sdk-go/config"
	sdklogger "github.com/yunify/qingcloud-sdk-go/logger"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"testing"
)

var loginKeyPair string
var vxNet string
var accessKeyID string
var secretAccessKey string
var zone string

func check(t *testing.T) {
	if loginKeyPair == "" || vxNet == "" || accessKeyID == "" || secretAccessKey == "" {
		t.Fatal("miss parameter.")
	}
}

func jsonString(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestClient(t *testing.T) {
	check(t)
	config, err := config.New(accessKeyID, secretAccessKey)
	if err != nil {
		t.Fatal(err)
	}
	sdklogger.SetLevel("debug")
	client, err := NewClient(config, zone)
	if err != nil {
		t.Fatal(err)
	}
	arg := &RunInstanceArg{
		CPU:          defaultCPU,
		Memory:       defaultMemory,
		ImageID:      defaultImage,
		LoginKeyPair: loginKeyPair,
		VxNet:        vxNet,
	}
	i, err := client.RunInstance(arg)
	if err != nil {
		t.Fatal(err)
	}
	instanceID := i.InstanceID
	fmt.Printf("run instance: %s\n", jsonString(i))
	i2, err := client.DescribeInstance(instanceID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("describe instance: %s\n", jsonString(i2))
	if i2.Status != "running" {
		t.Error("expect status running, but get ", i2.Status)
	}
	fmt.Printf("stoping instance: %s\n", instanceID)
	stopErr := client.StopInstance(instanceID, false)
	if stopErr != nil {
		t.Fatal(stopErr)
	}
	i3, err := client.DescribeInstance(instanceID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("describe instance: %s\n", jsonString(i3))
	if i3.Status != "stopped" {
		t.Error("expect status stopped, but get ", i3.Status)
	}
	fmt.Printf("starting instance: %s \n", instanceID)
	startErr := client.StartInstance(instanceID)
	if startErr != nil {
		t.Fatal(startErr)
	}
	i4, err := client.DescribeInstance(instanceID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("describe instance: %s\n", jsonString(i4))
	if i4.Status != "running" {
		t.Error("expect status running, but get ", i4.Status)
	}

	restartErr := client.RestartInstance(instanceID)
	if restartErr != nil {
		t.Fatal(restartErr)
	}

	fmt.Printf("terminate instance: %s\n", instanceID)
	delErr := client.TerminateInstance(instanceID)
	if delErr != nil {
		t.Fatal(delErr)
	}
	i5, err := client.DescribeInstance(instanceID)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("describe instance: %s\n", jsonString(i5))
	if i5.Status != "terminated" {
		t.Error("expect status terminated, but get ", i5.Status)
	}

}

func TestClientKeyPair(t *testing.T) {
	check(t)
	config, err := config.New(accessKeyID, secretAccessKey)
	if err != nil {
		t.Fatal(err)
	}
	//config.Services.IaaS.Host = "api.test.com"
	//config.Services.IaaS.Protocol = "http"
	//config.Services.IaaS.Port = 8880
	//client, err := NewClient(config, "allinone")
	client, err := NewClient(config, defaultZone)
	if err != nil {
		t.Fatal(err)
	}
	u, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	publicKey, err := ioutil.ReadFile(fmt.Sprintf("%s/.ssh/id_rsa.pub", u.HomeDir))
	if err != nil {
		t.Fatal(err)
	}
	publicKeyStr := strings.TrimSpace(string(publicKey))
	keyPairID, err := client.CreateKeyPair("test keypair", publicKeyStr)
	if err != nil {
		t.Fatal(err)
	}
	keyPair, err := client.DescribeKeyPair(keyPairID)
	if err != nil {
		t.Fatal(err)
	}
	println(keyPair.PubKey)
	err = client.DeleteKeyPair(keyPairID)
	if err != nil {
		t.Fatal(err)
	}

}

func init() {
	flag.StringVar(&loginKeyPair, "loginKeyPair", "", "loginKeyPair")
	flag.StringVar(&vxNet, "vxNet", "", "vxNet")
	flag.StringVar(&accessKeyID, "accessKeyID", "", "accessKeyID")
	flag.StringVar(&secretAccessKey, "secretAccessKey", "", "secretAccessKey")
	flag.StringVar(&zone, "zone", defaultZone, "zone")
	log.SetDebug(true)
}

func TestMain(m *testing.M) {
	flag.Parse()
	exit := m.Run()
	os.Exit(exit)
}
