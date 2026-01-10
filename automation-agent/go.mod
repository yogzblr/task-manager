module github.com/automation-platform/agent

go 1.21

require (
	github.com/centrifugal/centrifuge-go v0.10.2
	github.com/prometheus/client_golang v1.18.0
	github.com/yogzblr/probe v0.0.0
	golang.org/x/sys v0.15.0
)

replace github.com/yogzblr/probe => ../probe
