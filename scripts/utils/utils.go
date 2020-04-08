package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
 	"encoding/json"
	"io/ioutil" 
	"flag"
)

//SetupLogging sets up logging to the file /var/lib/slurm/<fn>
//returned func pointer should be defered to do cleanup
func SetupLogging(fn string) (func() error, error) {
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file %s,  %s", fn, err)
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

func parseMap(m map[string]interface{}, prefix string) error{
	var err error
	for k, v := range m {
		switch i := v.(type) {
			case string:
				flag.String(prefix+k, i, prefix+k)
				log.Printf("DEBUG:utils: Adding flag %s: %s\n", prefix+k, i)
			case map[string]interface{}:
				err = parseMap(i, k+"_")
			default:
				fmt.Errorf("Can't parse config option %s: %s", k, i)
		}
	}
	return err
}

//ParseConfig parses the given config file
func ParseConfig(filepath string) error{
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	conf := make(map[string]interface{})
	
	err = json.Unmarshal(f, &conf)
	if err != nil {
		return err
	}

	parseMap(conf, "")

	// parse conf into flags
	flag.Parse()

	return nil
}

func Config(name string) string {
	//TODO check flag.Lookup(name)[.Value?] exists before calling .String() on it 
	return flag.Lookup(name).Value.String()
}
