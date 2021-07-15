package main

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/fccn/gofetch-snmp/config"
	"github.com/fccn/gofetch-snmp/data"
	"github.com/fccn/gofetch-snmp/devices"
	. "github.com/fccn/gofetch-snmp/log"
	"github.com/matryer/runner"
	"golang.org/x/sync/semaphore"
	"gopkg.in/yaml.v2"
)

//------------------------------------------------------------------------------------------
//----------------------------------------VARIABLES-----------------------------------------
//------------------------------------------------------------------------------------------
var wg, forever sync.WaitGroup
var ss *semaphore.Weighted
var ctx context.Context
var tasks []*runner.Task
var fetchedData []*data.Data

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c) //Waits For Func() Return To Close Chan "c"
		wg.Wait()
	}()
	select {
	case <-c:
		return false //OK
	case <-time.After(timeout):
		return true //Timed Out
	}
}

func stopAllTasks() {
	for _, task := range tasks {
		task.Stop()
	}
}

func fetchData(host devices.Host) {
	//Fetch Data From Device
	if dev := devices.NewDevice(host); dev != nil {
		dat := data.NewData()
		fetchedData = append(fetchedData, &dat)

		//Multithreading Sync
		wg.Add(1)
		ss.Acquire(ctx, 1)

		//Run Fetch On GoRoutine And Store Task To Stop On Timeout
		tasks = append(tasks, runner.Go(func(s runner.S) error {
			//Multithreading Sync
			defer wg.Done()
			defer ss.Release(1)

			dev.Fetch(&dat, &s)

			return nil
		}))
	}
}

func writeData() {
	//Test Connection To The InfluxDB
	influx := data.InfluxTestConnection()

	//If Connection To InfluxDB Is Possible
	if influx {
		for i := 0; i < len(fetchedData); {
			//Try To Write To InfluxDB And Remove Position If Writing Is Successful
			if success := fetchedData[i].WriteInflux(); success {
				fetchedData = fetchedData[i+1:]
			}
		}
	}
	//If Connection To InfluxDB Is Not Possible, Or Writing Was Not Successful
	if !influx || len(fetchedData) > 0 {
		//Try To Write Locally And Empty Array If Writing Is Successful
		if success := data.LocalWrite(fetchedData); success {
			DebugLog("Successfully Stored Fetched Data Locally")
			fetchedData = []*data.Data{}
		}
	}
}

func main() {
	//Get The Flags From The Execution Command
	var confFile, hostsConfFile, dbConfFile string
	flag.StringVar(&confFile, "c", confFile, "General - Configuration File")
	flag.StringVar(&hostsConfFile, "h", hostsConfFile, "Hosts - Configuration File")
	flag.StringVar(&dbConfFile, "d", dbConfFile, "Database - Configuration File")
	flag.Parse()

	//Get General Configurations Struct
	conf := config.GetConfigs(confFile)

	//Set Debug Flag On Util Module
	Debug(conf.Debug)

	//Configuring Log Output
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//Alive Acknowledgement
	Log(fmt.Sprintf("GoFetch v%s running", version))

	//Get Hosts' Configurations
	var hosts devices.Hosts
	h, err := ioutil.ReadFile(hostsConfFile)
	if err == nil {
		err = yaml.Unmarshal(h, &hosts)
	}
	if err != nil {
		FatalLog(fmt.Sprintf("Could Not Decode Hosts Configuration File: %v", err))
	}

	//Initialize The InfluxDB Connection
	data.InfluxInit(dbConfFile)

	//Set A Ticker That Defines The Running Interval
	ticker := time.NewTicker(conf.Interval)

	//To Limit Number Of Routines Running
	ss = semaphore.NewWeighted(conf.MaxRoutines)
	ctx = context.TODO()

	//Starting The Infinite Loop
	forever.Add(1)
	go func() {
		for firstRun := true; ; firstRun = false {

			//Wait 1 Minute, If It's Not The First Iteration
			if !firstRun {
				<-ticker.C
			}

			//Collection Control Information
			DebugLog("Collection Started")

			//Retrieve Data For All Hosts
			for _, host := range hosts.Hosts {
				fetchData(host)
			}

			//Timeout The Thread After Given Time In Seconds
			if waitTimeout(&wg, conf.Timeout) {
				stopAllTasks()
			}

			//Write Fetched Data To InfluxDB or Disk
			writeData()

			//Collection Control Information
			DebugLog("Collection Ended")
		}
	}()
	forever.Wait()
}
