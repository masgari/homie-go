package main

import (
	"fmt"
	"time"

	homie "github.com/masgari/homie-go/homie"
	cpu "github.com/shirou/gopsutil/cpu"
	load "github.com/shirou/gopsutil/load"
	mem "github.com/shirou/gopsutil/mem"
)

func formatByteCount(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func configureMemoryNode(device homie.Device, publisher homie.PeriodicPublisher) {
	memNode := device.NewNode("Memory", "MemoryNode")
	memNode.NewProperty("total", "integer")
	memNode.NewProperty("free", "integer")
	publisher.AddNodePublisher(memNode, func(n homie.Node) {
		totalProp := n.GetProperty("total")
		freeProp := n.GetProperty("free")
		v, _ := mem.VirtualMemory()
		totalProp.SetValue(formatByteCount(v.Total))
		freeProp.SetValue(formatByteCount(v.Available))
		totalProp.Publish()
		freeProp.Publish()
	})
}

func configureCPUNode(device homie.Device, publisher homie.PeriodicPublisher) {
	cpuNode := device.NewNode("CPU", "CPUNode")
	cpuNode.NewProperty("usage", "float")
	cpuNode.NewProperty("load", "float")
	publisher.AddNodePublisher(cpuNode, func(n homie.Node) {
		usageProp := n.GetProperty("usage")
		loadProp := n.GetProperty("load")
		cpuUsage, _ := cpu.Percent(0, false)
		usageProp.SetValue(fmt.Sprintf("%.2f", cpuUsage))
		loadStats, _ := load.Avg()
		loadProp.SetValue(fmt.Sprintf("%.2f", loadStats.Load1))
		usageProp.Publish()
		loadProp.Publish()
	})
}

func main() {
	// publish system stats every 5 seconds
	statsPublisher := homie.NewPeriodicPublisher(time.Duration(5 * time.Second))

	device := homie.NewDevice("sys-info", &homie.Config{
		Mqtt: homie.MqttConfig{
			Host:     "localhost",
			Port:     1883,
			Username: "user",
			Password: "password",
		},
		BaseTopic:           "devices/",
		StatsReportInterval: 60,
	})
	configureMemoryNode(device, statsPublisher)
	configureCPUNode(device, statsPublisher)

	homie.NewDevicePublisher(device) // report uptime every 60s
	device.Run(true)
}
