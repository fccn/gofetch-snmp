package devices

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fccn.pt/glopes/gofetch/data"
	. "fccn.pt/glopes/gofetch/log"
	"fccn.pt/glopes/gofetch/snmp"
	"fccn.pt/glopes/gofetch/util"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//----------------------------------------CONSTANTS-----------------------------------------
//------------------------------------------------------------------------------------------
const (
	STATISTICS = "statistics_info"
	UPTIME     = "uptime_info"
	INTERFACE  = "interface_info"
	BGP        = "bgp_info"
	CELL       = "cell_info"
	MEMORY     = "memory_info"
	CPU        = "cpu_info"
	SENSOR     = "sensor_info"
)

//------------------------------------------------------------------------------------------
//----------------------------------------INTERFACES----------------------------------------
//------------------------------------------------------------------------------------------
type Device interface {
	//Initialize()
	Init()
	//Collect Uptime Data
	Uptime()
	//Collect Interface Counters Data
	InterfaceCounters()
	//Collect Acl Data
	NetworkAcl()
	//Collect Policy Data
	NetworkPolicy()
	//Collect Bgp Data
	BgpPeers()
	//Collect GSM Modem Data
	CellInfo()
	//Collect Memory Usage Data
	Memory()
	//Collect CPU Usage Data
	Cpu()
	//Collect Sensor Data
	Sensors()
	//Fetch All Data And Write To InfluxDB
	Fetch(dat *data.Data, s *runner.S)
}

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
//Struct That Receives Host Information From YAML
type Host struct {
	IP         string     `yaml:"IP"`
	Type       string     `yaml:"Type"`
	SnmpConfig snmpconfig `yaml:"SnmpConfig"`
	Features   features   `yaml:"Features"`
}

//Struct That Receives Host Snmp Configurations From YAML
type snmpconfig struct {
	Version   int    `yaml:"Version"`
	Timeout   int    `yaml:"Timeout"`
	Retries   int    `yaml:"Retries"`
	Port      uint16 `yaml:"Port"`
	Community string `yaml:"Community"`
	Flags     string `yaml:"Flags"`
	Username  string `yaml:"Username"`
	AuthProt  string `yaml:"AuthProt"`
	AuthPass  string `yaml:"AuthPass"`
	PrivProt  string `yaml:"PrivProt"`
	PrivPass  string `yaml:"PrivPass"`
}

//Struct That Receives Host Features Information From YAML
type features struct {
	GofetchStatistics bool `yaml:"GofetchStatistics"`
	Uptime            bool `yaml:"Uptime"`
	InterfaceCounters bool `yaml:"InterfaceCounters"`
	NetworkAcl        bool `yaml:"NetworkAcl"`
	NetworkPolicy     bool `yaml:"NetworkPolicy"`
	BgpPeers          bool `yaml:"BgpPeers"`
	CellInfo          bool `yaml:"CellInfo"`
	Memory            bool `yaml:"Memory"`
	Cpu               bool `yaml:"Cpu"`
	Sensors           bool `yaml:"Sensors"`
}

//Struct That Carries All Hosts' Information From YAML
type Hosts struct {
	Hosts []Host `yaml:"Hosts"`
}

