package devices

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"math"
	"strconv"
	"strings"

	"github.com/fccn/gofetch-snmp/data"
	. "github.com/fccn/gofetch-snmp/log"
	"github.com/fccn/gofetch-snmp/snmp"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type ios struct {
	*device         //Extends Device Struct
	physicalEntries map[string]map[string]interface{}
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *ios) Init() {
	d.device.Init()

	//Unsupported Features
	d.Features.NetworkPolicy = false
	d.Features.CellInfo = false
	d.Features.Ntp = false

	if !d.Cancel && (d.Features.Memory || d.Features.Cpu || d.Features.Sensors) {
		//---------------------------------------OIDs---------------------------------------
		const entPhysicalDescr = ".1.3.6.1.2.1.47.1.1.1.1.2"
		const entPhysicalName = ".1.3.6.1.2.1.47.1.1.1.1.7"
		//-------------------------------------Entries--------------------------------------
		entries := data.Entries{
			{"descr", entPhysicalDescr},
			{"name", entPhysicalName},
		}
		//--------------------------------Result Processing---------------------------------
		d.physicalEntries = map[string]map[string]interface{}{}

		for i := range entries {
			entry := entries[i]
			metric := snmp.WalkAll(d.SnmpConf, d.Bulk, entry.Oid)
			for i := range metric {
				index := snmp.GetIndex(metric[i], entry.Oid)
				if d.physicalEntries[index] == nil {
					d.physicalEntries[index] = map[string]interface{}{}
				}
				switch metric[i].Value.(type) {
				case []uint8:
					d.physicalEntries[index][entry.Name] = string(metric[i].Value.([]uint8))
				case string:
					d.physicalEntries[index][entry.Name] = metric[i].Value.(string)
				}
			}
		}
	}
}

func (d *ios) Uptime() {
	d.device.Uptime()
}

func (d *ios) InterfaceCounters() {
	d.device.InterfaceCounters()
}

func (d *ios) NetworkAcl() {
	//---------------------------------------OIDs---------------------------------------
	const ccarConfigAccIdx = ".1.3.6.1.4.1.9.9.113.1.1.1.1.4"
	const ccarStatHCSwitchedBytes = ".1.3.6.1.4.1.9.9.113.1.2.1.1.11"
	const ccarStatHCFilteredBytes = ".1.3.6.1.4.1.9.9.113.1.2.1.1.13"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"accIndex", ccarConfigAccIdx},
		{"permit_bytes", ccarStatHCSwitchedBytes},
		{"drop_bytes", ccarStatHCFilteredBytes},
	}
	//--------------------------------Result Processing---------------------------------
	//Map To Store The ACL Numbers
	acl := map[string]string{}

	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(pdu.Name, ".")
		accIndex := strings.Join(split[len(split)-3:], ".")

		switch entry.Oid {
		case ccarConfigAccIdx:
			acl[accIndex] = strconv.Itoa(pdu.Value.(int))
		default:
			index := split[len(split)-3]

			dir := map[string]string{"1": "in", "2": "out"}[split[len(split)-2]]

			m.AddField(index, "interface_"+dir+"_acl_"+acl[accIndex]+"_"+entry.Name, pdu.Value.(uint64))
		}
	})
	d.AddDataFromEntries(INTERFACE, entries, function)
}

func (d *ios) BgpPeers() {
	//---------------------------------------OIDs---------------------------------------
	const cbgpPeerAcceptedPrefixes = ".1.3.6.1.4.1.9.9.187.1.2.4.1.1"
	const cbgpPeerDeniedPrefixes = ".1.3.6.1.4.1.9.9.187.1.2.4.1.2"
	const cbgpPeerPrefixAdminLimit = ".1.3.6.1.4.1.9.9.187.1.2.4.1.3"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"bgp_accepted_prefixes", cbgpPeerAcceptedPrefixes},
		{"bgp_denied_prefixes", cbgpPeerDeniedPrefixes},
		{"bgp_limit_prefixes", cbgpPeerPrefixAdminLimit},
	}
	//--------------------------------Result Processing---------------------------------
	function := func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(strings.TrimPrefix(pdu.Name, entry.Oid+"."), ".")
		index := strings.Join(split[:4], ".")

		m.AddField(index, entry.Name, pdu.Value)
		m.AddTag(index, "bgp_neighbour", index)
	}
	d.AddDataFromEntries(BGP, entries, function)
}

