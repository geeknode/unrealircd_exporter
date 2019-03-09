package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/sorcix/irc.v2"
	"strconv"
	"strings"
)

// Example:
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P SendQ SendM SendBytes RcveM RcveBytes Open_since Idle
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P wirefull.geeknode.org[@11.22.33.44.45900][s] 0 241472 15832 12009 563 87657 0
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P united.geeknode.org[@22.33.44.55.41183][s] 0 313812 20219 86695 5187 124411 0
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P services.geeknode.org[@33.44.55.66.52577][] 0 886999 46386 38403 2604 284301 0
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P fdn.geeknode.org[@44.55.66.77.60589][s] 0 564462 36871 324176 20894 290496 0
// 2019/03/04 18:24:43 :hivane.geeknode.org 211 P icanhaz.geeknode.org[@::ffff:55.66.77.88.0][s] 0 872711 56582 3833 261 290504 0
func StatsLHandler(_ *Context, _ *irc.Encoder, message *irc.Message, logger log.Logger) {
	if message.Params[1] == "SendQ" {
		// We skip the header
		return
	}

	hostname := message.Prefix.Name
	to := strings.Split(message.Params[1], "[")[0]

	sendQ, err := strconv.ParseFloat(message.Params[2], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (SendQ) for %s to float", message.Params[2], hostname))
		return
	}
	statsSendQ.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(sendQ)

	sendM, err := strconv.ParseFloat(message.Params[3], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (SendM) for %s to float", message.Params[3], hostname))
		return
	}
	statsSendM.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(sendM)

	sendBytes, err := strconv.ParseFloat(message.Params[4], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (SendBytes) for %s to float", message.Params[4], hostname))
		return
	}
	statsSendBytes.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(sendBytes)

	rcveM, err := strconv.ParseFloat(message.Params[5], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (RcveM) for %s to float", message.Params[5], hostname))
		return
	}
	statsRcveM.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(rcveM)

	rcveBytes, err := strconv.ParseFloat(message.Params[6], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (RcveBytes) for %s to float", message.Params[6], hostname))
		return
	}
	statsRcveBytes.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(rcveBytes)

	openSince, err := strconv.ParseFloat(message.Params[7], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (Open_Since) for %s to float", message.Params[7], hostname))
		return
	}
	statsOpenSince.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(openSince)

	idle, err := strconv.ParseFloat(message.Params[8], 64)
	if err != nil {
		level.Error(logger).Log("error", fmt.Sprintf("Can't convert %s (Idle) for %s to float", message.Params[8], hostname))
		return
	}
	statsIdle.With(prometheus.Labels{
		"from": hostname,
		"to":   to,
	}).Set(idle)
}

// SERVER icanhaz.geeknode.org 1 :U4017-Fhin6OoEM-042 GeekNode Server
func ServerHandler(context *Context, _ *irc.Encoder, message *irc.Message, _ log.Logger) {
	hostname := message.Params[0]
	sid := strings.Split(
		strings.Split(message.Params[2], "-")[2],
		" ",
	)[0]

	context.AddServer(hostname, sid)

	serversCount.Inc()

	// initiate the user count, it'll increase with every UidHandler call
	users.With(prometheus.Labels{
		"server":     hostname,
		"encryption": "plaintext",
	}).Set(0)

	users.With(prometheus.Labels{
		"server":     hostname,
		"encryption": "tls",
	}).Set(0)
}

// Example
// PING icanhaz.geeknode.org
func PingHandler(_ *Context, encoder *irc.Encoder, message *irc.Message, logger log.Logger) {
	response := irc.Message{
		Prefix:  nil,
		Command: irc.PONG,
		Params: []string{
			message.Params[0],
		},
	}

	level.Debug(logger).Log("msg", "--> %s", response.String())

	err := encoder.Encode(&response)
	if err != nil {
		level.Error(logger).Log("error", err.Error())
	}
}

func SidHandler(context *Context, _ *irc.Encoder, message *irc.Message, _ log.Logger) {
	hostname := message.Params[0]
	sid := message.Params[2]

	context.AddServer(hostname, sid)
	serversCount.Inc()

	// initiate the user count, it'll increase with every UidHandler call
	users.With(prometheus.Labels{
		"server":     hostname,
		"encryption": "plaintext",
	}).Set(0)

	users.With(prometheus.Labels{
		"server":     hostname,
		"encryption": "tls",
	}).Set(0)
}

func SquitHandler(context *Context, _ *irc.Encoder, message *irc.Message, logger log.Logger) {
	hostname := message.Params[0]
	context.RemoveServer(hostname)

	serversCount.Dec()

	// remove the user count metric as the server doesn't exist anymore
	users.DeleteLabelValues(hostname, "plaintext")
	users.DeleteLabelValues(hostname, "tls")
}

func QuitHandler(context *Context, _ *irc.Encoder, message *irc.Message, logger log.Logger) {
	user, err := context.GetUser(message.Prefix.String())
	if err != nil {
		level.Error(logger).Log("error", err.Error())
		return
	}

	if user.Encryption {
		users.With(prometheus.Labels{
			"server":     user.Server.Hostname,
			"encryption": "tls",
		}).Dec()
	} else {
		users.With(prometheus.Labels{
			"server":     user.Server.Hostname,
			"encryption": "plaintext",
		}).Dec()
	}
}

// UID nickname hopcount timestamp username hostname uid servicestamp umodes virthost cloakedhost ip :gecos
func UidHandler(context *Context, _ *irc.Encoder, message *irc.Message, logger log.Logger) {
	server, err := context.GetServer(message.Prefix.String())
	if err != nil {
		level.Error(logger).Log("error", err.Error())
		return
	}

	encryption := "plaintext"
	if strings.Contains(message.Params[7], "z") {
		encryption = "tls"
		context.AddUser(message.Params[0], message.Params[5], true, server.Hostname)
	} else {
		context.AddUser(message.Params[0], message.Params[5], false, server.Hostname)
	}

	users.With(prometheus.Labels{
		"server":     server.Hostname,
		"encryption": encryption,
	}).Inc()
}