//Defines a Generic Device With All The Features Common Among The Devices
type device struct {
	Device              //Implements The Device Interface
	IP       string     //Indicates The Device's IP Address
	Type     string     //Indicates The Device's Type
	SnmpConf g.GoSNMP   //SNMP Configurations Struct
	Features features   //Features Activated For Fetching Data
	Data     *data.Data //Data Collected For Each Of The Device's Metrics
	Bulk     bool       //Indicates If Device Can Use BulkWalk
	Cancel   bool       //Indicates That Fetch Should Not Run
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func NewDevice(host Host) *device {
	//Create A Struct With The SNMP Configurations Specified For This Host
	snmpConf := g.GoSNMP{
		Target:         host.IP,
		Port:           host.SnmpConfig.Port,
		Timeout:        time.Duration(host.SnmpConfig.Timeout) * time.Second,
		Retries:        host.SnmpConfig.Retries,
		MaxOids:        g.MaxOids,
		MaxRepetitions: 100,
	}

	//Define The Specific Configurations For The Chosen SNMP Version
	switch host.SnmpConfig.Version {

	//No need for "case 1", since no new information is added to the struct

	//SNMPv2 Introduces A "Community" Parameter
	case 2:
		snmpConf.Version = g.Version2c
		snmpConf.Community = host.SnmpConfig.Community

	//SNMPv3 Introduces Authentication And Privacy Protocols And Passwords For Each
	case 3:
		flagsmap := map[string]g.SnmpV3MsgFlags{
			"NoAuthNoPriv": g.NoAuthNoPriv, //No Authentication & No Privacy
			"AuthNoPriv":   g.AuthNoPriv,   //Authentication & No Privacy
			"AuthPriv":     g.AuthPriv,     //Authentication & Privacy
			"Reportable":   g.Reportable,   //Report PDU must be sent
		}

		authmap := map[string]g.SnmpV3AuthProtocol{
			"NoAuth": g.NoAuth, //No Authentication
			"SHA":    g.SHA,    //Secure Hash Algorithm
			"MD5":    g.MD5,    //Message-Digest Algorithm 5
		}

		privmap := map[string]g.SnmpV3PrivProtocol{
			"NoPriv": g.NoPriv, //No Privacy
			"AES":    g.AES,    //Advanced Encryption Standard
			"DES":    g.DES,    //Data Encryption Standard
		}

		snmpConf.Version = g.Version3
		snmpConf.MsgFlags = flagsmap[host.SnmpConfig.Flags]
		snmpConf.SecurityModel = g.UserSecurityModel
		snmpConf.SecurityParameters = &g.UsmSecurityParameters{
			UserName:                 host.SnmpConfig.Username,
			AuthenticationProtocol:   authmap[host.SnmpConfig.AuthProt],
			AuthenticationPassphrase: host.SnmpConfig.AuthPass,
			PrivacyProtocol:          privmap[host.SnmpConfig.PrivProt],
			PrivacyPassphrase:        host.SnmpConfig.PrivPass,
		}
	}

	//Returned Device Must Have A Device Struct, Which Is Initialized Here
	d := device{
		SnmpConf: snmpConf,
		Features: host.Features,
		IP:       host.IP,
		Type:     strings.ToLower(host.Type),
	}

	return &d
}

func (d *device) GetSpecific() Device {
	switch d.Type {
	case "cisco-ios-xr":
		d.Bulk = true
		return &iosxr{device: d}
	case "cisco-ios":
		d.Bulk = true
		return &ios{device: d}
	case "opengear":
		d.Bulk = true
		return &opengear{device: d}
	case "mrv":
		return &mrv{device: d}
	case "junos":
		//d.Bulk = true
		return &ios{device: d}
	default:
		return &generic{device: d}
	}
}

func (d *device) Init() {
	//Get The Device's Tags (Device Name)
	d.GetTags()

	//Get The Interfaces' Tags
	d.GetInterfaceTags()

	//Initialize Statistics Metric
	if d.Features.GofetchStatistics {
		d.Data.AddMetric(STATISTICS)
	}
}

func (d *device) AddDataFromEntries(metric string, entries data.Entries, function data.Function) {
	d.Data.AddFromEntries(d.SnmpConf, d.Bulk, metric, entries, function)
}

func (d *device) AddMetricTagsFromEntries(metric string, entries data.Entries) {
	d.Data.AddFromEntries(d.SnmpConf, d.Bulk, metric, entries, data.AddTags)
}

func (d *device) AddMetricFieldsFromEntries(metric string, entries data.Entries) {
	d.Data.AddFromEntries(d.SnmpConf, d.Bulk, metric, entries, data.AddFields)
}

func (d *device) GetTags() {
	//Return If Tag Name Was Obtained Already
	//---------------------------------------OIDs---------------------------------------
	const sysName = ".1.3.6.1.2.1.1.5.0"
	//----------------------------------SNMP Requests-----------------------------------
	name := snmp.Get(d.SnmpConf, []string{sysName})
	//--------------------------------Result Processing---------------------------------
	if name != nil && len(name.Variables) > 0 {
		d.Data.AddTag("device_name", strings.ToLower(string(name.Variables[0].Value.([]byte))))
		d.Data.AddTag("device_ip", d.IP)
		d.Data.AddTag("device_type", d.Type)
	} else {
		d.Cancel = true
	}
}

func (d *device) GetInterfaceTags() {
	if d.Cancel || (!d.Features.NetworkPolicy && !d.Features.InterfaceCounters) {
		return
	}
	//---------------------------------------OIDs---------------------------------------
	const ifDescr = ".1.3.6.1.2.1.2.2.1.2"
	const ifName = ".1.3.6.1.2.1.31.1.1.1.1"
	const ifAlias = ".1.3.6.1.2.1.31.1.1.1.18"
	const ipAddrIfIndex = ".1.3.6.1.2.1.4.20.1.2"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"interface_descr", ifDescr},
		{"interface_name", ifName},
		{"interface_alias", ifAlias},
		{"interface_addr", ipAddrIfIndex},
	}
	//--------------------------------Result Processing---------------------------------
	function := func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		switch entry.Oid {
		case ifDescr, ifName, ifAlias:
			m.AddTag(snmp.GetIndex(pdu, entry.Oid), entry.Name, string(pdu.Value.([]uint8)))
		case ipAddrIfIndex:
			index := strconv.Itoa(pdu.Value.(int))
			var ip string
			if tag := m.Tags[index][entry.Name]; tag != "" {
				ip = tag + ";"
			}
			m.AddTag(index, entry.Name, ip+strings.TrimPrefix(pdu.Name, entry.Oid+"."))
		}
	}
	d.AddDataFromEntries(INTERFACE, entries, function)
}

