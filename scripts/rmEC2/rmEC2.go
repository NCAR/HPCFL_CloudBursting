package main

import (
	"log"
	"os"
	"github.com/Will-Shanks/terraformScripts/terraform"
//    "github.com/Will-Shanks/terraformScripts/salt"
    "github.com/Will-Shanks/terraformScripts/slurm"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <instance name glob>\n", os.Args[0])
	}
	os.Setenv("PWD", "/home/slurm/terraform")
	for _, inst := range slurm.NodeNames(os.Args[1]){
		log.Printf("INFO: Removing %s instance\n", inst)
		terraform.Del(inst)
	//	slurm.Del(inst)
	//	salt.Del(inst)
	}
	log.Printf("INFO: Pushing Updates\n")
	terraform.Update()
}
