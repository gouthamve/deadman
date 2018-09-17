# Deadman's Snitch.

A dead simple snitch for the Prometheus Alertmanager. An external service is needed
with deadman's snitch functionality to make sure the alerting pipeline is working.
I could not find anything simple to use, so I made one.

To install: `go get github.com/gouthamve/deadman`

To run: `./deadman -h`

To build Docker Image : `make docker`

Add this rule to the Prometheus server to continuously generate alerts:
```
- alert: DeadManBoy
  expr: vector(1)
  labels:
    severity: deadman
  annotations:
    description: This is a DeadMansSwitch meant to ensure that the entire Alerting
      pipeline is functional.

```

And in the Alertmanager cluster, add a route to send webhook notifications to our
deployed deadman process.

```
- receiver: deadmans-switch
  group_wait: 0s
  group_interval: 0s
  repeat_interval: 15s
  match:
    severity: deadman

- name: deadmans-switch
  webhook_configs:
  - url: http://deadman-ip:9095
```

Run an alertmanager co-located with the deadman process which will notify you for
all the alerts it receives.
