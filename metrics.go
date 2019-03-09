package main

import "github.com/prometheus/client_golang/prometheus"

var (
	users = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "irc_users",
			Help: "Number of currently connected users per server.",
		},
		[]string{"server", "encryption"},
	)

	eventsCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "events_total",
			Help: "Number of events",
		},
		[]string{"event", "server"},
	)

	statsSendQ = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_sendq",
			Help: "SendQ between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsSendM = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_sendm",
			Help: "SendM between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsSendBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_send_bytes",
			Help: "Bytes sent between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsRcveM = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_rcvem",
			Help: "RcveM sent between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsRcveBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_rcve_bytes",
			Help: "Bytes received between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsOpenSince = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_open_since_seconds",
			Help: "Time in seconds since the link has been made between server from and server to",
		},
		[]string{"from", "to"},
	)

	statsIdle = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "stats_idle_seconds",
			Help: "Idle in seconds between server from and server to",
		},
		[]string{"from", "to"},
	)

	serversCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "servers_total",
			Help: "Number of currently connected servers as per the exporter point of view.",
		},
	)
)
