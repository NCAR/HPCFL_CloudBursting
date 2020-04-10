package slurm

import (
	"log"
	"os/exec"
	"strings"

	"github.com/NCAR/HPCFL_TerraformScripts/scripts/utils"
)

//NodeNames takes a glob string and returns a slice of all of the names contained in the glob
//The glob string is of the format slurm uses
func NodeNames(glob string) []string {
	//TODO handle weird glob strings
	log.Printf("DEBUG:slurm: expanding glob %s\n", glob)
	//use slurm to expand glob
	out, err := exec.Command(utils.Config("slurm.dir")+"/bin/scontrol", "show", "hostname", glob).Output()
	if err != nil {
		log.Printf("ERROR:slurm: Error expanding nodename glob %s\n", err)
	}
	names := strings.Split(string(out), "\n")
	//last element is empty/just whitespace so drop it
	return names[:len(names)-1]
}
