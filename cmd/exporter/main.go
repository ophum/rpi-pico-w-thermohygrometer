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

var (
	tempperatureGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "temperature",
		Help:      "temperature",
	})
	humidityGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "humidity",
		Help:      "humidity",
	})
)

var address string

func init() {
	flag.StringVar(&address, "address", "192.168.0.1:1234", "rpi pico W ip:port")
	flag.Parse()
}
func main() {

	prometheus.MustRegister(tempperatureGauge, humidityGauge)
	go collector(context.TODO())
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":1235", nil); err != nil {
		log.Fatal(err)
	}
}

func collector(ctx context.Context) error {
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("connected")
		for {
			// 温度 湿度\r\n
			// 20 40\r\n
			// 12345 6 7
			var n int
			b := make([]byte, 7)
			n, err = conn.Read(b)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				log.Fatal(err)
			}
			log.Println("read", string(b))

			if n != 7 {
				log.Println("n != 7")
				// invalid
				continue
			}

			tempStr, humiStr, ok := strings.Cut(strings.TrimSpace(string(b)), " ")
			if !ok {
				// invalid
				continue
			}
			log.Println(tempStr, humiStr)

			temp, err := strconv.ParseFloat(string(tempStr), 64)
			if err != nil {
				continue
			}
			humi, err := strconv.ParseFloat(string(humiStr), 64)
			if err != nil {
				continue
			}

			tempperatureGauge.Set(temp)
			humidityGauge.Set(humi)
		}
		conn.Close()
	}
}
