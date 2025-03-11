package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/prometheus/common/version"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	_ "modernc.org/sqlite"
)

type EmailTrafficCollector struct {
	db *sql.DB
}

func (c *EmailTrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("email_upload_bytes_total", "Total bytes uploaded by each email.", []string{"email", "enable"}, nil)
	ch <- prometheus.NewDesc("email_download_bytes_total", "Total bytes downloaded by each email.", []string{"email", "enable"}, nil)
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
		var enable int
		if err := rows.Scan(&email, &up, &down, &enable); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		enableLabel := fmt.Sprintf("%d", enable)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("email_upload_bytes_total", "Total bytes uploaded by each email.", []string{"email", "enable"}, nil),
			prometheus.CounterValue,
			float64(up),
			email,
			enableLabel,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("email_download_bytes_total", "Total bytes downloaded by each email.", []string{"email", "enable"}, nil),
			prometheus.CounterValue,
			float64(down),
			email,
			enableLabel,
		)
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
		logger.Error(fmt.Sprintf("Error: %v", err))
		return
	}
	if !isSQLite {
		logger.Error(fmt.Sprintf("It doesn't look like a sqlite file: %s", *dbPath))
		return
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to open database: %v", err))
	}
	defer db.Close()

	runtime.GOMAXPROCS(*maxProcs)
	logger.Debug("Go MAXPROCS", "procs", runtime.GOMAXPROCS(0))

	collector := &EmailTrafficCollector{db: db}
	prometheus.MustRegister(collector)

	http.Handle(*metricsPath, promhttp.Handler())
	if *metricsPath != "/" {
		landingConfig := web.LandingConfig{
			Name:        "3x-ui-traffic-exporter",
			Description: "3x-ui-traffic-exporter",
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
			logger.Error(err.Error())
			os.Exit(1)
		}
		http.Handle("/", landingPage)
	}

	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
