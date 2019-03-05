unrealircd_exporter
===================

`unrealircd_exporter` acts as a service to your unrealircd network. It will gather everything that is happening on the network and will provide the following metrics:
- `irc_users`: count of users per server on the network
- `events_total`: count of events per server and per type
- `stats_sendq`: SendQ size for each link
- `stats_sendm`: SendM for each link
- `stats_rcvem`: RcveM for each link
- `stats_send_bytes`: Sent bytes for each link
- `stats_rcve_bytes`: Received bytes for each link
- `stats_open_since_seconds`: Time in seconds since a link has been established
- `stats_idle_seconds`: Time in seconds that a link has been idle
- `servers_total`: count of linked servers

Note: you might want to remove the snomask `e` from any oper as it will be quite noisy, the exporter is requesting a `/stats L` from each server every 15 seconds.

Build
-----
```
go build
```

Exporter Configuration
----------------------
```toml
# Which IP and port to listen on (ex: localhost:6660)
Listen = "localhost:6660"

# Server to connect to (ex: icanhaz.geeknode.org:7000)
Link = "server.geeknode.org:7000"

# Name and SID of the exporter (ex: 999)
Name = "prometheus.geeknode.org"
Sid = 999

# Path to key and cert to use
Cert = "./cert.pem"
Key = "./key.pem"
```

See `examples/config.toml`

Unrealircd Configuration
------------------------
```
ulines {
       prometheus.geeknode.org;
};


link prometheus.geeknode.org
{
        incoming {
                mask *;
        };

        password "C/1QwI6imh8wdrfXnVfTDf6q+ijcTbmhmgbVXhoe+2o=" { spkifp; };
        class servers;
};
```

Need help?
----------
Either open an issue or come chat with us on `ircs://irc.geeknode.org:6697/geeknode`.
