# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf target/
	mkdir target/
	cp cmd/comet/comet-example.toml target/comet.toml
	cp cmd/logic/logic-example.toml target/logic.toml
	cp cmd/job/job-example.toml target/job.toml
	$(GOBUILD) -o target/comet cmd/comet/main.go
	$(GOBUILD) -o target/logic cmd/logic/main.go
	$(GOBUILD) -o target/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf target/

run:
	nohup target/logic -conf=target/logic.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 2>&1 > target/logic.log &
	nohup target/comet -conf=target/comet.toml -region=sh -zone=sh001 -deploy.env=dev -weight=10 -addrs=127.0.0.1 -debug=true 2>&1 > target/comet.log &
	nohup target/job -conf=target/job.toml -region=sh -zone=sh001 -deploy.env=dev 2>&1 > target/job.log &

stop:
	pkill -f target/logic
	pkill -f target/job
	pkill -f target/comet

.PYONY: protobuf
protobuf:
	@mkdir -p target
	protoc \
		-I . \
		--go_out=plugins=grpc:./target \
		api/logic/grpc/api.proto
	protoc \
		-I . \
		--go_out=plugins=grpc:./target \
		api/comet/grpc/api.proto
	mv ./target/github.com/Terry-Mao/goim/api/comet/grpc/api.pb.go api/comet/grpc/
	mv ./target/github.com/Terry-Mao/goim/api/logic/grpc/api.pb.go api/logic/grpc/

# 定义一个制作镜像的函数
define docker_build_image
	@# 第一个参数是程序名称
	@# 第二个参数是镜像的tag
	@# 第三个参数Dockerfile文件路径
	@# 第四个参数Docker制作镜像的路径
	docker build ${BUILD_ARGS} -t clouderwork/${1}:${2} -f ${3} ${4}
endef

# 编译可执行程序 打包基础镜像
.PHONY: build-exe build-base-images
build-exe:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/comet/comet cmd/comet/main.go
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/job/job cmd/job/main.go

build-base-images: build-exe
	$(call docker_build_image,comet,latest,./cmd/comet/Dockerfile,./cmd/comet)
	$(call docker_build_image,job,latest,./cmd/job/Dockerfile,./cmd/job)

