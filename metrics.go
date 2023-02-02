package main

import (
	"bytes"
	"fmt"
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

func HandleMetrics(conf *Conf) {
	if conf == nil {
		log.Error().Msg("empty config supplied")
		return
	}

	if len(conf.PushgatewayUrl) > 0 {
		err := PushMetrics(conf.PushgatewayUrl)
		if err != nil {
			log.Error().Msgf("could not push metrics: %v", err)
		}
	}

	err := WriteMetrics(defaultMetricsFile)
	if err != nil {
		log.Error().Msgf("could not write metrics: %v", err)
	}
}

func PushMetrics(url string) error {
	pusher := push.New(url, "conditional_reboot").Client(defaultHttpclient.StandardClient())
	return pusher.Push()
}

func WriteMetrics(path string) error {
	log.Debug().Msgf("Dumping metrics to %s", path)
	metrics, err := dumpMetrics()
	if err != nil {
		return fmt.Errorf("unable to dump metrics: %v", err)
	}

	err = os.WriteFile(path, []byte(metrics), 0644)
	if err != nil {
		return fmt.Errorf("can't write metrics to '%s': %v", path, err)
	}
	return nil
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