func (d *ios) Memory() {
	//Check If Physical data.Entries Table Was Obtained
	if d.physicalEntries == nil || len(d.physicalEntries) == 0 {
		Log("Physical data.Entries Was Not Found - Memory Feature Was Cancelled")
		return
	}

	//---------------------------------------OIDs---------------------------------------
	const cempMemPoolName = ".1.3.6.1.4.1.9.9.221.1.1.1.1.3"
	const cempMemPoolUsed = ".1.3.6.1.4.1.9.9.221.1.1.1.1.7"
	const cempMemPoolFree = ".1.3.6.1.4.1.9.9.221.1.1.1.1.8"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"name", cempMemPoolName},
		{"used_bytes", cempMemPoolUsed},
		{"free_bytes", cempMemPoolFree},
	}
	//--------------------------------Result Processing---------------------------------
	//Stores The Memory Pool Names
	names := map[string]string{}

	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		//Index Is The Number Before The Last, Last Indicates The Memory Pool
		split := strings.Split(pdu.Name, ".")
		index := split[len(split)-2]
		poolIndex := index + "." + split[len(split)-1]

		switch entry.Oid {
		//Get Each Of The Pool Names
		case cempMemPoolName:
			//Get Pool Name
			names[poolIndex] = strings.ToLower(string(pdu.Value.([]uint8)))

			//Add Respective Description And Name
			for tag := range d.physicalEntries[index] {
				m.AddTag(index, "memory_"+tag, d.physicalEntries[index][tag].(string))
			}
		//Get The Memory Used/Free
		default:
			m.AddField(index, "memory_"+names[poolIndex]+"_"+entry.Name, pdu.Value.(uint))
		}
	})
	d.AddDataFromEntries(MEMORY, entries, function)
}

func (d *ios) Cpu() {
	//Check If Physical data.Entries Table Was Obtained
	if d.physicalEntries == nil || len(d.physicalEntries) == 0 {
		Log("Physical data.Entries Was Not Found - CPU Feature Was Cancelled")
		return
	}

	//---------------------------------------OIDs---------------------------------------
	const cpmCPUTotal1minRev = ".1.3.6.1.4.1.9.9.109.1.1.1.1.7"
	const cpmCPUTotal5minRev = ".1.3.6.1.4.1.9.9.109.1.1.1.1.8"
	const cpmCPUTotalPhysicalIndex = ".1.3.6.1.4.1.9.9.109.1.1.1.1.2"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"index", cpmCPUTotalPhysicalIndex},
		{"cpu_one_minute_percent", cpmCPUTotal1minRev},
		{"cpu_five_minutes_percent", cpmCPUTotal5minRev},
	}
	//--------------------------------Result Processing---------------------------------
	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		index := snmp.GetIndex(pdu, entry.Oid)

		switch entry.Oid {
		case cpmCPUTotalPhysicalIndex:
			phyIndex := strconv.Itoa(pdu.Value.(int))

			//Add Respective Description And Name
			for tag := range d.physicalEntries[phyIndex] {
				m.AddTag(index, "cpu_"+tag, d.physicalEntries[phyIndex][tag].(string))
			}
		default:
			m.AddField(index, entry.Name, pdu.Value)
		}
	})
	d.AddDataFromEntries(CPU, entries, function)
}

