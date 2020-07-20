package devices

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"github.com/fccn/gofetch/data"
	"github.com/fccn/gofetch/snmp"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type mrv struct {
	*device //Extends Device Struct
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *mrv) Init() {
	d.device.Init()

	//Unsupported Features
	d.Features.NetworkAcl    = false
	d.Features.NetworkPolicy = false
	d.Features.BgpPeers 	 = false
	d.Features.Memory   	 = false
	d.Features.Cpu      	 = false
}

func (d *mrv) Uptime(){
	d.device.Uptime()
}

func (d *mrv) InterfaceCounters(){
	d.device.InterfaceCounters()
}

func (d *mrv) CellInfo(){
	//---------------------------------------OIDs---------------------------------------
	const irGsmPortRcvSigStrength = ".1.3.6.1.4.1.33.100.2.13.1.2"
	const irGsmPortBitErrorRate   = ".1.3.6.1.4.1.33.100.2.13.1.3"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"cell_signal_strength", irGsmPortRcvSigStrength},
		{"cell_bit_error_rate",  irGsmPortBitErrorRate},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(CELL, entries)
}

func (d *mrv) Sensors(){
	//---------------------------------------OIDs---------------------------------------
	const irSysCurrentTemp 		 = ".1.3.6.1.4.1.33.100.1.1.14"
	const irSysTempThresholdLow  = ".1.3.6.1.4.1.33.100.1.1.15"
	const irSysTempThresholdHigh = ".1.3.6.1.4.1.33.100.1.1.16"

	const irPowerInputStatus     = ".1.3.6.1.4.1.33.100.1.6.1.1.3"
	const irPowerOutputStatus    = ".1.3.6.1.4.1.33.100.1.6.1.1.4"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"sensor_value_celsius",       irSysCurrentTemp},
		{"sensor_thresh_low_celsius",  irSysTempThresholdLow},
		{"sensor_thresh_high_celsius", irSysTempThresholdHigh},
		{"sensor_input_status_bool",  irPowerInputStatus},
		{"sensor_output_status_bool", irPowerOutputStatus},
	}
	//--------------------------------Result Processing---------------------------------
	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		index := snmp.GetIndex(pdu, entry.Oid)
		m.AddTag(index, "sensor_index", index)
		switch entry.Oid{
		case irPowerInputStatus, irPowerOutputStatus:
			m.AddField(index, entry.Name, pdu.Value.(int) == 1)
		default:
			m.AddField(index, entry.Name, pdu.Value)
		}
	})
	d.AddDataFromEntries(SENSOR, entries, function)
}

func (d *mrv) Fetch(dat *data.Data, s *runner.S){
	d.device.Fetch(dat, s)
}