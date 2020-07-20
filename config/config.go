package config

import (
	. "github.com/fccn/gofetch/log"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Config struct{
	Version 	string
	Interval 	time.Duration
	Timeout 	time.Duration
	MaxRoutines int64
}

type config struct{
	Version 	string 		`yaml:"version"`
	Interval 	interface{} `yaml:"interval"`
	Timeout 	interface{}	`yaml:"timeout"`
	MaxRoutines int64 		`yaml:"maxroutines"`
}

func getDuration(i interface{})(time.Duration, error){
	switch i.(type) {
	case int:
		if i.(int) > 0{
			return time.Duration(i.(int)) * time.Minute, nil
		}
	case string:
		return time.ParseDuration(i.(string))
	}
	return -1, fmt.Errorf("Error: %v Is Not A Valid Time Value", i)
}

func GetConfigs(configFile string)(c *Config){
	//Initialize Struct With Default Values
	c = &Config{}

	//Use Auxiliary Struct To Receive Unprocessed Values
	aux := config{}

	//Decode The Configurations File To The Config Struct
	if conf, err := ioutil.ReadFile(configFile); err == nil{
		if err := yaml.Unmarshal(conf, &aux); err != nil {
			FatalLog(fmt.Sprintf("Could Not Decode Configuration File: %v", err))
		}
		if t, err := getDuration(aux.Interval); err == nil{
			c.Interval = t
		}else{
			Log(err.Error())
		}
		if t, err := getDuration(aux.Timeout); err == nil{
			c.Timeout = t
		}else{
			Log(err.Error())
		}
		c.MaxRoutines = aux.MaxRoutines
		c.Version     = aux.Version
	} else{
		FatalLog(fmt.Sprintf("Could Not Decode Configuration File: %v", err))
	}
	return
}