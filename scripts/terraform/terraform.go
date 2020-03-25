package terraform

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/aws"
)

//Instance represents a cloud instance
type Instance interface {
	Name() string
	IP() string
	String() string
	Setup() error
}

//configCmd sets up the environment to run terraform commands
func configCmd(cmd *exec.Cmd) {
	cmd.Dir = TERRAFORM_DIR
	cmd.Env = make([]string, 2)
	cmd.Env[0] = "PWD=" + TERRAFORM_DIR
	cmd.Env[1] = "HOME=/var/lib/slurm"
}

//Info returns all of the information terraform knows about the given instance name
func Info(name string) Instance {
	//setup and run command
	cmd := exec.Command(TERRAFORM_DIR+"terraform", "output", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-json", "-no-color", name)
	configCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR:terraform: Could not get info for %s %s\n", name, err)
	}

	//Demux via instance name
	//read cmd output json into appropriate struct type, and return it
	switch {
	case strings.HasPrefix(name, "aws"):
		//json field names defined in ec2Instance.tmpl
		var instance struct {
			Name      string `json:"nodeName"`
			PrivateIP string `json:"privateIP"`
			PublicIP  string `json:"publicIP"`
		}
		err = json.Unmarshal([]byte(out), &instance) //[]byte conversion is redundant
		if err != nil {
			log.Printf("ERROR:terraform: Could not decode info for %s %s\n", name, err)
		}
		return aws.New(instance.Name, instance.PrivateIP, instance.PublicIP)

	default:
		log.Printf("Error: Don't know how to get info for instance %s\n", name)
		return nil
	}
}

//Del deletes the config for the given instance
//NOTE: Update must be called afterwards to update the cloud infrastructure
func Del(name string) {
	log.Printf("DEBUG:terraform: deleting config file %s\n", CONFIG_DIR+name+".tf")
	os.Remove(CONFIG_DIR + name + ".tf")
}

//Update updates the running cloud infrastructure to reflect the current config files
//Should be run after calls to Add or Del
func Update() {
	cmd := exec.Command(TERRAFORM_DIR+"terraform", "apply", "-auto-approve", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-lock=true", "-input=false", CONFIG_DIR)
	configCmd(cmd)
	//TODO put a limit on the number of retries
	for out, err := cmd.CombinedOutput(); err != nil; out, err = cmd.CombinedOutput() {
		log.Printf("ERROR:terraform: Problem updating cloud resources %s %s\n", out, err)
		time.Sleep(time.Second * 5)
		log.Printf("Info:terraform: Trying command again\n")
		cmd = exec.Command(TERRAFORM_DIR+"terraform", "apply", "-auto-approve", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-lock=true", "-input=false", CONFIG_DIR)
		configCmd(cmd)
	}
}

//TODO test
func instanceNames() []string {
	matches, err := filepath.Glob(CONFIG_DIR + "[!infra].tf")
	if err != nil {
		log.Fatal(err)
	}
	names := []string{}
	for _, match := range matches {
		_, file := filepath.Split(match)
		names = append(names, file[:len(file)-3])
	}
	return names
}

//Stop stops all running cloud infrastructure and deletes config for all compute instances
//Note: currently untested but probably works
func Stop() {
	for _, inst := range instanceNames() {
		Del(inst)
	}
	Update()
}

//Add creates a config file for the given instance, but does not create it
//NOTE: Update must be called afterwards to update the cloud infrastructure
//NOTE: instance name number must be less than 205
func Add(name string) Instance {
	switch {
	case strings.HasPrefix(name, "aws"):
		ip := ip(name)
		if ip == "" {
			log.Printf("Error: Unable to add node %s\n", name)
			return nil
		}
		return add("ec2Instance.tmpl", aws.New(name, ip, ""))
	default:
		log.Printf("Error: don't know how to add instance %s\n", name)
		return nil
	}
}

//add populates the given template with the given Instance
func add(tmpl string, inst Instance) Instance {
	t, err := template.New(tmpl).ParseFiles(TERRAFORM_DIR + "scripts/terraform/" + tmpl)
	if err != nil {
		log.Printf("ERROR:terraform: Could not open config file template file\n")
	}
	fh, err := os.Create(CONFIG_DIR + inst.Name() + ".tf")
	if err != nil {
		log.Printf("ERROR:terraform: Error creating config file for instance %s %s\n", inst.Name(), err)
	}
	defer fh.Close()
	err = t.Execute(fh, inst)
	if err != nil {
		log.Printf("ERROR:terraform: Could not write to config file for instance %s %s\n", inst.Name(), err)
	}
	return inst
}

func ip(name string) string {
	addr, err := net.LookupHost(name)
	if err != nil {
		log.Printf("Error: Could not find ip for host %s\n", name)
		return ""
	}
	return addr[0]
}
