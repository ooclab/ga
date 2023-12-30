FROM golang:1.21 AS builder

ENV GO111MODULE=on GOPROXY=https://goproxy.cn
WORKDIR /go/src/github.com/ooclab/ga
COPY . .
RUN make


FROM debian:12-slim

# Install ca-certificates to update certificates
RUN apt-get update && apt-get install -y ca-certificates curl && rm -rf /var/lib/apt/lists/*

RUN mkdir -pv /etc/ga/middlewares/
COPY --from=builder /go/src/github.com/ooclab/ga/ga /usr/bin/ga
COPY --from=builder /go/src/github.com/ooclab/ga/*.so /etc/ga/middlewares/
EXPOSE 2999
CMD ["/usr/bin/ga", "serve"]
