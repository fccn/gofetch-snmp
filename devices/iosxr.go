package devices

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"fccn.pt/glopes/gofetch/data"
	. "fccn.pt/glopes/gofetch/log"
	"fccn.pt/glopes/gofetch/snmp"
	"fccn.pt/glopes/gofetch/util"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type iosxr struct {
	*device         //Extends Device Struct
	physicalEntries map[string]map[string]interface{}
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *iosxr) Init() {
	d.device.Init()

	//Unsupported Features
	d.Features.CellInfo = false

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

func (d *iosxr) Uptime() {
	d.device.Uptime()
}

func (d *iosxr) InterfaceCounters() {
	d.device.InterfaceCounters()
	d.ipv6()
}

func (d *iosxr) ipv6() {
	//---------------------------------------OIDs---------------------------------------
	const ipIfStatsHCInOctets = ".1.3.6.1.2.1.4.31.3.1.6"
	const ipIfStatsHCOutOctets = ".1.3.6.1.2.1.4.31.3.1.33"
	const ipIfStatsHCInMcastOctets = ".1.3.6.1.2.1.4.31.3.1.37"
	const ipIfStatsHCOutMcastOctets = ".1.3.6.1.2.1.4.31.3.1.41"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"interface_in_ipv6_uni_bytes", ipIfStatsHCInOctets},
		{"interface_in_ipv6_multi_bytes", ipIfStatsHCInMcastOctets},
		{"interface_out_ipv6_uni_bytes", ipIfStatsHCOutOctets},
		{"interface_out_ipv6_multi_bytes", ipIfStatsHCOutMcastOctets},
	}
	//--------------------------------Result Processing---------------------------------
	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		index := strings.TrimPrefix(pdu.Name, entry.Oid+".2.")
		m.AddField(index, entry.Name, pdu.Value.(uint64))
	})
	d.AddDataFromEntries(INTERFACE, entries, function)
}

func (d *iosxr) NetworkPolicy() {
	//---------------------------------------OIDs---------------------------------------
	const cbQosCMName = ".1.3.6.1.4.1.9.9.166.1.7.1.1.1"
	const cbQosIFPolicyIndex = ".1.3.6.1.4.1.9.9.166.1.2.1.1.1"
	const cbQosConfigIndex = ".1.3.6.1.4.1.9.9.166.1.5.1.1.2"
	const cbQosParentObjectsIndex = ".1.3.6.1.4.1.9.9.166.1.5.1.1.4"
	const cbQosPolicyMapName = ".1.3.6.1.4.1.9.9.166.1.6.1.1.1"
	const cbQosCMPrePolicyByte64 = ".1.3.6.1.4.1.9.9.166.1.15.1.1.6"
	const cbQosCMDropByte64 = ".1.3.6.1.4.1.9.9.166.1.15.1.1.17"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"cmname", cbQosCMName},
		{"pmname", cbQosPolicyMapName},
		{"policy", cbQosIFPolicyIndex},
		{"config", cbQosConfigIndex},
		{"parent", cbQosParentObjectsIndex},
		{"permit_bytes", cbQosCMPrePolicyByte64},
		{"drop_bytes", cbQosCMDropByte64},
	}
	//--------------------------------Result Processing---------------------------------
	qosCMName := map[string]string{}
	qosPolicyMapName := map[string]string{}
	qosIFPolicy := map[string]string{}
	qosConfigIndex := map[string]string{}
	qosParentIndex := map[string]string{}

	/*
		function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU){
			index := strings.TrimPrefix(pdu.Name, entry.Oid)[1:]
			switch entry.Oid{
			case cbQosCMName:
				qosName[index]   = string(pdu.Value.([]byte))
			case cbQosIFPolicyIndex:
				qosPolicy[strconv.Itoa(int(pdu.Value.(uint)))] = index
			case cbQosConfigIndex:
				qosConfig[index] = strconv.Itoa(int(pdu.Value.(uint)))
			case cbQosCMPrePolicyByte64, cbQosCMDropByte64:
				name  := strings.ToLower(qosName[qosConfig[index]])
				split := strings.Split(qosPolicy[strings.Split(index, ".")[0]], ".")
				index := split[0]
				dir   := map[string]string{"1":"in","2":"out"}[split[1]]
				m.AddField(index, "interface_" + dir + "_" + name + "_" + entry.Name, pdu.Value)
			}
		})
		d.AddDataFromEntries(INTERFACE, entries, function)
	*/

	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		index := strings.TrimPrefix(pdu.Name, entry.Oid)[1:]

		var policyIndex, objectIndex string
		if entry.Oid != cbQosCMName && entry.Oid != cbQosPolicyMapName {
			split := strings.Split(index, ".")
			policyIndex = split[0]
			objectIndex = split[1]
		}

		switch entry.Oid {
		case cbQosCMName:
			qosCMName[index] = string(pdu.Value.([]byte))
		case cbQosPolicyMapName:
			qosPolicyMapName[index] = string(pdu.Value.([]byte))
		case cbQosIFPolicyIndex:
			qosIFPolicy[strconv.Itoa(int(pdu.Value.(uint)))] = index
		case cbQosConfigIndex:
			qosConfigIndex[objectIndex] = strconv.Itoa(int(pdu.Value.(uint)))
		case cbQosParentObjectsIndex:
			qosParentIndex[objectIndex] = strconv.Itoa(int(pdu.Value.(uint)))
		case cbQosCMPrePolicyByte64, cbQosCMDropByte64:
			//Get Classmap Name
			name := strings.ToLower(qosCMName[qosConfigIndex[objectIndex]])

			//Split Interface Index and Direction
			split := strings.Split(qosIFPolicy[policyIndex], ".")
			index := split[0]
			dir := map[string]string{"1": "in", "2": "out"}[split[1]]

			//Get Parent Policy Name
			policyName := qosPolicyMapName[qosConfigIndex[qosParentIndex[objectIndex]]]
			if qosParentIndex[qosParentIndex[qosParentIndex[objectIndex]]] != "0" {
				paidopai := qosCMName[qosConfigIndex[qosParentIndex[qosParentIndex[objectIndex]]]]
				if paidopai != "" {
					policyName = policyName + "." + paidopai
				}
				fmt.Println(policyName)
			}

			//Initialize Tags
			newIndex := index + "_" + qosParentIndex[objectIndex]
			if m.Tags[newIndex]["interface_policy"] == "" {
				for k, v := range m.Tags[index] {
					m.AddTag(newIndex, k, v)
				}
			}

			//Add Tag And Field
			m.AddTag(newIndex, "interface_policy_parent", policyName)
			m.AddField(newIndex, "interface_"+dir+"_"+name+"_"+entry.Name, pdu.Value)
		}
	})
	d.AddDataFromEntries(INTERFACE, entries, function)
}

