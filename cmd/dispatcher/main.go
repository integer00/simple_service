package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/integer00/simple_service/pkg/objects"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceName = "TestDispatcher"
	servicePort = 8081

	workersNum = 70
	// 	sudo ulimit -n 1024
	// 	sudo sysctl -w kern.ipc.somaxconn=1024

	dispatcherNumRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "dispatcher_requests",
		Help: "Dispatcher requests num",
	}, []string{"status"})

	dispatcherLatency = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "dispatcher_request_latency",
		Help: "Latency for requests",
	})
)

func doRequest() {
	start := time.Now()
	resp, err := http.Get("http://192.168.88.253:8080/health")
	// resp, err := http.Get("http://192.168.88.253:8080/bufferedHealth")
	time := time.Since(start).Seconds()
	if err != nil {
		dispatcherNumRequests.With(prometheus.Labels{"status": "500"}).Inc()
		dispatcherLatency.Set(0)

		fmt.Printf("%s\n", err.Error())
		return
	}
	defer resp.Body.Close()
	dispatcherNumRequests.With(prometheus.Labels{"status": resp.Status}).Inc()
	dispatcherLatency.Set(time)
	fmt.Printf("status - %d took - %f\n", resp.StatusCode, time)
}
func main() {

	service := objects.Service{
		Name:    serviceName,
		Address: "192.168.88.253",
		Port:    servicePort,
		ID:      serviceName,
	}

	client := service.CreateAndRegisterConsulService()

	// httpClient := http.Client{
	// 	Transport: &http.Transport{
	// 		MaxConnsPerHost: 800,
	// 	},
	// }

	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c

		done <- true
		ticker.Stop()

		client.Agent().ServiceDeregister(service.Name)

		os.Exit(0)
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				for i := 0; i < workersNum; i++ {
					go func() {
						doRequest()
					}()

				}
			}
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("starting service at :8081")
	http.ListenAndServe(":8081", nil)
}
