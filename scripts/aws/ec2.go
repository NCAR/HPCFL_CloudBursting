package aws

import (
	"log"
	"os/exec"
	"strings"
)

//EC2Instance represents an EC2 compute instance
type EC2Instance struct {
	name      string //`json:"nodeName"`
	privateIP string //`json:"privateIP"`
	publicIP  string //`json:"publicIP"`
}

//New creates and returns a new EC2Instance
//NOTE: Does not create or setup instance
func New(name, privateIP, publicIP string) EC2Instance {
	return EC2Instance{name: name, privateIP: privateIP, publicIP: publicIP}
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

//Setup does all the provisioning neccesary to setup the instance
//TODO return error when it makes sense to
func (i EC2Instance) Setup() error {
	//TODO: remove old key from ~/.ssh/known_hosts if exists
	if strings.HasPrefix(i.Name(), "router") {
		log.Printf("aws: Setting up a router instance\n")
		err := exec.Command(AWS_DIR+"wgInstall.sh", i.Name(), i.PublicIP()).Run()
		if err != nil {
			log.Printf("ERROR:aws: error setting up instance %s %s\n", i.Name(), err)
		}
	} else if strings.HasPrefix(i.Name(), "aws") {
		log.Printf("aws: Setting up a compute instance\n")
		//TODO do some error handling
		// must be able to deal with the exit status of script is 255 b/c of reboot command
		err := exec.Command(AWS_DIR+"nodeSetup.sh", i.Name(), AWS_DIR).Run()
		if err != nil {
			log.Printf("ERROR:aws: error setting up instance %s %s\n", i.Name(), err)
		}
	} else {
		log.Printf("ERROR:aws: Unknown node type %s\n", i.name)
	}
	return nil
}
