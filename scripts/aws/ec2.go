package aws

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

//EC2Instance represents an EC2 compute instance
type EC2Instance struct {
	name      string
	privateIP string
	publicIP  string
}

//New creates and returns a new EC2Instance
//NOTE: Does not create or setup actual cloud resources
func New(name, privateIP, publicIP string) EC2Instance {
	return EC2Instance{name: name, privateIP: privateIP, publicIP: publicIP}
}

//Setup does all the provisioning neccesary to setup the instance
func (i EC2Instance) Setup() error {
	//TODO: remove old key from ~/.ssh/known_hosts if exists
	switch {
	case strings.HasPrefix(i.Name(), "aws"):
		log.Printf("aws: Setting up a compute instance\n")
		// must be able to deal with the exit status of script is 255 b/c of reboot command
		err := exec.Command(AWS_DIR+"nodeSetup.sh", i.Name(), AWS_DIR).Run()
		if err != nil && !strings.Contains(err.Error(), "255") {
			return fmt.Errorf("aws: could not setup instance %s due to %v", i.Name(), err)
		}
		/*
			case strings.HasPrefix(i.Name(), "router"):
				log.Printf("aws: Setting up a router instance\n")
				//need root to run wgInstall b/c it sets up wireguard vpn interface
				err := exec.Command(AWS_DIR+"wgInstall.sh", i.Name(), i.PublicIP()).Run()
				if err != nil {
					log.Printf("ERROR:aws: error setting up instance %s %s\n", i.Name(), err)
				}
		*/
	default:
		return errors.New("aws: Unable to setup unrecognized node " + i.name)
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

//PublicIP returns the publicIP of the instance
//If it has none an empty string is returned
func (i EC2Instance) PublicIP() string {
	if i.publicIP == "" {
		log.Printf("ERROR: Attempting to get public IP of %s, which is not know\n", i.name)
	}
	return i.publicIP
}
