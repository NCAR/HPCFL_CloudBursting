package terraform

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/aws"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

//Instance represents a cloud instance
type Instance interface {
	Name() string
	IP() string
	String() string
	Setup() error
	Teardown() error
}

//configCmd sets up the environment to run terraform commands
func configCmd(cmd *exec.Cmd) {
	cmd.Dir = utils.Config("terraform.dir").Self()
	cmd.Env = make([]string, 2)
	cmd.Env[0] = "PWD=" + utils.Config("terraform.dir").Self()
	cmd.Env[1] = "HOME=/var/lib/slurm"
}

func partition(name string) utils.Elem {
	//figure out what partition node is from
	for _, part := range utils.Config("partitions").Contains() {
		regex := part.Lookup("regex")
		if regex == nil {
			log.Printf("CRITICAL: could not find regex for partition %s\n", part.Self())
			continue
		}
		match, err := regexp.Match(regex.Self(), []byte(name))
		if err != nil {
			log.Printf("WARNING:terraform: error while finding %s's partition: %s\n", name, err)
		}
		if match {
			return part
		}
	}
	log.Printf("ERROR:terraform: Could not find partition for node %s\n", name)
	return nil
}

//Info returns all of the information terraform knows about the given instance name
func Info(name string) Instance {
	//setup and run command
	cmd := exec.Command(utils.Config("terraform.dir").Self()+"terraform", "output", "-state="+utils.Config("terraform.dir").Self()+"terraform.tfstate", "-json", "-no-color", name)
	configCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR:terraform: Could not get info for %s %s\n", name, err)
	}
	part := partition(name)
	if part == nil {
		log.Printf("Error: Don't know how to get info for instance %s\n", name)
		return nil
	}
	switch part.Lookup("type").Self() {
	case "aws":
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
		return aws.New(instance.Name, instance.PrivateIP, instance.PublicIP, part.Self())

	default:
		log.Printf("ERROR:terraform: Unrecognized node type %s for node %s\n", part.Lookup("type").Self(), name)
		return nil

	}
}

//Del deletes the config for the given instance
//NOTE: Update must be called afterwards to update the cloud infrastructure
func Del(name string) {
	//TODO if Info takes a while maybe check this is neccessary before doing
	// e.g. check n has a teardown script
	n := Info(name)
	if n != nil {
		n.Teardown()
	}
	log.Printf("DEBUG:terraform: deleting config file %s\n", utils.Config("terraform.tf_files").Self()+name+".tf")
	os.Remove(utils.Config("terraform.tf_files").Self() + name + ".tf")
}

//Update updates the running cloud infrastructure to reflect the current config files
//Should be run after calls to Add or Del
func Update() {
	cmd := exec.Command(utils.Config("terraform.dir").Self()+"terraform", "apply", "-auto-approve", "-state="+utils.Config("terraform.dir").Self()+"terraform.tfstate", "-lock=true", "-input=false", utils.Config("terraform.tf_files").Self())
	configCmd(cmd)
	//TODO put a limit on the number of retries
	for out, err := cmd.CombinedOutput(); err != nil; out, err = cmd.CombinedOutput() {
		log.Printf("ERROR:terraform: Problem updating cloud resources %s %s\n", out, err)
		time.Sleep(time.Second * 5)
		log.Printf("Info:terraform: Trying command again\n")
		cmd = exec.Command(utils.Config("terraform.dir").Self()+"terraform", "apply", "-auto-approve", "-state="+utils.Config("terraform.dir").Self()+"terraform.tfstate", "-lock=true", "-input=false", utils.Config("terraform.tf_files").Self())
		configCmd(cmd)
	}
}

//TODO test
func instanceNames() []string {
	matches, err := filepath.Glob(utils.Config("terraform.tf_files").Self() + "[!infra].tf")
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
func Add(name string) error {
	part := partition(name)
	switch part.Lookup("type").Self() {
	case "aws":
		ip := ip(name)
		if ip == "" {
			log.Printf("Error: Unable to add node %s\n", name)
			return nil
		}
		return confFunc(part.Lookup("template").Self(), name)(aws.New(name, ip, "", part.Self()))
	default:
		return fmt.Errorf("Error: don't know how to add instance %s\n", name)
	}
}

//add populates the given template with the given Instance
func confFunc(tmpl, name string) func(interface{}) error {
	n := strings.Split(tmpl, "/")
	t, err := template.New(n[len(n)-1]).ParseFiles(tmpl)
	if err != nil {
		log.Printf("ERROR:terraform: Could not open template file %s\n", tmpl)
	}
	fh, err := os.Create(utils.Config("terraform.tf_files").Self() + name + ".tf")
	if err != nil {
		log.Printf("ERROR:terraform: Error creating config file for instance %s %s\n", name, err)
	}
	return (func(inst interface{}) error { err := t.Execute(fh, inst); fh.Close(); return err })
}

func ip(name string) string {
	addr, err := net.LookupHost(name)
	if err != nil {
		log.Printf("Error: Could not find ip for host %s\n", name)
		return ""
	}
	return addr[0]
}
