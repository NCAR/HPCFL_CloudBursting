package main

import (
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"log"
	"os"
	//"github.com/NCAR/HPCFL_TerraformScripts/scripts/salt"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
)

func main() {
	//setup logging
	f, err := os.OpenFile("/var/lib/slurm/log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("ERROR: Could not open log file /var/lib/slurm/log, %s\n", err)
	}
	defer f.Close()
	log.SetOutput(f)

	//check was given an arg
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <instance name glob>\n", os.Args[0])
	}

	//TODO does this line do anything?
	os.Setenv("PWD", "/home/slurm/terraform")

	// clean up each instance to be removed
	for _, inst := range slurm.NodeNames(os.Args[1]) {
		log.Printf("INFO: Removing %s instance\n", inst)
		terraform.Del(inst)
		//salt.Del(inst)
	}
	log.Printf("INFO: Pushing Updates\n")

	// update cloud infrastructure
	terraform.Update()
	log.Printf("INFO: Done Updating\n")
}
