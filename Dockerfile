FROM golang:1.14 AS builder

WORKDIR /go/src/github.com/ooclab/ga
COPY . .
RUN make


FROM debian:10
RUN mkdir -pv /etc/ga/middlewares/
COPY --from=builder /go/src/github.com/ooclab/ga/ga /usr/bin/ga
COPY --from=builder /go/src/github.com/ooclab/ga/*.so /etc/ga/middlewares/
EXPOSE 2999
CMD ["/usr/bin/ga", "serve"]
