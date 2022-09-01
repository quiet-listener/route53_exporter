package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	route53 "route53_exporter/route53"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	csi_env                      = os.Getenv("ACCOUNT")
	resourcerecordsetcount_gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "aws_route53",
			Subsystem:   "hostedzone",
			Name:        "resourcerecordsetcount",
			Help:        "Route53 RecordSet Count",
			ConstLabels: prometheus.Labels{"account": csi_env},
		}, []string{"route53_hostname", "route53_hostedzoneid", "route53_privateZone"})
	resourcerecordsetlimit_gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace:   "aws_route53",
			Subsystem:   "hostedzone",
			Name:        "resourcerecordsetlimit",
			Help:        "Route53 RecordSet Limit",
			ConstLabels: prometheus.Labels{"account": csi_env},
		}, []string{"route53_hostname", "route53_hostedzoneid", "route53_privateZone"})
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	histogramVec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "prom_request_time",
		Help: "Time it has taken to retrieve the metrics",
		//ConstLabels: prometheus.Labels{"env": csi_env, "cluster": csi_cluster},
	}, []string{"time"})
	router.HandleFunc("/", homeLink)
	router.Handle("/metrics", newHandlerWithHistogram(promhttp.Handler(), histogramVec))
	prometheus.Register(histogramVec)
	prometheus.MustRegister(resourcerecordsetcount_gauge)
	prometheus.MustRegister(resourcerecordsetlimit_gauge)
	go func() {
		for {
			route53dataList, err := route53.Route53Metrics()
			if err != nil {
				log.Print(err)
				resourcerecordsetcount_gauge.WithLabelValues("fetcherror", "fetcherror", "fetcherror").Add(1)
				resourcerecordsetlimit_gauge.WithLabelValues("fetcherror", "fetcherror", "fetcherror").Add(1)
			} else {
				for _, route53Data := range route53dataList {
					resourcerecordsetcount_gauge.WithLabelValues(route53Data.Name, route53Data.Hostedzoneid, route53Data.PrivateZone).Set(route53Data.Count)
					resourcerecordsetlimit_gauge.WithLabelValues(route53Data.Name, route53Data.Hostedzoneid, route53Data.PrivateZone).Set(route53Data.Limit)
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()
	log.Fatal(http.ListenAndServe(":8090", router))
}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func newHandlerWithHistogram(handler http.Handler, histogram *prometheus.HistogramVec) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		status := http.StatusOK

		defer func() {
			histogram.WithLabelValues(fmt.Sprintf("%d", status)).Observe(time.Since(start).Seconds())
		}()

		if req.Method == http.MethodGet {
			handler.ServeHTTP(w, req)
			return
		}
		status = http.StatusBadRequest

		w.WriteHeader(status)
	})
}
