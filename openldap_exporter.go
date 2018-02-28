package main

import (
	"crypto/tls"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/ldap.v2"
	"log"
	"net/http"
	"time"
)

var (
	syncCookie = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openldap_contextCSN",
			Help: "The first contextCSN sync cookie",
		},
		[]string{"index"},
	)
	numEntries = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "openldap_entries_num",
			Help: "Number of entries in directory",
		})
)

// Actually collect values from ldap
func ldapWorker(ldapHost, baseDN string) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	l, err := ldap.DialTLS("tcp", ldapHost, conf)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		baseDN, // The base dn to search
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",      // The filter to apply
		[]string{"contextCSN"}, // A list attributes to retrieve
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range sr.Entries {
		// log.Printf(entry.contextCSN)
		for _, csn := range entry.GetAttributeValues("contextCSN") {
			log.Printf(csn)
		}
		// log.Printf("%v\n", entry.GetAttributeValue("contextCSN"))
	}
	for {
		numEntries.Inc()
		syncCookie.WithLabelValues("001").Set(123456)
		syncCookie.WithLabelValues("002").Set(987654321)
		time.Sleep(60 * time.Second)
	}
}

func init() {
	// Register metrics
	prometheus.MustRegister(syncCookie)
	prometheus.MustRegister(numEntries)
}

func main() {
	var (
		addr        = flag.String("telemetry.addr", ":9129", "host:port for ceph exporter")
		metricsPath = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")
		ldapHost    = flag.String("ldap.host", "localhost:12345", "hostname:port of the ldap server")
		baseDN      = flag.String("base.dn", "dc=selfnet,dc=de", "'dc=selfnet,dc=de' the base DN of the directory")
	)
	flag.Parse()
	log.Printf(*addr, *metricsPath, *baseDN, *ldapHost)
	go ldapWorker(*ldapHost, *baseDN)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>OpenLDAP Exporter</title></head>
			<body>
			<h1>OpenLDAP Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	log.Printf("Starting OpenLDAP exporter on %q", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
