package util

import (
	"time"
)

func FunctionDuration(f func())float64{
	//Get Starting Time
	start := time.Now()

	//Call Function
	f()

	//Return Time Difference In Seconds
	return time.Now().Sub(start).Seconds()
}
