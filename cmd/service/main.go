package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/integer00/simple_service/pkg/objects"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

var (
	serviceName  = "TestService"
	servicePort  = 8080
	workersNum   = 1000
	workersBurst = 10
	workTime     = 300

	serviceLatency = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "request_latency",
		Help: "Latency for requests",
	})
	serviceRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: serviceName + "_requests",
		Help: "Number of requests",
	})
)

func doWork() {
	time.Sleep(time.Duration(workTime) * time.Millisecond)
}

func rateLimit(rps, burst int, wait time.Duration, h http.HandlerFunc) http.HandlerFunc {
	l := rate.NewLimiter(rate.Limit(rps), burst)

	return func(w http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), wait)
		defer cancel()

		if err := l.Wait(ctx); err != nil {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		h(w, req)
		serviceRequests.Inc()
	}
}

func bufferedQueue(h http.HandlerFunc) http.HandlerFunc {
	queue := make(chan struct{}, workersNum)

	return func(w http.ResponseWriter, req *http.Request) {
		queue <- struct{}{}
		defer func() { <-queue }()

		h(w, req)
	}
}

func healthHandler(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// rand.Seed(time.Now().UnixNano())
	// n := rand.Intn(5000)

	log.Printf("[%s] - %s %s\n", req.RemoteAddr, req.Method, req.URL)

	w.Header().Add("Content-Type", "application/json")

	resp, err := json.Marshal(map[string]string{
		"status": "ok",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fprintf: %v\n", err)
	}

	// do work
	doWork()
	w.Write(resp)
	serviceLatency.Set(time.Since(start).Seconds())

}

func main() {

	service := objects.Service{
		Name:    serviceName,
		Address: "192.168.88.253",
		Port:    servicePort,
		ID:      serviceName,
	}

	client := service.CreateAndRegisterConsulService()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer := http.Server{
		Addr: ":8080",
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		//deregister client while exit
		client.Agent().ServiceDeregister(service.Name)
		//shutdown server
		httpServer.Shutdown(ctx)

		os.Exit(0)
	}()

	http.HandleFunc("/health", rateLimit(workersNum, workersBurst, 2*time.Second, healthHandler))
	http.HandleFunc("/bufferedHealth", bufferedQueue(healthHandler))

	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("starting service at :8080")
	httpServer.ListenAndServe()
}
