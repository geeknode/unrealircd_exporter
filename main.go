package main

import (
	"crypto/tls"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/sorcix/irc.v2"
	"net/http"
	"os"
	"time"
)

type handler func(*Context, *irc.Encoder, *irc.Message, log.Logger)

var (
	// maps irc commands (UID, SID, ...) to handler functions
	handlers = make(map[string]handler)
)

func SendRaw(conn *tls.Conn, command string, logger log.Logger) {
	raw := fmt.Sprintf("%s\n", command)
	level.Debug(logger).Log("msg", fmt.Sprintf("--> %s", raw))
	fmt.Fprintf(conn, raw)
}

func RegisterHandler(command string, handler handler) {
	handlers[command] = handler
}

func GetLinkStats(context *Context, conn *tls.Conn, sid int, logger log.Logger) {
	for {
		for _, hostname := range context.GetServersHostnames() {
			SendRaw(conn, fmt.Sprintf(":%d000000 STATS L %s", sid, hostname), logger)
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

	// Version metric from github.com/prometheus/common
	prometheus.MustRegister(version.NewCollector("unrealircd_exporter"))

	// Register IRC Handlers
	RegisterHandler("UID", UidHandler)
	RegisterHandler("SID", SidHandler)
	RegisterHandler("SQUIT", SquitHandler)
	RegisterHandler("QUIT", QuitHandler)
	RegisterHandler("SERVER", ServerHandler)
	RegisterHandler("PING", PingHandler)
	RegisterHandler("211", StatsLHandler)
}

func main() {
	promConfig := promlog.Config{
		Level:  &promlog.AllowedLevel{},
		Format: &promlog.AllowedFormat{},
	}

	flag.AddFlags(kingpin.CommandLine, &promConfig)
	kingpin.Version(version.Print("blackbox_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(&promConfig)

	level.Info(logger).Log("msg", "Starting unrealircd_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	var err error
	conf, err := LoadConfig("config.toml")
	if err != nil {
		level.Error(logger).Log("error", err.Error())
	}

	cert, err := tls.LoadX509KeyPair(conf.Cert, conf.Key)
	if err != nil {
		level.Error(logger).Log("error", err.Error())
	}

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		level.Info(logger).Log("msg", "Listening on address", "address", conf.Listen)
		if err := http.ListenAndServe(conf.Listen, nil); err != nil {
			level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
			os.Exit(1)
		}
	}()

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}
	tlsConfig.BuildNameToCertificate()

	// TODO: check cert fingerprint
	conn, err := tls.Dial("tcp", conf.Link, tlsConfig)
	if err != nil {
		level.Error(logger).Log("error", err.Error())
		os.Exit(1)
	}

	context := NewContext()

	decoder := irc.NewDecoder(conn)
	encoder := irc.NewEncoder(conn)
	SendRaw(conn, "PASS password", logger)
	SendRaw(conn, fmt.Sprintf("PROTOCTL EAUTH=%s SID=%d ", conf.Name, conf.Sid), logger)
	SendRaw(conn, "PROTOCTL NOQUIT NICKv2 SJOIN SJ3 CLK TKLEXT TKLEXT2 NICKIP ESVID MLOCK EXTSWHOIS", logger)
	SendRaw(conn, fmt.Sprintf("SERVER %s 1 :Prometheus exporter", conf.Name), logger)
	SendRaw(conn, "EOS", logger)

	// Create our own user so to have ircop capabilities
	// UID nickname hopcount timestamp username hostname uid servicestamp umodes virthost cloakedhost ip :gecos
	SendRaw(conn, fmt.Sprintf("UID P 0 0 Prometheus 127.0.0.1 %d000000 0 +Soip * %s * :Prometheus", conf.Sid, conf.Name), logger)

	// let's collect link stats
	go GetLinkStats(context, conn, conf.Sid, logger)

	// We already have the server we're connecting to and the exporter
	serversCount.Set(1)

	for {
		message, err := decoder.Decode()
		if err != nil {
			level.Error(logger).Log("error", err.Error())
			continue
		}
		level.Debug(logger).Log("msg", fmt.Sprintf("<-- %s\n", message.String()))

		// Pass the ball to the corresponding handler
		if _, ok := handlers[message.Command]; ok {
			handlers[message.Command](context, encoder, message, logger)
		}

		// we don't want to count local events
		if message.Prefix == nil {
			continue
		}

		server, err := context.GetServer(message.Prefix.String())
		if err != nil {
			// if it's not a server then it's probably an user
			continue
		}

		eventsCount.With(prometheus.Labels{
			"event":  message.Command,
			"server": server.Hostname,
		}).Inc()
	}

	conn.Close()
}
