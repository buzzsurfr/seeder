FROM golang:1.15-alpine as builder

RUN apk add --no-cache git

# WORKDIR /go/src/github.com/buzzsurfr/seeder/
# COPY * /go/src/github.com/buzzsurfr/seeder/
ENV CGO_ENABLED 0
RUN go get github.com/buzzsurfr/seeder

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/bin/seeder /seeder
ENTRYPOINT ["/seeder"]
