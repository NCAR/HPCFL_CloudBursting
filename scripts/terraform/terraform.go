/*
Copyright (c) 2020, University Corporation for Atmospheric Research
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors
may be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY,
WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

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
	return nil
}

//Info returns all of the information terraform knows about the given instance name
func Info(name string) (Instance, error) {
	//setup and run command
	cmd := exec.Command(utils.Config("terraform.dir").Self()+"terraform", "output", "-state="+utils.Config("terraform.dir").Self()+"terraform.tfstate", "-json", "-no-color", name)
	configCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not get info for %s %s", name, err)
	}
	part := partition(name)
	if part == nil {
		return nil, fmt.Errorf("Can't find partition containing %s", name)
	}
	switch part.Lookup("type").Self() {
	case "aws":
		//json field names defined in ec2Instance.tmpl
		var instance struct {
			Name      string `json:"nodeName"`
			PrivateIP string `json:"privateIP"`
			PublicIP  string `json:"publicIP"`
		}
		//[]byte conversion is redundant
		if json.Unmarshal([]byte(out), &instance) != nil {
			return nil, fmt.Errorf("could not decode info for %s %s", name, err)
		}
		return aws.New(instance.Name, instance.PrivateIP, instance.PublicIP, part.Self()), nil

	default:
		return nil, fmt.Errorf("unrecognized node type %s for node %s", part.Lookup("type").Self(), name)

	}
}

//Del deletes the config for the given instance
//NOTE: Update must be called afterwards to update the cloud infrastructure
func Del(name string) {
	//TODO if Info takes a while maybe check this is neccessary before doing
	// e.g. check n has a teardown script
	n, err := Info(name)
	if err != nil {
		log.Printf("ERROR:terraform: Could not find node %s\n", name)
	}
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
	tries := 0
	for out, err := cmd.CombinedOutput(); err != nil; out, err = cmd.CombinedOutput() {
		log.Printf("ERROR:terraform: Problem updating cloud resources %s %s\n", out, err)
		time.Sleep(time.Second * 5)
		log.Printf("Info:terraform: Trying command again\n")
		cmd = exec.Command(utils.Config("terraform.dir").Self()+"terraform", "apply", "-auto-approve", "-state="+utils.Config("terraform.dir").Self()+"terraform.tfstate", "-lock=true", "-input=false", utils.Config("terraform.tf_files").Self())
		configCmd(cmd)
		tries++
		if tries > 25 {
			log.Printf("CRITICAL:terraform: Could not update cloud resources\n")
		}
	}
}

func instanceNames() []string {
	matches, err := filepath.Glob(utils.Config("terraform.tf_files").Self() + "*.tf")
	if err != nil {
		log.Fatal(err)
	}
	names := []string{}
	for _, match := range matches {
		_, file := filepath.Split(match)
		if strings.Compare("infra.tf", file) == 0 {
			continue
		}
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
		ip, err := ip(name)
		if err != nil {
			return fmt.Errorf("unable to add node %s", err)
		}
		cf, err := confFunc(part.Lookup("template").Self(), name)
		if err != nil {
			return fmt.Errorf("unable to add node %s: %s", name, err)
		}
		return cf(aws.New(name, ip, "", part.Self()))
	default:
		return fmt.Errorf("don't know how to add instance %s", name)
	}
}

//add populates the given template with the given Instance
func confFunc(tmpl, name string) (func(interface{}) error, error) {
	n := strings.Split(tmpl, "/")
	t, err := template.New(n[len(n)-1]).ParseFiles(tmpl)
	if err != nil {
		return nil, fmt.Errorf("could not open template file %s", tmpl)
	}
	fh, err := os.Create(utils.Config("terraform.tf_files").Self() + name + ".tf")
	if err != nil {
		return nil, fmt.Errorf("error creating config file for instance %s %s", name, err)
	}
	return (func(inst interface{}) error { err := t.Execute(fh, inst); fh.Close(); return err }), nil
}

func ip(name string) (string, error) {
	addr, err := net.LookupHost(name)
	if err != nil {
		return "", fmt.Errorf("could not find ip for host %s, %s", name, err)
	}
	return addr[0], nil
}

//On returns a list of nodes terraform thinks are on
func On() []string {
	return instanceNames()
}
