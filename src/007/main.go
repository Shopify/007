package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ooyala/go-dogstatsd"
)

const (
	DogstatsdAddressFlag     = "dogstatsd-address"
	HTTPEnableFlag           = "http-enable"
	HTTPPortFlag             = "http-port"
	CalculateRate            = "calculate-rate"
	CollectRate              = "collect-rate"
	LogRate                  = "log-rate"
	StatsInterfaceFilterFlag = "stats-interface-filter"
	StatsFlag                = "stats"
	JsonLogFlag              = "json-log"
	HelpFlag                 = "h"
)

var (
	Log                            = logrus.New()
	StatsdClient *dogstatsd.Client = nil

	statsdAddress        = flag.String(DogstatsdAddressFlag, "", "The address of the Datadog DogStatsd server.")
	httpEnable           = flag.Bool(HTTPEnableFlag, false, "Enable HTTP server.")
	httpPort             = flag.Int(HTTPPortFlag, 8001, "HTTP server listening port.")
	calculatedRate       = flag.Int64(CalculateRate, 2, "Rate (in seconds) for which the rate stats are calculated.")
	collectRate          = flag.Int64(CollectRate, 2, "Rate (in seconds) for which the stats are collected.")
	logRate              = flag.Int64(LogRate, 5, "Rate (in seconds) for which the stats are logged.")
	statsInterfaceFilter = flag.String(StatsInterfaceFilterFlag, "", "Regular expression which filters out interfaces not reported to DogStatd.")
	statsList            = flag.String(StatsFlag, "", "The list of stats send to the DogStatsd server.")
	jsonLog              = flag.Bool(JsonLogFlag, false, "Enable JSON logging. The default format is text.")
	help                 = flag.Bool(HelpFlag, false, "Prints help info.")

	stopOnce sync.Once
	stopWg   sync.WaitGroup

	IfaceList                  = NewInterfaceList()
	ifaceRegExp *regexp.Regexp = nil
	StatsMap                   = make(map[string]string)
)

func withLogging(f func()) {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Recovered from panic(%+v)", r)

			Log.WithField("error", err).Panicf("Stopped with panic: %s", err.Error())
		}
	}()

	f()
}

func main() {
	flag.Parse()

	if *help {
		fmt.Println("Supported stats:")
		fmt.Print("---> ")
		fmt.Println(NetworkStatPath)
		for _, stat := range getNetworkDeviceStatsList() {
			fmt.Println(stat.StatName + ":" + stat.MetricName)
		}
		fmt.Println("<---")
		fmt.Print("---> ")
		fmt.Println(NetstatStatPath)
		for _, stat := range getNetstatStatsList() {
			fmt.Println(stat.StatName + ":" + stat.MetricName)
		}
		fmt.Println("<---")
		fmt.Println(SockstatStatPath)
		for _, stat := range getSockstatStatsList() {
			fmt.Println(stat.StatName + ":" + stat.MetricName)
		}
		fmt.Println("<---")
		fmt.Println(SnmpStatPath)
		for _, stat := range getSnmpStatsList() {
			fmt.Println(stat.StatName + ":" + stat.MetricName)
		}
		os.Exit(0)
	}

	if *jsonLog {
		Log.Formatter = new(logrus.JSONFormatter)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			Log.WithField("signal", sig).Infof("Signalled. Shutting down.")

			stopOnce.Do(func() { shutdown(0) })
		}
	}()

	if *statsdAddress != "" {
		Log.WithField("address", *statsdAddress).Infof("Attempting to dial DogStatsd server.")
		var err error
		StatsdClient, err = dogstatsd.New(*statsdAddress)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"address": *statsdAddress,
				"error":   err,
			}).Warn("Unable to dial StatsD.")
		}
	}

	if StatsdClient != nil {
		Log.WithField("address", *statsdAddress).Infof("Dialed DogStatsd server.")
		StatsdClient.Namespace = "007."
	}

	var err error
	if ifaceRegExp, err = regexp.Compile(*statsInterfaceFilter); err != nil {
		Log.WithField("error", err).Errorf("Unable to compile provided regular expression: %s", *statsInterfaceFilter)
	}

	if ifaceRegExp != nil {
		Log.WithField("regex", ifaceRegExp.String()).Infof("Compiled interface filter regualr expression.")
	}

	startCalculators()
	startLoggers()

	if StatsdClient != nil {
		for _, stat := range strings.Split(*statsList, ",") {
			StatsMap[stat] = stat
		}
		startCollectors()
	}

	if *httpEnable {
		if err := <-StartHTTPServer(*httpPort); err != nil {
			Log.WithField("error", err).Fatal("Error starting HTTP server.")
		}

		return
	}

	// If HTTP is not enabled we need to block with a wait on a WaitGroup.
	stopWg.Add(1)
	stopWg.Wait()
}

func shutdown(code int) {
	Log.WithField("code", code).Infof("Stopping.")

	// If HTTP is enabled we must exit in order to cause the HTTP server to shutdown.
	if *httpEnable {
		os.Exit(0)
	}

	stopWg.Done()
}

func startCalculators() {
	Log.Info("Starting calculators...")
	if err := calculateInterfaceRateStats(); err != nil {
		Log.WithField("error", err).Error("Error calculating interface rate stats.")
	} else {
		go withLogging(func() {
			c := time.Tick(time.Duration(*calculatedRate) * time.Second)
			for {
				select {
				case <-c:
					if err := calculateInterfaceRateStats(); err != nil {
						Log.WithField("error", err).Error("Error calculating interface rate stats.")
					}
				}
			}
		})
	}
}

func startLoggers() {
	Log.Info("Starting loggers...")
	go withLogging(func() {
		c := time.Tick(time.Duration(*logRate) * time.Second)
		for {
			select {
			case <-c:
				logNetworkDeviceStats()
				logNetstatStats()
				logSockstatStats()
				logSnmpStats()
			}
		}
	})
}

func startCollectors() {
	Log.Info("Starting collectors...")
	go withLogging(func() {
		c := time.Tick(time.Duration(*collectRate) * time.Second)
		for {
			select {
			case <-c:
				collectNetworkDeviceStats()
				collectNetstatStats()
				collectSockstatStats()
				collectSnmpStats()
			}
		}
	})
}
