package slurm

import (
	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
	"log"
	"os/exec"
	"strings"
)

const SLURM_DIR = "/opt/slurm/latest/"

func Add(inst utils.Instance) {
	log.Printf("INFO:slurm: adding %s\n", inst)
}

func Del(inst string) {
	log.Printf("INFO:slurm: removing %s\n", inst)
}

//NodeNames takes a glob string and returns a slice of all of the names contained in the glob
//The glob string is of the format slurm uses
func NodeNames(glob string) []string {
	//TODO handle weird glob strings
	log.Printf("DEBUG:slurm: expanding glob %s\n", glob)
	//use slurm to expand glob
	out, err := exec.Command(SLURM_DIR+"/bin/scontrol", "show", "hostname", glob).Output()
	if err != nil {
		log.Printf("ERROR:slurm: Error expanding nodename glob %s\n", err)
	}
	names := strings.Split(string(out), "\n")
	//last element is empty/just whitespace so drop it
	return names[:len(names)-1]
}