func (d *ios) Sensors() {
	//---------------------------------------OIDs---------------------------------------
	const ciscoEnvMonVoltageStatusDescr = ".1.3.6.1.4.1.9.9.13.1.2.1.2"
	const ciscoEnvMonVoltageStatusValue = ".1.3.6.1.4.1.9.9.13.1.2.1.3"
	const ciscoEnvMonVoltageThresholdLow = ".1.3.6.1.4.1.9.9.13.1.2.1.4"
	const ciscoEnvMonVoltageThresholdHigh = ".1.3.6.1.4.1.9.9.13.1.2.1.5"
	const ciscoEnvMonVoltageThresholdState = ".1.3.6.1.4.1.9.9.13.1.2.1.7"

	const ciscoEnvMonTemperatureStatusDescr = ".1.3.6.1.4.1.9.9.13.1.3.1.2"
	const ciscoEnvMonTemperatureStatusValue = ".1.3.6.1.4.1.9.9.13.1.3.1.3"
	const ciscoEnvMonTemperatureThreshold = ".1.3.6.1.4.1.9.9.13.1.3.1.4"
	const ciscoEnvMonTemperatureState = ".1.3.6.1.4.1.9.9.13.1.3.1.6"

	const ciscoEnvMonFanStatusDescr = ".1.3.6.1.4.1.9.9.13.1.4.1.2"
	const ciscoEnvMonFanState = ".1.3.6.1.4.1.9.9.13.1.4.1.3"

	const ciscoEnvMonSupplyStatusDescr = ".1.3.6.1.4.1.9.9.13.1.5.1.2"
	const ciscoEnvMonSupplyState = ".1.3.6.1.4.1.9.9.13.1.5.1.3"
	//-------------------------------------Entries--------------------------------------
	voltage := data.Entries{
		{"sensor_descr", ciscoEnvMonVoltageStatusDescr},
		{"sensor_value", ciscoEnvMonVoltageStatusValue},
		{"sensor_thresh_low", ciscoEnvMonVoltageThresholdLow},
		{"sensor_thresh_high", ciscoEnvMonVoltageThresholdHigh},
		{"sensor_state", ciscoEnvMonVoltageThresholdState},
	}

	temperature := data.Entries{
		{"sensor_descr", ciscoEnvMonTemperatureStatusDescr},
		{"sensor_value_celsius", ciscoEnvMonTemperatureStatusValue},
		{"sensor_thresh_celsius", ciscoEnvMonTemperatureThreshold},
		{"sensor_state", ciscoEnvMonTemperatureState},
	}

	fan := data.Entries{
		{"sensor_descr", ciscoEnvMonFanStatusDescr},
		{"sensor_state", ciscoEnvMonFanState},
	}

	supply := data.Entries{
		{"sensor_descr", ciscoEnvMonSupplyStatusDescr},
		{"sensor_state", ciscoEnvMonSupplyState},
	}
	//--------------------------------Result Processing---------------------------------

	//Auxiliary Map That Marks Which Sensors Are Measured In Millis
	millisMap := map[string]bool{}

	//Auxiliary Map That Marks Which Sensors Are Measured In Millis
	unitMap := map[string]string{}

	//Anonymous Function That Deals With Voltage Values
	voltageSensorData := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(pdu.Name, ".")

		//Index Will Have A Prefix That Indicates The Sensor Table, So That We Can Distinguish Them
		index := split[len(split)-4] + "." + split[len(split)-1]

		value := pdu.Value

		switch entry.Oid {
		//Description Goes Into The Tags
		case ciscoEnvMonVoltageStatusDescr:
			//Description Is A []uint8, Convert To String
			descr := string(value.([]uint8))

			//Look For "(in mV)" To See If Is Millis And To Convert To Unit
			inmv := strings.Index(descr, "(in mV)")

			//If There's "(in mV)", Remove It From Descr
			if inmv != -1 {
				descr = descr[:inmv]
			}

			//Can Be Volts Or Amperes, Depending On The Presence Of "amps"
			if strings.Index(descr, "amps") != -1 {
				unitMap[index] = "amperes"
			} else {
				unitMap[index] = "volts"
			}

			//Value On This Index May Be In Millis, Depending On The Presence Of "(in mV)"
			millisMap[index] = inmv != -1

			//Add The Description
			m.AddTag(index, entry.Name, descr)

		//Everything Else Goes Into The Fields
		default:
			name := entry.Name
			if entry.Oid != ciscoEnvMonVoltageThresholdState {
				//Convert From Millis
				if millisMap[index] {
					value = float64(value.(int)) / 1000
				} else {
					value = float64(value.(int))
				}
				//Add The Unit To The Name
				name += "_" + unitMap[index]
			}
			m.AddField(index, name, value)
		}
	})

	//Anonymous Function That Deals With All Other Sensor Values
	sensorData := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(pdu.Name, ".")

		//Index Will Have A Prefix That Indicates The Sensor Table, So That We Can Distinguish Them
		index := split[len(split)-4] + "." + split[len(split)-1]

		switch entry.Name {
		//Description Goes Into The Tags
		case "sensor_descr":
			//Description Is A []uint8, Convert To String
			descr := string(pdu.Value.([]uint8))

			//Add The Description
			m.AddTag(index, entry.Name, descr)

		//Everything Else Goes Into The Fields
		default:
			if entry.Name == "sensor_state" {
				m.AddField(index, entry.Name, pdu.Value)
			} else {
				switch pdu.Value.(type) {
				case uint:
					m.AddField(index, entry.Name, float64(pdu.Value.(uint)))
				case int:
					m.AddField(index, entry.Name, float64(pdu.Value.(int)))
				}
			}
		}
	})

	//Add All The Data
	d.AddDataFromEntries(SENSOR, voltage, voltageSensorData)
	d.AddDataFromEntries(SENSOR, temperature, sensorData)
	d.AddDataFromEntries(SENSOR, fan, sensorData)
	d.AddDataFromEntries(SENSOR, supply, sensorData)

	//Map - TypeNumber: TypeName
	types := map[int]string{
		1:  "", //Other
		2:  "", //Unknown
		3:  "volts",
		4:  "volts",
		5:  "amperes",
		6:  "watts",
		7:  "hertz",
		8:  "celsius",
		9:  "percent_rh",
		10: "rpm",
		11: "cmm",
		12: "bool",
		13: "special_enum",
		14: "dbm",
		15: "db",
	}
	//---------------------------------------OIDs---------------------------------------
	const entSensorType = ".1.3.6.1.4.1.9.9.91.1.1.1.1.1"
	const entSensorScale = ".1.3.6.1.4.1.9.9.91.1.1.1.1.2"
	const entSensorPrecision = ".1.3.6.1.4.1.9.9.91.1.1.1.1.3"
	const entSensorValue = ".1.3.6.1.4.1.9.9.91.1.1.1.1.4"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"type", entSensorType},
		{"scale", entSensorScale},
		{"precision", entSensorPrecision},
		{"sensor_value", entSensorValue},
	}
	//--------------------------------Result Processing---------------------------------
	sensorType := map[string]string{}
	sensorScale := map[string]int{}
	sensorPrecision := map[string]int{}

	sensorData = data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		index := snmp.GetIndex(pdu, entry.Oid)

		switch entry.Oid {
		case entSensorType:
			sensorType[index] = types[pdu.Value.(int)]
		case entSensorScale:
			sensorScale[index] = pdu.Value.(int)
		case entSensorPrecision:
			sensorPrecision[index] = pdu.Value.(int)
		case entSensorValue:
			var value interface{}
			if sensorType[index] == "bool" {
				value = pdu.Value.(int) == 1
			} else {
				value = float32(float64(pdu.Value.(int)) * math.Pow10((sensorScale[index]-9)*3-sensorPrecision[index]))
			}
			m.AddField(index, entry.Name+"_"+sensorType[index], value)
			m.AddTag(index, "sensor_descr", d.physicalEntries[index]["name"].(string)+" - "+d.physicalEntries[index]["descr"].(string))
		}
	})
	d.AddDataFromEntries(SENSOR, entries, sensorData)
}

func (d *ios) Fetch(dat *data.Data, s *runner.S) {
	d.device.Fetch(dat, s)
}
