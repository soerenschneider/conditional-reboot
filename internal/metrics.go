package internal

import (
	"bytes"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const namespace = "conditional_reboot"

var (
	ProcessStartTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "start_timestamp_seconds",
	})

	CheckerLastCheck = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "checker",
		Name:      "last_check_timestamp_seconds",
	}, []string{"name"})

	AgentState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "agent",
		Name:      "state",
	}, []string{"state", "name"})

	LastStateChange = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "agent",
		Name:      "state_change_timestamp_seconds",
	}, []string{"state", "name"})

	RebootErrors = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "invocation_errors_total",
	})
)

func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

func StartMetricsDumper(ctx context.Context, textFileDir string) {
	ticker := time.NewTicker(5 * time.Minute)

	file := path.Join(textFileDir, "conditional_reboot.prom")
	if err := WriteMetrics(file); err != nil {
		log.Error().Err(err).Msg("could not dump metrics")
	}

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
		case <-ticker.C:
			if err := WriteMetrics(file); err != nil {
				log.Error().Err(err).Msg("could not dump metrics")
			}
		}
	}
}

func WriteMetrics(path string) error {
	log.Info().Msgf("Dumping metrics to %s", path)
	metrics, err := dumpMetrics()
	if err != nil {
		return err
	}

	return os.WriteFile(path, []byte(metrics), 0644)
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
		if strings.HasPrefix(f.GetName(), namespace) {
			if err := enc.Encode(f); err != nil {
				log.Warn().Msgf("could not encode metric: %v", err)
			}
		}
	}

	return buf.String(), nil
}
