package main

import (
	"log"
	"os"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/slurm"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/terraform"
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

var config = "/opt/slurm/latest/etc/cloud_config.json"

func rm(names string) {
	for _, inst := range slurm.NodeNames(names) {
		log.Printf("INFO: Removing %s instance\n", inst)
		terraform.Del(inst)
	}
}

func main() {
	utils.ParseConfig(config)
	//setup logging
	cleanup, err := utils.SetupLogging(utils.Config("log.rm").Self())
	if err != nil {
		log.Fatalf("Critical: %s\n", err)
	}
	defer cleanup()

	//check was given an arg
	err = utils.CheckNumArgs(2, os.Args, "<instance name glob>")
	if err != nil {
		log.Fatalf("CRITICAL: %s\n", err)
	}

	// clean up each instance to be removed
	rm(os.Args[1])

	// update cloud infrastructure
	log.Printf("INFO: Pushing Updates\n")
	terraform.Update()
	log.Printf("INFO: Done Updating\n")
}
