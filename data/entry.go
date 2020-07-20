package data

import (
	"github.com/fccn/gofetch/snmp"
	g "github.com/soniah/gosnmp"
)

//Struct That Stores A Tag/Field Name And The Oid Used To Get Its Value
type Entry struct{
	Name string
	Oid  string
}

//Function Called To Obtain A Tag/Field Value From An Entry
type Function func(metric Metric, entry Entry, pdu g.SnmpPDU)

type Entries []Entry

var AddTags = Function(func(metric Metric, entry Entry, pdu g.SnmpPDU){
	index := snmp.GetIndex(pdu, entry.Oid)
	var value string
	switch pdu.Value.(type){
	case []uint8:
		value = string(pdu.Value.([]uint8))
	case string:
		value = pdu.Value.(string)
	default:
		return
	}
	metric.AddTag(index, entry.Name, value)
})

var AddFields = Function(func(metric Metric, entry Entry, pdu g.SnmpPDU){
	index := snmp.GetIndex(pdu, entry.Oid)
	metric.AddField(index, entry.Name, pdu.Value)
})