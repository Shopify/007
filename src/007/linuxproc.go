package main

import (
	"reflect"
	"strings"

	linuxproc "github.com/marc-barry/goprocinfo/linux/linux"
)

const (
	CPUInfoPath      = "/proc/cpuinfo"
	NetworkStatPath  = "/proc/net/dev"
	NetstatStatPath  = "/proc/net/netstat"
	SockstatStatPath = "/proc/net/sockstat"
)

type Stat struct {
	StatName   string
	MetricName string
}

func readCPUInfo() (*linuxproc.CPUInfo, error) {
	return linuxproc.ReadCPUInfo(CPUInfoPath)
}

func readNetworkDeviceStats() ([]linuxproc.NetworkStat, error) {
	return linuxproc.ReadNetworkStat(NetworkStatPath)
}

func getNetworkDeviceStatsList() []Stat {
	stat := linuxproc.NetworkStat{}

	elem := reflect.ValueOf(&stat).Elem()
	typeOfElem := elem.Type()

	list := make([]Stat, 0)

	for i := 0; i < elem.NumField(); i++ {
		if field := typeOfElem.Field(i); field.Name != "Iface" {
			list = append(list, Stat{field.Name, strings.Join([]string{networkDeviceStatsMetricPrefix, field.Tag.Get("json")}, "")})
		}
	}

	return list
}

func readNetstatStats() (*linuxproc.Netstat, error) {
	return linuxproc.ReadNetstat(NetstatStatPath)
}

func getNetstatStatsList() []Stat {
	stat := linuxproc.Netstat{}

	elem := reflect.ValueOf(&stat).Elem()
	typeOfElem := elem.Type()

	list := make([]Stat, 0)

	for i := 0; i < elem.NumField(); i++ {
		field := typeOfElem.Field(i)
		list = append(list, Stat{field.Name, strings.Join([]string{netstatStatsMetricPrefix, field.Tag.Get("json")}, "")})
	}

	return list
}

func readSockstatStats() (*linuxproc.SockStat, error) {
	return linuxproc.ReadSockStat(SockstatStatPath)
}

func getSockstatStatsList() []Stat {
	stat := linuxproc.SockStat{}

	elem := reflect.ValueOf(&stat).Elem()
	typeOfElem := elem.Type()

	list := make([]Stat, 0)

	for i := 0; i < elem.NumField(); i++ {
		field := typeOfElem.Field(i)
		list = append(list, Stat{field.Name, strings.Join([]string{sockstatStatsMetricPrefix, field.Tag.Get("json")}, "")})
	}

	return list
}
