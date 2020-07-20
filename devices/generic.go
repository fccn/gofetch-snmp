package devices

import (
	"github.com/fccn/gofetch/data"
	"github.com/matryer/runner"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type generic struct{
	*device //Extends Device Struct
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *generic) Init(){
	d.device.Init()

	//Unsupported Features
	d.Features.CellInfo 	 = false
	d.Features.NetworkAcl 	 = false
	d.Features.NetworkPolicy = false
	d.Features.BgpPeers 	 = false
	d.Features.Memory   	 = false
	d.Features.Cpu      	 = false
	d.Features.Sensors  	 = false
}

func (d *generic) Uptime(){
	d.device.Uptime()
}

func (d *generic) InterfaceCounters(){
	d.device.InterfaceCounters()
}

func (d *generic) Fetch(dat *data.Data, s *runner.S){
	d.device.Fetch(dat, s)
}
