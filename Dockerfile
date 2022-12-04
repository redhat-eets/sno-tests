FROM golang:1.19 AS builder
WORKDIR /usr/src/github.com/redhat-eets/sno-tests
COPY . .
RUN go mod tidy && make test-bin

FROM centos:7
COPY --from=builder /usr/src/github.com/redhat-eets/sno-tests/bin/ptptests /usr/bin/ptptests
CMD ["/usr/bin/ptptests", "-test.v"]
