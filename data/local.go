package data

import (
	"encoding/json"
	. "fccn.pt/glopes/gofetch/log"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func LocalWrite(d []*Data)bool{
	//Marshal The Data And Check For Errors
	if data, err := json.MarshalIndent(d, "", " "); err == nil {
		//Get The Current Time In Nanoseconds For The File's Name
		fileName := fmt.Sprint(time.Now().UnixNano()) + ".json"

		//Create A Write-Only File With The Specified name
		if f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0755); err == nil {
			//Close On Return
			defer f.Close()

			//Write The Content To The File
			if err := ioutil.WriteFile(fileName, data, 0755); err != nil {
				Log("Could not write to file: " + err.Error())
			}

			return true
		}
	}
	return false
}