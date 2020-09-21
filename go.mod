module github.com/chremoas/perms-srv

go 1.14

require (
	github.com/chremoas/services-common v1.2.5
	github.com/golang/protobuf v1.3.2
	github.com/micro/go-micro v1.9.1
	golang.org/x/net v0.0.0-20190724013045-ca1201d0de80
)

replace github.com/chremoas/perms-srv => ../perms-srv
replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
