package snmp

//------------------------------------------------------------------------------------------
//-----------------------------------------IMPORTS------------------------------------------
//------------------------------------------------------------------------------------------
import (
	"fmt"
	"strings"

	. "github.com/fccn/gofetch/log"
	g "github.com/soniah/gosnmp"
)

//------------------------------------------------------------------------------------------
//----------------------------------------FUNCTIONS-----------------------------------------
//------------------------------------------------------------------------------------------
//Returns The Index, Given An SnmpPDU And A Prefix
func GetIndex(pdu g.SnmpPDU, oid string) (index string) {
	if hasPrefix(pdu, oid) {
		start := len(oid) + 1 //After The Prefix And After The "."
		end := len(pdu.Name)  //End Of The String
		if next := strings.Index(pdu.Name[start:], "."); next != -1 {
			end = start + next //If There's More Than Just The Index End After The Next "."
		}
		index = pdu.Name[start:end]
	}
	return
}

//Checks If An Entry Has Prefix In Name
func hasPrefix(pdu g.SnmpPDU, prefix string) bool {
	if strings.LastIndex(pdu.Name, ".") != len(prefix)-1 {
		prefix += "." //Add A "." In Case Prefix Doesn't End With "." For Trimming Purposes
	}
	return strings.HasPrefix(pdu.Name, prefix)
}

func WalkAll(snmpConf g.GoSNMP, bulk bool, oid string) (result []g.SnmpPDU) {
	var err error
	if bulk {
		if result, err = snmpConf.BulkWalkAll(oid); err != nil {
			Log(fmt.Sprintf("%s - Could Not Perform Snmp BulkWalkAll - %s: %s", snmpConf.Target, oid, err.Error()))
		}
	} else {
		if result, err = snmpConf.WalkAll(oid); err != nil {
			Log(fmt.Sprintf("%s - Could Not Perform Snmp WalkAll - %s: %s", snmpConf.Target, oid, err.Error()))
		}
	}
	return
}

func Get(snmpConf g.GoSNMP, oids []string) (result *g.SnmpPacket) {
	var err error
	if result, err = snmpConf.Get(oids); err != nil {
		Log(fmt.Sprintf("%s - Could Not Perform Snmp Get - %s: %s", snmpConf.Target, oids, err.Error()))
	}
	return
}
