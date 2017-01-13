package main

import (
	"sync"
	"time"
)

var (
	interfaceRatesMu = sync.RWMutex{}
	interfaceRates   = make(map[string]InterfaceRateStat)
)

type InterfaceRateStat struct {
	Iface        string
	oldRxBytes   uint64
	oldRxPackets uint64
	oldTxBytes   uint64
	oldTxPackets uint64

	RxBytesRate   string // bps with SI prefix
	RxPacketsRate string // pps with SI prefix
	TxBytesRate   string // bps with SI prefix
	TxPacketsRate string // pps with SI prefix

	lastCheckTime time.Time
}

func calculateInterfaceRateStats() error {
	interfaceRatesMu.Lock()
	defer interfaceRatesMu.Unlock()

	stats, err := readNetworkDeviceStats()

	if err != nil {
		return err
	}

	for _, stat := range stats {
		nowTime := time.Now()

		if ifaceRateStats, exists := interfaceRates[stat.Iface]; exists {
			durationInSeconds := nowTime.Sub(ifaceRateStats.lastCheckTime).Seconds()

			ifaceRateStats.RxBytesRate = SI(float64((stat.RxBytes-ifaceRateStats.oldRxBytes)*8)/durationInSeconds, 2, " ", "bit/s")
			ifaceRateStats.RxPacketsRate = SI(float64(stat.RxPackets-ifaceRateStats.oldRxPackets)/durationInSeconds, 2, " ", "packet/s")
			ifaceRateStats.TxBytesRate = SI(float64((stat.TxBytes-ifaceRateStats.oldTxBytes)*8)/durationInSeconds, 2, " ", "bit/s")
			ifaceRateStats.TxPacketsRate = SI(float64(stat.TxPackets-ifaceRateStats.oldTxPackets)/durationInSeconds, 2, " ", "packet/s")

			ifaceRateStats.oldRxBytes = stat.RxBytes
			ifaceRateStats.oldRxPackets = stat.RxPackets
			ifaceRateStats.oldTxBytes = stat.TxBytes
			ifaceRateStats.oldTxPackets = stat.TxPackets

			ifaceRateStats.lastCheckTime = nowTime

			interfaceRates[stat.Iface] = ifaceRateStats
		} else {
			interfaceRates[stat.Iface] = InterfaceRateStat{
				Iface:         stat.Iface,
				oldRxBytes:    stat.RxBytes,
				oldRxPackets:  stat.RxPackets,
				oldTxBytes:    stat.TxBytes,
				oldTxPackets:  stat.TxPackets,
				RxBytesRate:   "",
				RxPacketsRate: "",
				TxBytesRate:   "",
				TxPacketsRate: "",
				lastCheckTime: nowTime}
		}
	}

	return nil
}

func copyInterfaceRateStats() []InterfaceRateStat {
	interfaceRatesMu.RLock()
	defer interfaceRatesMu.RUnlock()

	rates := make([]InterfaceRateStat, len(interfaceRates))

	i := 0
	for _, stat := range interfaceRates {
		rates[i] = stat
		i++
	}

	return rates
}
