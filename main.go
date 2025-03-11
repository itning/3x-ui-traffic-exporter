package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/alecthomas/kingpin/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
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

func main() {
	app := kingpin.New("email-traffic-exporter", "A Prometheus exporter for email traffic metrics.")

	webFlags := kingpinflag.AddFlags(app, ":9100")

	dbPath := app.Flag("db-path", "Path to the SQLite database").Default("/etc/x-ui/x-ui.db").String()

	kingpin.MustParse(app.Parse(nil))

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	collector := &EmailTrafficCollector{db: db}
	prometheus.MustRegister(collector)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Handler: mux,
	}

	logger := slog.New(slog.NewTextHandler(log.Writer(), nil))

	if err := web.ListenAndServe(server, webFlags, logger); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
