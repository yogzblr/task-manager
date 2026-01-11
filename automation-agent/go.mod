module github.com/automation-platform/agent

go 1.18

require (
	github.com/centrifugal/centrifuge-go v0.10.2
	github.com/prometheus/client_golang v1.18.0
	github.com/yogzblr/probe v0.0.0
	golang.org/x/sys v0.17.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/centrifugal/protocol v0.10.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/encoding v0.3.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/yogzblr/probe => ../probe
