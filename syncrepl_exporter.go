package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jinzhu/configor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/ldap.v2"
)

var (
	openldapUp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "openldap_up",
			Help: "Value whether a connection to OpenLDAP has been successful",
		})

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

var Config = struct {
	Ldap struct {
		Host     string `default:"localhost"`
		Port     string `default:"636"`
		Basedn   string `default:"dc=example,dc=org"`
		StartTLS bool   `default:"false"`
		Bind     bool   `default:"false"`
		Bindcn   string `default:""`
		Bindpass string `default:""`
	}
}{}

func ymdToUnix(contextCSN string) (timestamp int64, label string) {
	// This is a totally crude approach to set a well known base time to parse another date later
	format := "20060102150405"
	ymd := strings.Split(contextCSN, ".")[0]
	time, _ := time.Parse(format, ymd)
	label = strings.Split(contextCSN, "#")[2]
	return time.Unix(), label
}

// Actually collect values from ldap
func csnWorker() {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	var l *ldap.Conn
	var err error

	if Config.Ldap.StartTLS {
		// Connect to host
		l, err = ldap.Dial("tcp", Config.Ldap.Host + ":" + Config.Ldap.Port)
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()

		// Reconnect with TLS
		err = l.StartTLS(conf)
	} else {
		l, err = ldap.DialTLS("tcp", Config.Ldap.Host + ":" + Config.Ldap.Port, conf)
	}

	// Bind
	if Config.Ldap.Bind {
		err = l.Bind(Config.Ldap.Bindcn, Config.Ldap.Bindpass)
	}

	if err != nil {
		openldapUp.Set(0)
		log.Fatal(err)
	} else {
		defer l.Close()

		searchRequest := ldap.NewSearchRequest(
			Config.Ldap.Basedn, // The base dn to search
			ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=*)",      // The filter to apply
			[]string{"contextCSN"}, // A list attributes to retrieve
			nil,
		)

		sr, err := l.Search(searchRequest)
		if err != nil {
			openldapUp.Set(0)
			log.Println(err)
		} else {
			openldapUp.Set(1)
			for _, entry := range sr.Entries {
				for _, csn := range entry.GetAttributeValues("contextCSN") {
					epoch, label := ymdToUnix(csn)
					syncCookie.WithLabelValues(label).Set(float64(epoch))
				}
			}
		}
		searchRequest = ldap.NewSearchRequest(
			Config.Ldap.Basedn, // The base dn to search
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(objectClass=*)", // The filter to apply
			[]string{"dn"},    // A list attributes to retrieve
			nil,
		)

		sr, err = l.Search(searchRequest)
		if err != nil {
			openldapUp.Set(0)
			log.Println(err)
		} else {
			openldapUp.Set(1)
			numEntries.Set(float64(len(sr.Entries)))
		}

	}
}

func ldapWorker() {
	for {
		csnWorker()
		time.Sleep(60 * time.Second)
	}
}

func init() {
	// Register metrics
	prometheus.MustRegister(syncCookie)
	prometheus.MustRegister(numEntries)
	prometheus.MustRegister(openldapUp)
}

func main() {
	var (
		addr        = flag.String("telemetry.addr", ":9328", "host:port for syncrepl exporter")
		metricsPath = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")
		configFile  = flag.String("config.file", "config.yaml", "bind cn and password")
	)

	flag.Parse()
	log.Printf(*addr, *metricsPath, *configFile)

	configor.Load(&Config, *configFile)

	log.Printf("config: %#v", Config)

	go ldapWorker()

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
