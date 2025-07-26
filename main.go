package main

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	_ "modernc.org/sqlite"
)

const (
	namespace = "email"
	subsystem = "traffic"
)

var (
	upBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "upload_bytes_total"),
		"Total bytes uploaded by each email.",
		[]string{"email", "enable"}, nil,
	)

	downBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "download_bytes_total"),
		"Total bytes downloaded by each email.",
		[]string{"email", "enable"}, nil,
	)
)

type EmailTrafficCollector struct {
	db *sql.DB
}

func (c *EmailTrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upBytesDesc
	ch <- downBytesDesc
}

func (c *EmailTrafficCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query("SELECT email, up, down, enable FROM client_traffics")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var up, down int64
		var enable bool

		if err := rows.Scan(&email, &up, &down, &enable); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		enableStr := "false"
		if enable {
			enableStr = "true"
		}

		ch <- prometheus.MustNewConstMetric(upBytesDesc, prometheus.CounterValue, float64(up), email, enableStr)
		ch <- prometheus.MustNewConstMetric(downBytesDesc, prometheus.CounterValue, float64(down), email, enableStr)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}
}

func IsSQLiteFile(filePath string) (bool, error) {
	signature := []byte("SQLite format 3\x00")
	buf := make([]byte, len(signature))
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()
	_, err = file.ReadAt(buf, 0)
	if err != nil {
		return false, err
	}
	return bytes.Equal(buf, signature), nil
}

func main() {
	var (
		metricsPath  = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		maxProcs     = kingpin.Flag("runtime.gomaxprocs", "The target number of CPUs Go will run on (GOMAXPROCS)").Envar("GOMAXPROCS").Default("1").Int()
		dbPath       = kingpin.Flag("db-path", "Path to the SQLite database").Default("/etc/x-ui/x-ui.db").String()
		toolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":9100")
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.Version(version.Print("3x-ui-traffic-exporter"))
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promslog.New(promslogConfig)

	logger.Info("Starting 3x-ui-traffic-exporter", "version", version.Info())
	logger.Info("Build context", "build_context", version.BuildContext())

	isSQLite, err := IsSQLiteFile(*dbPath)
	if err != nil {
		logger.Error("Failed to check SQLite file", "error", err)
		return
	}
	if !isSQLite {
		logger.Error("Not a valid SQLite file", "path", *dbPath)
		return
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		logger.Error("Failed to open database", "error", err)
		return
	}
	defer db.Close()

	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	collector := &EmailTrafficCollector{db: db}
	prometheus.MustRegister(collector)

	http.Handle(*metricsPath, promhttp.Handler())
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "3x-ui Traffic Exporter",
			Description: "Exports email traffic statistics from 3x-ui SQLite DB",
			Version:     version.Info(),
			Links: []web.LandingLinks{
				{
					Address: *metricsPath,
					Text:    "Metrics",
				},
			},
		}
		landingPage, err := web.NewLandingPage(landingConfig)
		if err != nil {
			logger.Error("Failed to create landing page", "error", err)
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
