package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var config list

//Elem represents an element in config
//It can either be a map containg more elements that can be looked up
//or a string
type Elem interface {
	//value of self
	Contains() map[string]Elem
	//value of a child
	Lookup(string) Elem
	//Elem's value
	Self() string
}

type item struct {
	value string
}

func (i item) Contains() map[string]Elem {
	return nil
}

func (i item) Lookup(s string) Elem {
	if s == "" {
		return i
	}
	return nil
}

func (i item) Self() string {
	return i.value
}

type list struct {
	self  string
	value map[string]Elem
}

func (i list) Contains() map[string]Elem {
	return i.value
}

func (i list) Lookup(s string) Elem {
	m := strings.SplitN(s, ".", 2)
	if len(m) < 2 {
		return i
	}
	if _, ok := i.value[m[0]]; !ok {
		log.Printf("DEBUG:utils: key %s not found\n", s)
		return nil
	}
	return i.value[m[0]].Lookup(m[1])
}

func (i list) Self() string {
	return i.self
}

func (i list) add(k string, v Elem) {
	i.value[k] = v
}

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

func parseMap(m map[string]interface{}, p *list) {
	for k, v := range m {
		switch i := v.(type) {
		case string:
			log.Printf("DEBUG:utils: Adding flag %s: %s\n", k, i)
			n := item{value: i}
			p.add(k, n)
		case map[string]interface{}:
			log.Printf("Debug:utils: Adding map %s: {%v}\n", k, i)
			n := list{self: k, value: make(map[string]Elem)}
			parseMap(i, &n)
			p.add(k, n)
		default:
			log.Fatalf("Can't parse config option %s: %s", k, i)
		}
	}
}

//ParseConfig parses the given config file
func ParseConfig(filepath string) error {
	f, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	conf := make(map[string]interface{})

	err = json.Unmarshal(f, &conf)
	if err != nil {
		return err
	}

	conList := make(map[string]Elem)
	config = list{self: "", value: conList}

	log.Printf("about to parse config\n")
	parseMap(conf, &config)
	log.Printf("parsed config\n")

	// parse conf into flags
	for _, option := range []string{"log.add", "log.rm", "terraform.dir", "terraform.tf_files", "aws.ami", "aws.size", "aws.dir", "aws.name", "slurm.dir"} {
		Config(option)
	}
	return nil
}

//Config returns the value of the given config option
//If the config option was not set the Config panics
func Config(name string) Elem {
	return config.Lookup(name)
}
