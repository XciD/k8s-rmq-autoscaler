FROM golang:1.11-stretch as builder

ENV PROJECT github.com/XciD/k8s-rmq-autoscaler
ENV GO111MODULE on
WORKDIR /go/src/$PROJECT

COPY go.mod /go/src/$PROJECT
COPY go.sum /go/src/$PROJECT

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /k8s-rmq-autoscaler .

FROM alpine as release
COPY --from=builder /k8s-rmq-autoscaler /k8s-rmq-autoscaler

ENTRYPOINT ["/k8s-rmq-autoscaler"]