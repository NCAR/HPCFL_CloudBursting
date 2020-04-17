package aws

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

//EC2Instance represents an EC2 compute instance
type EC2Instance struct {
	name      string
	privateIP string
	publicIP  string
	partition string
}

//New creates and returns a new EC2Instance
//NOTE: Does not create or setup actual cloud resources
func New(name, privateIP, publicIP, part string) EC2Instance {
	return EC2Instance{name: name, privateIP: privateIP, publicIP: publicIP, partition: part}
}

func (i EC2Instance) MakeConfig(mkfunc func(interface{}) error) error {
	return mkfunc(i)
}

//Setup does all the provisioning neccesary to setup the instance
func (i EC2Instance) Setup() error {
	log.Printf("DEBUG:aws: Setting up a compute instance\n")
	// must be able to deal with the exit status of script is 255 b/c of reboot command
	setup := utils.Config("partitions").Lookup(i.partition + ".setup").Self()
	if setup == "" {
		log.Printf("DEBUG:aws: No setup required\n")
		return nil
	}
	dirSplit := strings.SplitAfter(setup, "/")
	dir := strings.Join(dirSplit[:len(dirSplit)-1], "")
	err := exec.Command(setup, i.Name(), dir).Run()
	if err != nil && !strings.Contains(err.Error(), "255") {
		return fmt.Errorf("WARNING:aws: could not setup instance %s due to %v", i.Name(), err)
	}
	return nil
}

//Teardown does any cleanup neccessary to cleanup an instance
func (i EC2Instance) Teardown() error {
	log.Printf("DEBUG:aws: Tearing down a compute instance\n")
	// must be able to deal with the exit status of script is 255 b/c of reboot command
	setup := utils.Config("partitions").Lookup(i.partition + ".teardown").Self()
	if setup == "" {
		log.Printf("DEBUG:aws: No teardown required\n")
		return nil
	}
	dirSplit := strings.SplitAfter(setup, "/")
	dir := strings.Join(dirSplit[:len(dirSplit)-1], "")
	err := exec.Command(setup, i.Name(), dir).Run()
	if err != nil && !strings.Contains(err.Error(), "255") {
		return fmt.Errorf("WARNING:aws: could not teardown instance %s due to %v", i.Name(), err)
	}
	return nil
}

//String returns a string describing the given ec2 instance
func (i EC2Instance) String() string {
	if i.publicIP == "" {
		return "ec2Instance " + i.name + ":" + i.privateIP
	}
	return "ec2Instance " + i.name + ":" + i.privateIP + " Public:" + i.publicIP
}

// IP returns the private ip for the instance
func (i EC2Instance) IP() string {
	return i.privateIP
}

//Name returns the name of the instance
func (i EC2Instance) Name() string {
	return i.name
}

//AMI returns the instance ami
func (i EC2Instance) AMI() string {
	return utils.Config("partitions."+i.partition+".ami").Self()
}

//Size returns the instance size
func (i EC2Instance) Size() string {
	return utils.Config("partitions."+i.partition+".size").Self()
}

//PublicIP returns the publicIP of the instance
//If it has none an empty string is returned
func (i EC2Instance) PublicIP() string {
	if i.publicIP == "" {
		log.Printf("ERROR: Attempting to get public IP of %s, which is not know\n", i.name)
	}
	return i.publicIP
}
