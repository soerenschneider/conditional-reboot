package main

import (
	"bytes"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

const (
	metricsNamespace = "conditional_reboot"
)

var (
	MetricStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "condition_status",
		Help:      "Status of conditions",
	}, []string{"name", "state"})

	QueryStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "query_status_success_bool",
		Help:      "Status of a single query",
	}, []string{"name"})

	MetricSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "success_bool",
		Help:      "Whether the tool ran successfully",
	})

	MetricRebootNeeded = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "needs_reboot_bool",
		Help:      "Reflects whether a reboot is needed",
	})

	MetricStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "start_time_seconds",
		Help:      "Timestamp in seconds when conditional-reboot was started",
	})

	MetricConditions = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "conditions_total",
		Help:      "Amount of configured conditions",
	})
)

func HandleMetrics(conf *Conf) error {
	if conf == nil {
		return errors.New("empty config supplied")
	}

	if len(conf.PushgatewayUrl) > 0 {
		PushMetrics(conf.PushgatewayUrl)
	}

	return WriteMetrics(defaultMetricsFile)
}

func PushMetrics(url string) error {
	pusher := push.New(url, "conditional_reboot").Client(defaultHttpclient.StandardClient())
	return pusher.Push()
}

func WriteMetrics(path string) error {
	log.Info().Msgf("Dumping metrics to %s", path)
	metrics, err := dumpMetrics()
	if err != nil {
		log.Info().Msgf("Error dumping metrics: %v", err)
		return err
	}

	err = os.WriteFile(path, []byte(metrics), 0644)
	if err != nil {
		log.Error().Msgf("Error writing metrics to '%s': %v", path, err)
	}
	return err
}

func dumpMetrics() (string, error) {
	var buf = &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return "", err
	}

	for _, f := range families {
		// Writing these metrics will cause a duplication error with other tools writing the same metrics
		if strings.HasPrefix(f.GetName(), metricsNamespace) {
			if err := enc.Encode(f); err != nil {
				log.Info().Msgf("could not encode metric: %s", err.Error())
			}
		}
	}

	return buf.String(), nil
}
