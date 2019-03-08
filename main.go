package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/sorcix/irc.v2"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type handler func(*irc.Encoder, *irc.Message)

var (
	// holds current list of server, map SID to hostname
	servers = make(map[string]string)

	// maps irc commands (UID, SID, ...) to handler functions
	handlers = make(map[string]handler)

	// configuration items
	conf *Config
)

func SendRaw(conn *tls.Conn, command string) {
	fmt.Fprintf(conn, fmt.Sprintf("%s\r\n", command))
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(a, e) {
			return true
		}
	}
	return false
}

func ResolveServer(prefix string) (server string) {
	// SID
	if res, _ := regexp.MatchString("^[0-9]{3}$", prefix); res {
		return servers[prefix]
	}

	// UID
	if res, _ := regexp.MatchString("^[0-9]{3}", prefix); res {
		return servers[prefix[:3]]
	}

	// hostname
	if res, _ := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.){2,}([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9]){2,}$`, prefix); res {
		return prefix
	}

	// we don't know what it is
	// most probably it's an user action but we don't have a user <> server map... yet
	return "unknown"
}

func FindSidByHostname(hostname string) (sid string, err error) {
	for key, value := range servers {
		if value == hostname {
			sid = key
			return
		}
	}
	return "", errors.New(fmt.Sprintf("Couldn't find a server named %s", hostname))
}

func RegisterHandler(command string, handler handler) {
	handlers[command] = handler
}

func GetLinkStats(conn *tls.Conn) {
	for {
		for _, hostname := range servers {
			SendRaw(conn, fmt.Sprintf(":%s000000 STATS L %s", conf.Sid, hostname))
		}

		time.Sleep(15 * time.Second)
	}
}

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(eventsCount)
	prometheus.MustRegister(serversCount)
	prometheus.MustRegister(statsIdle)
	prometheus.MustRegister(statsOpenSince)
	prometheus.MustRegister(statsRcveBytes)
	prometheus.MustRegister(statsRcveM)
	prometheus.MustRegister(statsSendBytes)
	prometheus.MustRegister(statsSendM)
	prometheus.MustRegister(statsSendQ)
	prometheus.MustRegister(users)

	// Register IRC Handlers
	RegisterHandler("UID", UidHandler)
	RegisterHandler("SID", SidHandler)
	RegisterHandler("SQUIT", SquitHandler)
	RegisterHandler("QUIT", QuitHandler)
	RegisterHandler("PROTOCTL", ProtoctlHandler)
	RegisterHandler("PING", PingHandler)
	RegisterHandler("211", StatsLHandler)
}

func main() {
	var err error
	conf, err = LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err.Error())
	}

	cert, err := tls.LoadX509KeyPair(conf.Cert, conf.Key)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Fatal(http.ListenAndServe(conf.Listen, nil))
	}()

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate()

	// TODO: check cert fingerprint
	conn, err := tls.Dial("tcp", conf.Link, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}

	decoder := irc.NewDecoder(conn)
	encoder := irc.NewEncoder(conn)
	SendRaw(conn, "PASS password")
	SendRaw(conn, fmt.Sprintf("PROTOCTL EAUTH=%s SID=%d ", conf.Name, conf.Sid))
	SendRaw(conn, "PROTOCTL NOQUIT NICKv2 SJOIN SJ3 CLK TKLEXT TKLEXT2 NICKIP ESVID MLOCK EXTSWHOIS")
	SendRaw(conn, fmt.Sprintf("SERVER %s 345 :Prometheus exporter", conf.Name))
	SendRaw(conn, "EOS")

	// Create our own user so to have ircop capabilities
	// UID nickname hopcount timestamp username hostname uid servicestamp umodes virthost cloakedhost ip :gecos
	SendRaw(conn, fmt.Sprintf("UID P 0 0 Prometheus 127.0.0.1 %d000000 0 +Soip * %s * :Prometheus", conf.Sid, conf.Name))

	// let's collect link stats
	go GetLinkStats(conn)

	// We already have the server we're connecting to and the exporter
	serversCount.Set(2)

	for {
		message, err := decoder.Decode()
		if err != nil {
			log.Fatal(err)
		}

		// Pass the ball to the corresponding handler
		if _, ok := handlers[message.Command]; ok {
			handlers[message.Command](encoder, message)
		}

		// we don't want to count local events
		if message.Prefix == nil {
			continue
		}

		eventsCount.With(prometheus.Labels{
			"event":  message.Command,
			"server": ResolveServer(message.Prefix.String()),
		}).Inc()
	}

	conn.Close()
}
