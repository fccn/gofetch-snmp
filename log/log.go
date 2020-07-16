package log

import (
	"fmt"
	"os"
	"time"
)

//------------------------------------------------------------------------------------------
//----------------------------------------VARIABLES-----------------------------------------
//------------------------------------------------------------------------------------------
var debug bool

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
//Sets The Debug Flag
func Debug(d bool){
	debug = d
}

//Prints A Debug Message If The Flag Is Active
func DebugLog(str string){
	if debug {
		log(str, "")
	}
}

func Log(str string){
	log(str, now() + ": ")
}

func FatalLog(str string){
	log(str, now() + " - FATAL: ")
	os.Exit(1)
}

func log(str string, prefix string){
	fmt.Printf("%s%s\n", prefix, str)
}

func now()string{
	t := time.Now()
	return fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}