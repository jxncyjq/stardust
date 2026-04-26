//go:build linux

package load

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var cpuUsagePerMille int64 // 0-1000，与 cpuThreshold 单位一致（900 = 90%）

func init() {
	go sampleCPU()
}

func sampleCPU() {
	var prevIdle, prevTotal uint64
	for {
		idle, total := readCPUStat()
		if prevTotal > 0 && total > prevTotal {
			idleDelta := idle - prevIdle
			totalDelta := total - prevTotal
			usage := (1.0 - float64(idleDelta)/float64(totalDelta)) * 1000
			atomic.StoreInt64(&cpuUsagePerMille, int64(usage))
		}
		prevIdle = idle
		prevTotal = total
		time.Sleep(250 * time.Millisecond)
	}
}

func readCPUStat() (idle, total uint64) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return
	}
	fields := strings.Fields(scanner.Text())
	if len(fields) < 8 || fields[0] != "cpu" {
		return
	}
	var vals [8]uint64
	for i := 0; i < 8; i++ {
		vals[i], _ = strconv.ParseUint(fields[i+1], 10, 64)
	}
	// user nice system idle iowait irq softirq steal
	idle = vals[3] + vals[4] // idle + iowait
	for _, v := range vals {
		total += v
	}
	return
}

func getCpuUsage() int64 {
	return atomic.LoadInt64(&cpuUsagePerMille)
}
