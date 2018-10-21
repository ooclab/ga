FROM golang:1.11 AS builder

WORKDIR /go/src/github.com/ooclab/ga
COPY . .

RUN CGO_ENABLED=0 \
    go build -o ga \
    -a -installsuffix cgo \
    -ldflags "-s -X main.buildstamp=`date '+%Y-%m-%d_%H:%M:%S_%z'` -X main.githash=`git rev-parse HEAD`"


FROM scratch
COPY --from=builder /go/src/github.com/ooclab/ga/ga /usr/bin/
EXPOSE 2999
CMD ["/usr/bin/ga"]
