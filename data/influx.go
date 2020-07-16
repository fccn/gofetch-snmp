package data

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	. "fccn.pt/glopes/gofetch/log"
	"fmt"
	"github.com/influxdata/influxdb1-client/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type influx struct {
	Server   string `yaml:"server"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	Ping     int 	`yaml:"ping"`
}

//------------------------------------------------------------------------------------------
//----------------------------------------VARIABLES-----------------------------------------
//------------------------------------------------------------------------------------------
var db influx
var c client.Client

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------

//Creates The InfluxDB Connection And Checks Server
func InfluxInit(dbConfigFile string) {
	//Decode The Configurations File To The DB Struct
	if conf, err := ioutil.ReadFile(dbConfigFile); err == nil{
		if err := yaml.Unmarshal(conf, &db); err != nil {
			FatalLog(fmt.Sprintf("Could Not Decode InfluxDB Configuration File: %v", err))
		}
		//Use The Configurations From The File To Initialize The DB Connection
		if c, err = client.NewHTTPClient(
			client.HTTPConfig{
				Addr:     db.Server,
				Username: db.Username,
				Password: db.Password,
			}); err != nil{
			FatalLog(fmt.Sprintf("Could Not Initialize InfluxDB Client: %v", err))
		}
	} else{
		FatalLog(fmt.Sprintf("Could Not Decode InfluxDB Configuration File: %v", err))
	}
}

func InfluxTestConnection()bool{
	var err error
	if _, _, err = c.Ping(time.Duration(db.Ping) * time.Second); err != nil {
		Log(fmt.Sprintf("Could Not Estabilish InfluxDB Connection: %s", err.Error()))
	}
	return err == nil
}

//Creates A Batch To An InfluxDB
func InfluxCreateBatch() (bp client.BatchPoints) {
	var err error
	if bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Database:  db.Database,
		Precision: "s",
	}); err != nil {
		Log(fmt.Sprintf("Could Not Create A BatchPoints Instance: %s", err.Error()))
	}
	return
}

//Creates A Point And Adds It To A Batch
func InfluxAddPoint(bp client.BatchPoints, name string, tags *map[string]string, fields *map[string]interface{}, timestamp time.Time) {
	pt, err := client.NewPoint(name, *tags, *fields, timestamp)
	if err != nil {
		DebugLog(fmt.Sprintf("Could Not Add Point %s: %s", name, err.Error()))
		return
	}
	bp.AddPoint(pt)
}

//Writes The Batch Point And Closes The Connection
func InfluxWriteBatch(bp client.BatchPoints)bool{
	var err error
	if err = c.Write(bp); err != nil {
		Log(fmt.Sprintf("Could Not Write BatchPoints: %s", err.Error()))
		return false
	}
	if err = c.Close(); err != nil {
		Log(fmt.Sprintf("Could Not Close InfluxDB Connection: %s", err.Error()))
		return false
	}
	DebugLog("Batch Was Written To DB")
	return true
}