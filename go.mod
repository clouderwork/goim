module github.com/Terry-Mao/goim

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/bilibili/discovery v0.0.0-00010101000000-000000000000
	github.com/bsm/sarama-cluster v2.1.15+incompatible
	github.com/clouderwork/workchat v0.0.0-20200903071415-17b743c6236d
	github.com/couchbase/goutils v0.0.0-20191018232750-b49639060d85 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/stretchr/testify v1.5.1
	github.com/zhenjl/cityhash v0.0.0-20131128155616-cdd6a94144ab
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/Shopify/sarama.v1 v1.19.0
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
)

replace (
	cloud.google.com/go => github.com/googleapis/google-cloud-go v0.26.0
	github.com/bilibili/discovery => github.com/clouderwork/discovery v1.1.3-0.20200729091656-642d463186b9
	github.com/bsm/sarama-cluster => github.com/clouderwork/sarama-cluster v1.0.3-0.20200805073453-e867694c526e
	github.com/go-kit/kit => github.com/clouderwork/kit v0.10.2 // indirect
	golang.org/x/lint => github.com/golang/lint v0.0.0-20190227174305-5b3e6a55c961
	golang.org/x/net => github.com/golang/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/oauth2 => github.com/golang/oauth2 v0.0.0-20180821212333-d2e6202438be
	golang.org/x/sync => github.com/golang/sync v0.0.0-20180314180146-1d60e4601c6f
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180830151530-49385e6e1522
	golang.org/x/text => github.com/golang/text v0.3.0
	golang.org/x/tools => github.com/golang/tools v0.0.0-20180828015842-6cd1fcedba52
	google.golang.org/genproto => github.com/google/go-genproto v0.0.0-20200729003335-053ba62fc06f
	google.golang.org/grpc => github.com/grpc/grpc-go v1.30.0
	google.golang.org/protobuf => github.com/protocolbuffers/protobuf-go v1.24.0
)