func (d *iosxr) BgpPeers() {
	//---------------------------------------OIDs---------------------------------------
	const cbgpPeer2AcceptedPrefixes = ".1.3.6.1.4.1.9.9.187.1.2.8.1.1"
	const cbgpPeer2DeniedPrefixes = ".1.3.6.1.4.1.9.9.187.1.2.8.1.2"
	const cbgpPeer2PrefixAdminLimit = ".1.3.6.1.4.1.9.9.187.1.2.8.1.3"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"bgp_accepted_prefixes", cbgpPeer2AcceptedPrefixes},
		{"bgp_denied_prefixes", cbgpPeer2DeniedPrefixes},
		{"bgp_limit_prefixes", cbgpPeer2PrefixAdminLimit},
	}
	//--------------------------------Result Processing---------------------------------
	function := func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(strings.TrimPrefix(pdu.Name, entry.Oid+"."), ".")
		size, _ := strconv.Atoi(split[1])

		var ip string

		switch split[0] {
		case "1":
			ip = strings.Join(split[2:2+size], ".")
		case "2":
			ip = util.GetIPv6Address(split[2 : 2+size])
		}
		m.AddField(ip, entry.Name, pdu.Value)
		m.AddTag(ip, "bgp_neighbour", ip)
	}
	d.AddDataFromEntries(BGP, entries, function)
}

func (d *iosxr) Memory() {
	//Check If Physical Entries Table Was Obtained
	if d.physicalEntries == nil || len(d.physicalEntries) == 0 {
		Log("Physical Entries Was Not Found - Memory Feature Was Cancelled")
		return
	}
	//---------------------------------------OIDs---------------------------------------
	const cempMemPoolName = ".1.3.6.1.4.1.9.9.221.1.1.1.1.3"
	const cempMemPoolHCUsed = ".1.3.6.1.4.1.9.9.221.1.1.1.1.18"
	const cempMemPoolHCFree = ".1.3.6.1.4.1.9.9.221.1.1.1.1.20"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"name", cempMemPoolName},
		{"used_bytes", cempMemPoolHCUsed},
		{"free_bytes", cempMemPoolHCFree},
	}
	//--------------------------------Result Processing---------------------------------
	//Stores The Memory Pool Names
	names := map[string]string{}

	function := func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
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
			m.AddField(index, "memory_"+names[poolIndex]+"_"+entry.Name, pdu.Value.(uint64))
		}
	}
	d.AddDataFromEntries(MEMORY, entries, function)
}

func (d *iosxr) Cpu() {
	//Check If Physical Entries Table Was Obtained
	if d.physicalEntries == nil || len(d.physicalEntries) == 0 {
		Log("Physical Entries Was Not Found - CPU Feature Was Cancelled")
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

func (d *iosxr) Sensors() {
	//Check If Physical Entries Table Was Obtained
	if d.physicalEntries == nil || len(d.physicalEntries) == 0 {
		Log("Physical Entries Was Not Found - Sensors Feature Was Cancelled")
		return
	}

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

	sensorData := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
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

func (d *iosxr) Fetch(dat *data.Data, s *runner.S) {
	d.device.Fetch(dat, s)
}
