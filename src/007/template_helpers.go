package main

import (
	"net"
	"sort"
	"strings"

	linuxproc "github.com/marc-barry/goprocinfo/linux/linux"
)

// InterfaceRateStats implements sort.Interface for []InterfaceRateStat based on Iface (i.e. interface name).
type InterfaceRateStats []InterfaceRateStat

func (l InterfaceRateStats) Len() int           { return len(l) }
func (l InterfaceRateStats) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l InterfaceRateStats) Less(i, j int) bool { return l[i].Iface < l[j].Iface }

// NetworkStats implements sort.Interface for []linuxproc.NetworkStat based on Iface (i.e. interface name).
type NetworkStats []linuxproc.NetworkStat

func (l NetworkStats) Len() int           { return len(l) }
func (l NetworkStats) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
func (l NetworkStats) Less(i, j int) bool { return l[i].Iface < l[j].Iface }

func getInterfaces() []net.Interface {
	collectInterfaces()

	return IfaceList.All()
}

func getInterfaceIPAddressesString(iface net.Interface) string {
	addrs, err := iface.Addrs()
	if err != nil {
		return err.Error()
	}

	var addrStrings = make([]string, 0, len(addrs))
	for _, addr := range addrs {
		addrStrings = append(addrStrings, addr.String())
	}

	return strings.Join(addrStrings, ", ")
}

func getCPUInfo() *linuxproc.CPUInfo {
	info, err := readCPUInfo()
	if err != nil {
		Log.WithField("error", err).Error("Error reading cpuinfo.")
	}
	return info
}

func getNetworkDeviceStats() []linuxproc.NetworkStat {
	stats, err := readNetworkDeviceStats()
	if err != nil {
		Log.WithField("error", err).Error("Error reading network device stats.")
	}

	sort.Sort(NetworkStats(stats))

	return stats
}

func getInterfaceRateStats() []InterfaceRateStat {
	stats := copyInterfaceRateStats()

	sort.Sort(InterfaceRateStats(stats))

	return stats
}
