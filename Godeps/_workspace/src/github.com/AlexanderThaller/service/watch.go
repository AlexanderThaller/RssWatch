package service

import (
	"log/syslog"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var (
	// DefaultWatchInterval is the default interval for which we will record
	// values in the Watch functions.
	DefaultWatchInterval = time.Minute * 1
)

func writeSyslog(name, value string) {
	writer, err := syslog.New(syslog.LOG_LOCAL6, "metric."+name)
	if err != nil {
		c := exec.Command("logger", "-p", "local6.info", "-t",
			"metric."+name, value)
		c.Run()
		return
	}

	writer.Info(value)
}

// RecordFloat will take a metricname and a float value and record it to syslog.
func RecordFloat(name string, value float64) {
	message := strconv.FormatFloat(value, 'f', -1, 64)
	writeSyslog(name, message)
}

// RecordUint will take a metricname and a uint value and record it to syslog.
func RecordUint(name string, value uint) {
	message := strconv.FormatUint(uint64(value), 10)
	writeSyslog(name, message)
}

// RecordInt will take a metricname and a int value and record it to syslog.
func RecordInt(name string, value int) {
	message := strconv.Itoa(value)
	writeSyslog(name, message)
}

// WatchFloat will take a function that returns a float value and record the
// output of that function every DefaultWatchInterval.
func WatchFloat(name string, value func() float64) {
	go func() {
		for {
			RecordFloat(name, value())
			time.Sleep(DefaultWatchInterval)
		}
	}()
}

// WatchUint will take a function that returns a uint value and record the
// output of that function every DefaultWatchInterval.
func WatchUint(name string, value func() uint) {
	go func() {
		for {
			RecordUint(name, value())
			time.Sleep(DefaultWatchInterval)
		}
	}()
}

// WatchInt will take a function that returns a int value and record the
// output of that function every DefaultWatchInterval.
func WatchInt(name string, value func() int) {
	go func() {
		for {
			RecordInt(name, value())
			time.Sleep(DefaultWatchInterval)
		}
	}()
}

// WatchRuntimeMemory will record the runtime memory stats every
// DefaultWatchInterval.
func WatchRuntimeMemory(name string) {
	go func() {
		memstats := new(runtime.MemStats)

		for {
			runtime.ReadMemStats(memstats)

			RecordUint(name+".alloc.current", uint(memstats.Alloc))
			RecordUint(name+".alloc.total", uint(memstats.TotalAlloc))

			time.Sleep(DefaultWatchInterval)
		}
	}()
}
