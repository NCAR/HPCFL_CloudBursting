package terraform

import (
	"encoding/json"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/aws"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"text/template"
)

func Info(name string) aws.EC2Instance {
	cmd := exec.Command(TERRAFORM_DIR+"terraform", "output", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-json", "-no-color", name)
	cmd.Dir = TERRAFORM_DIR
	cmd.Env = make([]string, 2)
	cmd.Env[0] = "PWD=" + TERRAFORM_DIR
	cmd.Env[1] = "HOME=/var/lib/slurm"
	out, err := cmd.Output()
	if err != nil {
		log.Printf("ERROR:terraform: Could not get info for %s %s\n", name, err)
	}
	instance := aws.EC2Instance{}
	err = json.Unmarshal([]byte(out), &instance) //[]byte conversion is redundant
	if err != nil {
		log.Printf("ERROR:terraform: Could not decode info for %s %s\n", name, err)
	}
	return instance
}

func instanceNames() []string {
	matches, err := filepath.Glob(CONFIG_DIR + "aws*.tf")
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

func allInfo() []aws.EC2Instance {
	insts := instanceNames()
	instances := []aws.EC2Instance{}
	for _, inst := range insts {
		instances = append(instances, Info(inst))
	}
	return instances
}

func nextOpenNode() aws.EC2Instance {
	currInsts := instanceNames()
	sort.Strings(currInsts)
	for i := 0; i < 10; i++ {
		if i >= len(currInsts) || currInsts[i] != "aws"+strconv.Itoa(i) {
			return aws.New("aws"+strconv.Itoa(i), "192.168.2."+strconv.Itoa(50+i))
		}
	}
	return aws.EC2Instance{} //TODO handle no open names for instances
}

func add(inst aws.EC2Instance) aws.EC2Instance {
	t, err := template.New("ec2Instance.tmpl").ParseFiles(TERRAFORM_DIR + "ec2Instance.tmpl")
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

func Add(inst string) aws.EC2Instance {
	i, err := strconv.Atoi(inst[3:])
	if err != nil {
		log.Printf("ERROR:terraform: could not add instance %s\n", inst)
	}
	return add(aws.New(inst, "192.168.2."+strconv.Itoa(50+i)))
}

func AddEC2() aws.EC2Instance {
	//AddEC2 creates the config for a new ec2 instance
	//It does not create the instance, one must run Update() to update running infrastructure
	return add(nextOpenNode())
}

func Del(name string) {
	log.Printf("DEBUG:terraform: deleting config file %s\n", CONFIG_DIR+name+".tf")
	/*err :=*/ os.Remove(CONFIG_DIR + name + ".tf")
	/*if err != nil {
	    log.Fatal(err)
	}*/
}

func Update() {
	cmd := exec.Command(TERRAFORM_DIR+"terraform", "apply", "-auto-approve", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-lock=true", "-input=false", CONFIG_DIR)
	cmd.Dir = TERRAFORM_DIR
	cmd.Env = make([]string, 2)
	cmd.Env[0] = "PWD=" + TERRAFORM_DIR
	cmd.Env[1] = "HOME=/var/lib/slurm"
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("ERROR:terraform: Problem updating cloud resources %s %s\n", out, err)
	}
}

func Stop() {
	cmd := exec.Command(TERRAFORM_DIR+"terraform", "destroy", "-auto-approve", "-state="+TERRAFORM_DIR+"terraform.tfstate", "-lock=true", "-input=false", CONFIG_DIR)
	cmd.Dir = TERRAFORM_DIR
	cmd.Env = make([]string, 2)
	cmd.Env[0] = "PWD=" + TERRAFORM_DIR
	cmd.Env[1] = "HOME=/var/lib/slurm"
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	for _, inst := range instanceNames() {
		err := os.Remove(CONFIG_DIR + inst + ".tf")
		if err != nil {
			log.Fatal(err)
		}
	}
}