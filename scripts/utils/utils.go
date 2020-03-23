package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
)

//Instance represents a cloud instance
type Instance interface {
	Name() string
	IP() string
	String() string
	Setup() error
}

//SetupLogging sets up logging to the file /var/lib/slurm/<fn>
func SetupLogging(fn string) (func() error, error) {
	f, err := os.OpenFile("/var/lib/slurm/"+fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file /var/lib/slurm/%s, %s", fn, err)
	}
	log.SetOutput(f)

	return f.Close, nil
}

//CheckNumArgs checks the number of arguments given is correct
//Creates a usefull error message if not
func CheckNumArgs(num int, args []string, usage string) error {
	if len(os.Args) != num {
		var errString string
		if len(os.Args) >= 1 {
			errString = fmt.Sprintf("Incorrect usage should do: %s %s", args[0], usage)
		} else {
			errString = fmt.Sprintf("Incorrect usage should do: scriptname %s", usage)
		}
		return errors.New(errString)
	}
	return nil
}
