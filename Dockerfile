FROM quay.io/projectquay/golang:1.19 AS builder
WORKDIR /usr/src/github.com/redhat-eets/sno-tests
COPY . .
RUN go mod tidy && make test-bin

FROM centos:7
COPY --from=builder /usr/src/github.com/redhat-eets/sno-tests/bin/ptptests /usr/bin/ptptests
COPY hack/runtest-in-pod.sh /usr/bin/runtest-in-pod.sh
CMD ["/usr/bin/runtest-in-pod.sh"]
