module github.com/batchcorp/gophercon2021/svc

go 1.16

// Necessary hack for etcd client
replace google.golang.org/grpc => google.golang.org/grpc v1.26.0

require (
	github.com/DataDog/datadog-go/v5 v5.0.1
	github.com/batchcorp/rabbit v0.1.16
	github.com/coreos/etcd v3.3.27+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/relistan/go-director v0.0.0-20200406104025-dbbf5d95248d // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/streadway/amqp v1.0.0
	go.etcd.io/etcd v3.3.27+incompatible
	go.uber.org/zap v1.19.1 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
