package data

import (
	. "fccn.pt/glopes/gofetch/log"
	"fccn.pt/glopes/gofetch/snmp"
	g "github.com/soniah/gosnmp"
	"time"
)

type Data struct{
	Timestamp time.Time 		`json:"timestamp"`
	Tags      map[string]string `json:"tags"`
	Metrics   map[string]Metric `json:"metrics"`
}

func NewData()(d Data) {
	d 		  = Data{}
	d.Tags    = map[string]string{}
	d.Metrics = map[string]Metric{}
	return
}

func (d *Data)AddTag(name, value string){
	d.Tags[name] = value
}

func (d *Data)GetTag(name string)string{
	return d.Tags[name]
}

func (d *Data)AddMetric(metric string){
	d.Metrics[metric] = newMetric()
}

func (d *Data)GetMetric(metric string)*Metric{
	m := d.Metrics[metric]
	return &m
}

func (d *Data)AddFromEntries(snmpConf g.GoSNMP, bulk bool, metric string, entries Entries, function Function){
	//Initialize The Metric If It Wasn't Initialized Already
	m := d.GetMetric(metric)
	if m.IsEmpty(){
		d.AddMetric(metric)
		m = d.GetMetric(metric)
	}

	//Go Through Each Entry, Make SNMP Request, Process It
	for i := range entries{
		entry := entries[i]
		pdus  := snmp.WalkAll(snmpConf, bulk, entry.Oid)
		for j := range pdus{
			pdu := pdus[j]
			function(*m, entry, pdu)
		}
	}
}

/*
 * Call This When All Features Are Collected, To Associate Them To A Timestamp
 */
func (d *Data)SetTimestamp(Timestamp time.Time){
	d.Timestamp = Timestamp
}

/*
 * Call This To Write The Data To The InfluxDB
 */
func (d *Data)WriteInflux()bool{
	bp := InfluxCreateBatch()
	for name := range d.Metrics {
		m := d.Metrics[name]
		for index := range m.Fields {
			tags   := m.Tags[index]
			fields := m.Fields[index]

			//Tags Can Be Nil, Initialize In That Case
			if tags == nil {
				tags = map[string]string{}
			}
			//Add Data Main Tags To The Metric Defined Tags
			for k, v := range d.Tags{
				tags[k] = v
			}
			InfluxAddPoint(bp, name, &tags, &fields, d.Timestamp)
		}
	}
	return InfluxWriteBatch(bp)
}

type Metric struct{
	Tags   map[string]map[string]string 	 `json:"tags"`
	Fields map[string]map[string]interface{} `json:"fields"`
}

//"Constructor"
func newMetric()(m Metric) {
	m 		 = Metric{}
	m.Tags 	 = map[string]map[string]string{}
	m.Fields = map[string]map[string]interface{}{}
	return
}

//Add Tag For Given Index
func (m *Metric)AddTag(index string, name, value string) {
	m.initTag(index)
	m.Tags[index][name] = value
}

//Add Field For Given Index
func (m *Metric)AddField(index string, name string, value interface{}) {
	m.initField(index)

	switch value.(type) {
	case uint64:
		value = int64(value.(uint64))
	}

	m.Fields[index][name] = value
}

func (m *Metric)IsEmpty()bool{
	return m.Tags == nil && m.Fields == nil
}

//Initialize Map In Case It Doesn't Exist For Given Index
func (m *Metric)initTag(index string){
	if m.Tags[index] == nil{
		m.Tags[index] = map[string]string{}
	}
}

//Initialize Map In Case It Doesn't Exist For Given Index
func (m *Metric)initField(index string){
	if m.Fields[index] == nil{
		m.Fields[index] = map[string]interface{}{}
	}
}

//Checks If Tags Are Empty
func (m *Metric)hasTags(index string)bool{
	if m.Tags[index] == nil || len(m.Tags[index]) == 0{
		DebugLog("No Tags Were Found")
		return false
	}
	return true
}

//Checks If Fields Are Empty
func (m *Metric)hasFields(index string)bool{
	if m.Fields[index] == nil || len(m.Fields[index]) == 0{
		DebugLog("No Fields Were Found")
		return false
	}
	return true
}