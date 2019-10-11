package aws

import(
	"os/exec"
	"strings"
	"log"
)

type EC2Instance struct {
    NameStr string `json:"nodeName"`
    PrivateIPStr string `json:"privateIP"`
    PublicIPStr string `json:"publicIP"`
}

func New(name, ip string) EC2Instance{
	return EC2Instance{NameStr:name, PrivateIPStr:ip}
} 

func (i EC2Instance) String() string {
	return "ec2Instance " + i.NameStr + ":" + i.PrivateIPStr + " Public:" + i.PublicIPStr
}

func (i EC2Instance) IP() string {
	return i.PrivateIPStr
}

func (i EC2Instance) Name() string {
	return i.NameStr
} 

func (i EC2Instance) PublicIP() string {
	if i.PublicIPStr == "" {
		log.Printf("ERROR: Attempting to get public IP of %s, which is not know\n", i.NameStr)
	}
	return i.PublicIPStr
}

func (i EC2Instance) Setup() {
	//TODO: remove old key from ~/.ssh/known_hosts if exists
	if strings.HasPrefix(i.NameStr, "router"){
		log.Printf("aws: Setting up a router instance\n")
		err := exec.Command(AWS_DIR+"wgInstall.sh", i.NameStr, i.PublicIPStr).Run()
		if err != nil {
			log.Printf("ERROR:aws: error setting up instance %s\n", err)
		}
	}else if strings.HasPrefix(i.NameStr, "aws"){
		log.Printf("aws: Setting up a compute instance\n")
		//TODO do some error handling
		// must be able to deal with the exit status of script is 255 b/c of reboot command
		err := exec.Command(AWS_DIR+"nodeSetup.sh", i.NameStr, AWS_DIR).Run()
		if err != nil {
			log.Printf("ERROR:aws: error setting up instance %s\n", err)
		}
	}else{
		log.Printf("ERROR:aws: Unknown node type %s\n", i.NameStr)	
	}
}
