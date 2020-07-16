package devices

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"fccn.pt/glopes/gofetch/data"
	"fccn.pt/glopes/gofetch/snmp"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type opengear struct{
	*device //Extends Device Struct
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *opengear) Init() {
	d.device.Init()

	//Unsupported Features
	d.Features.NetworkAcl    = false
	d.Features.NetworkPolicy = false
	d.Features.BgpPeers 	 = false
}

func (d *opengear) Uptime(){
	d.device.Uptime()
}

func (d *opengear) InterfaceCounters(){
	d.device.InterfaceCounters()
}

func (d *opengear) CellInfo(){
	//---------------------------------------OIDs---------------------------------------
	const ogCellModemEnabled 		 = ".1.3.6.1.4.1.25049.17.17.1.4.1"
	const ogCellModemConnected 		 = ".1.3.6.1.4.1.25049.17.17.1.5.1"
	const ogCellModemRegistered 	 = ".1.3.6.1.4.1.25049.17.17.1.7.1"
	const ogCellModemTower 			 = ".1.3.6.1.4.1.25049.17.17.1.8.1"
	const ogCellModemRadioTechnology = ".1.3.6.1.4.1.25049.17.17.1.9.1"
	const ogCellModem3gRssi 		 = ".1.3.6.1.4.1.25049.17.17.1.11.1"
	const ogCellModem4gRssi 		 = ".1.3.6.1.4.1.25049.17.17.1.12.1"
	const ogCellModemSessionTime 	 = ".1.3.6.1.4.1.25049.17.17.1.13.1"
	const ogCellModemSelectedSimCard = ".1.3.6.1.4.1.25049.17.17.1.14.1"
	const ogCellModemTemperature 	 = ".1.3.6.1.4.1.25049.17.17.1.15.1"
	const ogCellModemCounter 		 = ".1.3.6.1.4.1.25049.17.17.1.16.1"

	entries := data.Entries{
		{"cell_modem_enabled", 		ogCellModemEnabled},
		{"cell_modem_connected", 		ogCellModemConnected},
		{"cell_modem_registered", 	ogCellModemRegistered},
		{"cell_modem_tower", 			ogCellModemTower},
		{"cell_modem_tech", 			ogCellModemRadioTechnology},
		{"cell_modem_3g_rssi", 		ogCellModem3gRssi},
		{"cell_modem_4g_rssi", 		ogCellModem4gRssi},
		{"cell_modem_session_time", 	ogCellModemSessionTime},
		{"cell_modem_sim_card", 		ogCellModemSelectedSimCard},
		{"cell_modem_temperature", 	ogCellModemTemperature},
		{"cell_modem_counter", 		ogCellModemCounter},
	}

	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(CELL, entries)
}

func (d *opengear) Memory(){
	//---------------------------------------OIDs---------------------------------------
	const memTotalReal = ".1.3.6.1.4.1.2021.4.5"
	const memTotalFree = ".1.3.6.1.4.1.2021.4.11"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"memory_total", memTotalReal},
		{"memory_free",  memTotalFree},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(MEMORY, entries)
}

func (d *opengear) Cpu(){
	//---------------------------------------OIDs---------------------------------------
	const ssCpuRawUser   = ".1.3.6.1.4.1.2021.11.50"
	const ssCpuRawSystem = ".1.3.6.1.4.1.2021.11.52"
	const ssCpuRawIdle   = ".1.3.6.1.4.1.2021.11.53"
	const ssCpuRawWait   = ".1.3.6.1.4.1.2021.11.54"
	const ssCpuRawKernel = ".1.3.6.1.4.1.2021.11.55"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"cpu_user",   ssCpuRawUser},
		{"cpu_system", ssCpuRawSystem},
		{"cpu_idle",   ssCpuRawIdle},
		{"cpu_wait",   ssCpuRawWait},
		{"cpu_kernel", ssCpuRawKernel},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(CPU, entries)
}

func (d *opengear) Sensors(){
	if !d.Features.Sensors{return}
	//---------------------------------------OIDs---------------------------------------
	const ogEmdTemperatureName 		  = ".1.3.6.1.4.1.25049.17.9.1.3"
	const ogEmdTemperatureDescription = ".1.3.6.1.4.1.25049.17.9.1.4"
	const ogEmdTemperatureValue		  = ".1.3.6.1.4.1.25049.17.9.1.5"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"sensor_name",  			ogEmdTemperatureName},
		{"sensor_descr", 			ogEmdTemperatureDescription},
		{"sensor_value_celsius", 	ogEmdTemperatureValue},
	}
	//--------------------------------Result Processing---------------------------------
	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU){
		index := snmp.GetIndex(pdu, entry.Oid)
		switch entry.Oid{
		case ogEmdTemperatureValue:
			m.AddField(index, entry.Name, pdu.Value)
		default:
			m.AddTag(index, entry.Name, pdu.Value.(string))
		}
	})
	d.AddDataFromEntries(SENSOR, entries, function)
}

func (d *opengear) Fetch(dat *data.Data, s *runner.S){
	d.device.Fetch(dat, s)
}