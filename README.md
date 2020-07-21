# Gofetch-SNMP

A software application that collects metrics from network devices through SNMP, exporting them into InfluxDB.

The metrics are collected periodically and simiultaneously from a set of network devices configured by the user.

## Devices

Gofetch-SNMP can poll five types of devices:

* The `generic` device is an abstraction that encompasses most network devices, but is limited on the metrics collected.
* The `ios-xr` device corresponds to a CISCO switch or router with the IOS-XR operating system.
* The `ios` device corresponds to a CISCO switch or router with the IOS operating system.
* The `mrv` device corresponds to a MRV-LX console server.
* The `opengear` device corresponds to an Opengear console server.

## Features

Through SNMP it is possible to obtain a series of metrics, by choosing which features to monitor. 

* The `Uptime` feature gets the elapsed time in seconds since the device was booted.
* The `InterfaceCounters` feature gets the number of inbound and outbound packets that are accepted, discarded or have errors, for each of the device's interface.
* The `NetworkACL` feature gets the number of bytes permitted or dropped for each Access Control List (ACL). 
* The `NetworkPolicy` feature gets the number of bytes permitted or dropped for each Policy Map.
* The `BgpPeers` feature gets the number of accepted, dropped and limit route prefixes for each BGP connection.
* The `CellInfo` feature gets data related to the cellular modem.
* The `Memory` feature gets the amount of used and free memory.
* The `Cpu` feature gets the data related to the CPU utilization.
* The `Sensors` feature gets data related to the device's sensors.

|  | Generic | IOS-XR | IOS | MRV | Opengear |
|-|-|-|-|-|-|
| Uptime | ✓ | ✓ | ✓ | ✓ | ✓ |
| InterfaceCounters | ✓ | ✓ | ✓ | ✓ | ✓ |
| NetworkACL | ✕ | ✓ | ✓ | ✕ | ✕ |
| NetworkPolicy | ✕ | ✓ | ✕ | ✕ | ✕ |
| BGPPeers | ✕ | ✓ | ✓ | ✕ | ✕ |
| CellInfo | ✕ | ✕ | ✕ | ✓ | ✓ |
| Memory | ✕ | ✓ | ✓ | ✕ | ✓ |
| CPU | ✕ | ✓ | ✓ | ✕ | ✓ |
| Sensors | ✕ | ✓ | ✓ | ✓ | ✓ |

## Configurations

The functioning of the application can be configured through YAML files. There must be three different files, for the Application, InfluxDB and Devices respectively.

### Application 

This configuration file defines the general functioning of the application.

* The `version` field indicates the version of Gofetch.
* The `interval` field indicates the time between consecutive collections of metrics.
* The `timeout` field indicates the maximum amount of time the application waits for the response from a device.
* The `maxroutines` field indicates the maximum number of routines the application may create.

```
version: v1.1.0
interval: 1m
timeout: 55s
maxroutines: 2
```

### InfluxDB

This configuration file defines the InfluxDB instance to which the collected metrics are exported.

* The `server` field indicates the address where the InfluxDB instance is running.
* The `username` and `password` fields are used as credentials to access the InfluxDB.
* The `database` field indicates the database where the metrics are stored.
* The `ping` field indicates the duration of the ping that determines whether the InfluxDB instance is available.

```
server: http://my.database.net:8086
username: myusername
password: mypassword
database: mydatabase
ping: 2s
```

### Devices

This configuration file defines the devices from which the metrics are collected. Multiple devices can be queried simultaneously if included in the configurations file.

* There must be a `Hosts` field, with a list of `Host` elements.
* The `IP` field indicates the device's IP address.
* The `Type` field indicates the type of the device being monitored.
* In the `SnmpConfig`, the `Version`, `Port`, `Timeout`, `Retries` and `Community` fields should match the SNMP configurations of the device in order to have access to it.
* In the `Features`, the `Uptime`, `InterfaceCounters`, `NetworkACL`, `NetworkPolicy`, `BgpPeers`, `CellInfo`, `Memory`, `Cpu`and `Sensors` indicate `true` if the feature is monitored and `false` (or ommitted) otherwise.

```
Hosts:
  - Host:
    IP: 192.1.1.1
    Type: cisco-ios-xr
    SnmpConfig:
      Version: 3
      Port: 161
      Timeout: 6
      Retries: 1
      Community: mycommunity
    Features:
      Uptime: true
      InterfaceCounters: true
  - Host:
    IP: 192.1.1.2
    Type: cisco-ios
    SnmpConfig:
      Version: 2
      Port: 161
      Timeout: 6
      Retries: 1
      Community: mycommunity
    Features:
      Uptime: true
      InterfaceCounters: true
```

## Running the Application

The application must be provided with the appropriately structured configuration files.

* The `-c` flag indicates the path to the Application configuration file.
* The `-d` flag indicates the path to the InfluxDB configuration file.
* The `-h` flag indicates the path to the Devices configuration file.

```
gofetch -c config.yml -d db.yml -h hosts.yml
```