func (d *device) Uptime() {
	if d.Cancel || !d.Features.Uptime {
		return
	}
	//---------------------------------------OIDs---------------------------------------
	const snmpEngineTime = ".1.3.6.1.6.3.10.2.1.3"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"uptime_seconds", snmpEngineTime},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(UPTIME, entries)
}

func (d *device) InterfaceCounters() {
	if d.Cancel || !d.Features.InterfaceCounters {
		return
	}
	//---------------------------------------OIDs---------------------------------------
	const ifInDiscards = ".1.3.6.1.2.1.2.2.1.13"
	const ifOutDiscards = ".1.3.6.1.2.1.2.2.1.19"
	const ifInErrors = ".1.3.6.1.2.1.2.2.1.14"
	const ifOutErrors = ".1.3.6.1.2.1.2.2.1.20"
	const ifHCInOctets = ".1.3.6.1.2.1.31.1.1.1.6"
	const ifHCOutOctets = ".1.3.6.1.2.1.31.1.1.1.10"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"interface_in_discards", ifInDiscards},
		{"interface_out_discards", ifOutDiscards},
		{"interface_in_errors", ifInErrors},
		{"interface_out_errors", ifOutErrors},
		{"interface_in_hc_bytes", ifHCInOctets},
		{"interface_out_hc_bytes", ifHCOutOctets},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(INTERFACE, entries)
}

func (d *device) Fetch(dat *data.Data, s *runner.S) {
	//For Statistic Purposes
	start := time.Now()

	//----------------------------------Initialization----------------------------------
	//Start SNMP Connection
	if err := d.SnmpConf.Connect(); err != nil {
		FatalLog(fmt.Sprintf("SNMP Connect() err: %v", err))
	}

	//Initialize Device Data
	d.Data = dat

	//Based On The Type Configured, Initialize A Specific Device
	dev := d.GetSpecific()

	//Variables Initialization
	dev.Init()

	defer func() {
		//Close Connection In The End, Or If Something Goes Wrong
		d.SnmpConf.Conn.Close()

		//Set The Timestamp From When The Data Was Collected
		d.Data.SetTimestamp(time.Now())

		//Time Passed Since Fetch Started
		delta := time.Now().Sub(start)

		//Collect Performance Statistics
		if d.Features.GofetchStatistics {
			d.Data.GetMetric(STATISTICS).AddField("0", "statistics_fetch_seconds", delta.Seconds())
		}
	}()

	//Return If Cancel
	if d.Cancel {
		Log("Fetch Was Cancelled")
		return
	}

	//-------------------------------------Features-------------------------------------
	features := []struct {
		n string
		c bool
		f func()
	}{
		{"uptime", d.Features.Uptime, dev.Uptime},
		{"interface_counters", d.Features.InterfaceCounters, dev.InterfaceCounters},
		{"network_acl", d.Features.NetworkAcl, dev.NetworkAcl},
		{"network_policy", d.Features.NetworkPolicy, dev.NetworkPolicy},
		{"bgp_peers", d.Features.BgpPeers, dev.BgpPeers},
		{"cell_info", d.Features.CellInfo, dev.CellInfo},
		{"memory", d.Features.Memory, dev.Memory},
		{"cpu", d.Features.Cpu, dev.Cpu},
		{"sensors", d.Features.Sensors, dev.Sensors},
	}

	for _, feature := range features {
		d.CollectFeature(feature.n, feature.c, feature.f)
		if (*s)() {
			Log("Fetch Was Interrupted")
			return
		}
	}

	//Debug Fetch Time
	DebugLog(d.Data.GetTag("device_name") + "'s Data Has Been Collected. Duration: " + time.Now().Sub(start).String())
}

func (d *device) CollectFeature(featureName string, featureEnabled bool, featureFunc func()) {
	if featureEnabled {
		duration := util.FunctionDuration(featureFunc)
		if d.Features.GofetchStatistics {
			d.Data.GetMetric(STATISTICS).AddField("0", "statistics_"+featureName+"_seconds", duration)
		}
	}
}
