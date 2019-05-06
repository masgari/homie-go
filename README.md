# homie-go

Homie (https://homieiot.github.io/specification/) MQTT convention in golang.


## Example
```go
package main

import (
	"time"

	homie "github.com/masgari/homie-go/homie"
)

func main() {
	device := homie.NewDevice("homie-go", &homie.Config{
		Mqtt: homie.MqttConfig{
			Host:     "localhost",
			Port:     1883,
			Username: "user",
			Password: "password",
		},
		BaseTopic:           "devices/",
		StatsReportInterval: 60,
	})

	timeNode := device.NewNode("time", "TimeNode")
	timeNode.NewProperty("currentTime", "time")

	publisher := homie.NewPeriodicPublisher(1 * time.Second)
	publisher.AddNodePublisher(timeNode, func(n homie.Node) {
		n.GetProperty("currentTime").
			SetValue(time.Now().String())
		n.Publish()
	})

	device.Run(true) // block forever
}
```

More examples:
* Basic: [examples/basic/main.go](examples/basic/main.go) with handler to change interval
* SysInfo: [examples/sysinfo/main.go](examples/sysinfo/main.go) report CPU and memory usage periodically
 