package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-ping/ping"
)

var (
	metricNamePrefix = os.Getenv("METRIC_NAME_PREFIX")
	targetIP         = os.Getenv("TARGET_IP")
	listenPort       = os.Getenv("LISTEN_PORT")
	pingTemplate     = []byte(
		"# HELP " + metricNamePrefix + "_ms Bytes Receive Rate\n" +
			"# TYPE unifipoller_site_receive_rate_bytes gauge\n" +
			metricNamePrefix + "_ms{target_ip=\"" + targetIP + "\"} ",
	)
)

func main() {
	handler := initHandler(initPinger)

	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0%s", listenPort), handler)
	log.Fatalln(err)
}

func initPinger() *ping.Pinger {
	p, err := ping.NewPinger(targetIP)
	if err != nil {
		fmt.Println(fmt.Errorf("couldn't initialize pinger for `%s`: %w", targetIP, err))
		return nil
	}
	p.Count = 1
	if err != nil {
		fmt.Println(fmt.Errorf("couldn't resolve IP `%s`: %w", targetIP, err))
		return nil
	}
	p.Timeout = time.Millisecond * 750
	p.Debug = true
	p.SetNetwork("udp")

	return p
}

func initHandler(pinger func() *ping.Pinger) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		p := pinger()
		if p == nil {
			return
		}
		p.OnRecv = func(packet *ping.Packet) {
			fmt.Printf("IP Addr: %s receive, RTT: %v\n", packet.Addr, packet.Rtt)
			writer.Write(append(pingTemplate, []byte(fmt.Sprintf("%d\n", packet.Rtt.Milliseconds()))...))
		}

		err := p.Run()

		if err != nil {
			fmt.Println(err)
		}
	})
}
