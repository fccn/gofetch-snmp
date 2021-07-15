package devices

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"strings"

	"github.com/fccn/gofetch-snmp/data"
	"github.com/matryer/runner"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//-----------------------------------------STRUCTS------------------------------------------
//------------------------------------------------------------------------------------------
type ntp struct {
	*device //Extends Device Struct
}

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
func (d *ntp) Init() {
	d.device.Init()

	//Unsupported Features
	d.Features.NetworkAcl = false
	d.Features.NetworkPolicy = false
	d.Features.BgpPeers = false
	d.Features.CellInfo = false
}

func (d *ntp) Uptime() {
	d.device.Uptime()
}

func (d *ntp) InterfaceCounters() {
	d.device.InterfaceCounters()
}

func (d *ntp) Ntp() {
	//---------------------------------------OIDs---------------------------------------
	const mbgLtNgNtpStratum = "1.3.6.1.4.1.5597.30.0.2.2"
	const mbgLtNgNtpRefclockOffset = "1.3.6.1.4.1.5597.30.0.2.4"
	const mbgLtNgFdmFreq = "1.3.6.1.4.1.5597.30.0.4.1"
	const mbgLtNgNtpCCTotalRequestsCurrentDay = "1.3.6.1.4.1.5597.30.0.2.8.5"
	const mbgLtNgNtpCCTotalRequestsLastMinute = "1.3.6.1.4.1.5597.30.0.2.8.7"
	const mbgLtNgNtpCCTodaysClients = "1.3.6.1.4.1.5597.30.0.2.8.8"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"ntp_stratum", mbgLtNgNtpStratum},
		{"ntp_clock_offset", mbgLtNgNtpRefclockOffset},
		{"ntp_frequency", mbgLtNgFdmFreq},
		{"ntp_requests_current_day", mbgLtNgNtpCCTotalRequestsCurrentDay},
		{"ntp_requests_last_minute", mbgLtNgNtpCCTotalRequestsLastMinute},
		{"ntp_clients", mbgLtNgNtpCCTodaysClients},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(NTP, entries)
}

func (d *ntp) Memory() {
	//---------------------------------------OIDs---------------------------------------
	const memTotalReal = ".1.3.6.1.4.1.2021.4.5"
	const memTotalFree = ".1.3.6.1.4.1.2021.4.11"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"memory_total", memTotalReal},
		{"memory_free", memTotalFree},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(MEMORY, entries)
}

func (d *ntp) Cpu() {
	//---------------------------------------OIDs---------------------------------------
	const ssCpuRawUser = ".1.3.6.1.4.1.2021.11.50"
	const ssCpuRawSystem = ".1.3.6.1.4.1.2021.11.52"
	const ssCpuRawIdle = ".1.3.6.1.4.1.2021.11.53"
	const ssCpuRawWait = ".1.3.6.1.4.1.2021.11.54"
	const ssCpuRawKernel = ".1.3.6.1.4.1.2021.11.55"
	//-------------------------------------Entries--------------------------------------
	entries := data.Entries{
		{"cpu_user", ssCpuRawUser},
		{"cpu_system", ssCpuRawSystem},
		{"cpu_idle", ssCpuRawIdle},
		{"cpu_wait", ssCpuRawWait},
		{"cpu_kernel", ssCpuRawKernel},
	}
	//--------------------------------Result Processing---------------------------------
	d.AddMetricFieldsFromEntries(CPU, entries)
}

func (d *ntp) Sensors() {
	//---------------------------------------OIDs---------------------------------------
	const mbgLtNgSysPsIndex = "1.3.6.1.4.1.5597.30.0.5.0.2.1.1"
	const mbgLtNgSysPsStatus = "1.3.6.1.4.1.5597.30.0.5.0.2.1.2"
	const mbgLtNgSysFanIndex = "1.3.6.1.4.1.5597.30.0.5.1.2.1.1"
	const mbgLtNgSysFanStatus = "1.3.6.1.4.1.5597.30.0.5.1.2.1.2"
	const mbgLtNgSysFanError = "1.3.6.1.4.1.5597.30.0.5.1.2.1.3"
	const mbgLtNgSysTempCelsius = "1.3.6.1.4.1.5597.30.0.5.2.1"
	//-------------------------------------Entries--------------------------------------
	power := data.Entries{
		{"sensor_status", mbgLtNgSysPsStatus},
	}
	fan := data.Entries{
		{"sensor_status", mbgLtNgSysFanStatus},
		{"sensor_error", mbgLtNgSysFanError},
	}
	temp := data.Entries{
		{"sensor_value_celsius", mbgLtNgSysTempCelsius},
	}
	//--------------------------------Result Processing---------------------------------
	function := data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(pdu.Name, ".")
		index := split[len(split)-4] + "." + split[len(split)-1]
		m.AddTag(index, "sensor_descr", "Power Supply "+split[len(split)-1])
		m.AddField(index, entry.Name, pdu.Value)
	})
	d.AddDataFromEntries(SENSOR, power, function)

	function = data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		split := strings.Split(pdu.Name, ".")
		index := split[len(split)-4] + "." + split[len(split)-1]
		m.AddTag(index, "sensor_descr", "Fan "+split[len(split)-1])
		m.AddField(index, entry.Name, pdu.Value)
	})
	d.AddDataFromEntries(SENSOR, fan, function)

	function = data.Function(func(m data.Metric, entry data.Entry, pdu g.SnmpPDU) {
		m.AddTag("", "sensor_descr", "Temperature")
		m.AddField("", entry.Name, float64(pdu.Value.(uint)))
	})
	d.AddDataFromEntries(SENSOR, temp, function)
}

func (d *ntp) Fetch(dat *data.Data, s *runner.S) {
	d.device.Fetch(dat, s)
}
