module github.com/chremoas/perms-srv

go 1.14

require (
	github.com/Masterminds/squirrel v1.5.0
	github.com/chremoas/services-common v1.3.2
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.18+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/gorilla/websocket v1.4.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.9.5 // indirect
	github.com/jmoiron/sqlx v1.3.1
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/lib/pq v1.10.0
	github.com/lucas-clemente/quic-go v0.14.1 // indirect
	github.com/micro/go-micro v1.9.1
	github.com/miekg/dns v1.1.27 // indirect
	github.com/nats-io/nats-server/v2 v2.1.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/spf13/viper v1.4.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200122045848-3419fae592fc // indirect
	go.etcd.io/bbolt v1.3.4 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f // indirect
	golang.org/x/net v0.0.0-20200222125558-5a598a2470a0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	golang.org/x/tools v0.0.0-20191216173652-a0e659d51361 // indirect
	google.golang.org/genproto v0.0.0-20191216164720-4f79533eabd1 // indirect
	google.golang.org/grpc v1.26.0 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace github.com/chremoas/perms-srv => ../perms-srv

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
