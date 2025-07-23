package agent

import (
	"github.com/Nikolay961996/metsys/models"
	"math/rand"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func Poll(metrics *Metrics) {
	var stats runtime.MemStats
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	runtime.ReadMemStats(&stats)

	metrics.Alloc = float64(stats.Alloc)
	metrics.BuckHashSys = float64(stats.BuckHashSys)
	metrics.Frees = float64(stats.Frees)
	metrics.GCCPUFraction = stats.GCCPUFraction
	metrics.GCSys = float64(stats.GCSys)
	metrics.HeapAlloc = float64(stats.HeapAlloc)
	metrics.HeapIdle = float64(stats.HeapIdle)
	metrics.HeapInuse = float64(stats.HeapInuse)
	metrics.HeapObjects = float64(stats.HeapObjects)
	metrics.HeapReleased = float64(stats.HeapReleased)
	metrics.HeapSys = float64(stats.HeapSys)
	metrics.LastGC = float64(stats.LastGC)
	metrics.Lookups = float64(stats.Lookups)
	metrics.MCacheInuse = float64(stats.MCacheInuse)
	metrics.MCacheSys = float64(stats.MCacheSys)
	metrics.MSpanInuse = float64(stats.MSpanInuse)
	metrics.MSpanSys = float64(stats.MSpanSys)
	metrics.Mallocs = float64(stats.Mallocs)
	metrics.NextGC = float64(stats.NextGC)
	metrics.NumForcedGC = float64(stats.NumForcedGC)
	metrics.NumGC = float64(stats.NumGC)
	metrics.OtherSys = float64(stats.OtherSys)
	metrics.PauseTotalNs = float64(stats.PauseTotalNs)
	metrics.StackInuse = float64(stats.StackInuse)
	metrics.StackSys = float64(stats.StackSys)
	metrics.Sys = float64(stats.Sys)
	metrics.TotalAlloc = float64(stats.TotalAlloc)
	metrics.PollCount++
	metrics.RandomValue = random.Float64()
}

func PollGopsutil(metrics *MetricsGopsutil) {
	v, _ := mem.VirtualMemory()
	metrics.TotalMemory = float64(v.Total)
	metrics.FreeMemory = float64(v.Free)

	cpuPercent, err := cpu.Percent(time.Second, true)
	if err != nil {
		models.Log.Error("Ошибка при сборе метрик CPU: %v")
	} else {
		metrics.CPUutilization1 = cpuPercent[0]
	}
}
