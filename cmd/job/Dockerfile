# clouderwork/job:latest
FROM alpine:latest

WORKDIR /job

COPY job /job/

ENTRYPOINT ["/job/job"]
CMD ["-config", "/workchat/config.toml"]
