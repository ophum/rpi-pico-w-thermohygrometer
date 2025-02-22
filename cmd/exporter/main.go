package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "thermohygrometer"

func newTemperatureGauge() prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "temperature",
		Help:      "temperature",
	})
}

func newHumidityGauge() prometheus.Gauge {
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "humidity",
		Help:      "humidity",
	})
}

var regs = map[string]*prometheus.Registry{}

type Col struct {
	temperatureGauge prometheus.Gauge
	humidityGauge    prometheus.Gauge
}

var collectors = map[string]Col{
	"rpi-pico-w-01": {
		temperatureGauge: newTemperatureGauge(),
		humidityGauge:    newHumidityGauge(),
	},
}

var address string

func init() {
	flag.StringVar(&address, "address", "192.168.0.1:1234", "rpi pico W ip:port")
	flag.Parse()

	for name, col := range collectors {
		reg := prometheus.NewRegistry()
		reg.MustRegister(col.temperatureGauge, col.humidityGauge)
		regs[name] = reg
	}
}
func main() {

	go collector(context.TODO())
	for name, reg := range regs {
		http.Handle("/"+name+"/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	}
	if err := http.ListenAndServe(":1235", nil); err != nil {
		log.Fatal(err)
	}
}

func collector(ctx context.Context) error {
	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		return err
	}
	defer l.Close()
	log.Println("listen :1234")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("connected")
		for {
			// 機器名 温度 湿度\r\n
			// <name> 20 40\r\n
			var n int
			b := make([]byte, 1024)
			n, err = conn.Read(b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Fatal(err)
			}
			log.Println("read", string(b[:n]))

			splited := strings.SplitN(strings.TrimSpace(string(b[:n])), " ", 3)
			if len(splited) != 3 {
				// invalid
				continue
			}
			name, tempStr, humiStr := splited[0], splited[1], splited[2]
			log.Println(name, tempStr, humiStr)

			col, ok := collectors[name]
			if !ok {
				// invalid
				log.Printf("%s is not exists", name)
				continue
			}
			temp, err := strconv.ParseFloat(string(tempStr), 64)
			if err != nil {
				continue
			}
			humi, err := strconv.ParseFloat(string(humiStr), 64)
			if err != nil {
				continue
			}

			col.temperatureGauge.Set(temp)
			col.humidityGauge.Set(humi)
		}
		conn.Close()
	}
}
