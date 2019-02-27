# K8S RMQ Autoscaler

[![version](https://img.shields.io/badge/status-alpha-orange.svg)](https://github.com/XciD/k8s-rmq-autoscaler)
[![Build Status](https://travis-ci.org/XciD/k8s-rmq-autoscaler.svg?branch=master)](https://travis-ci.org/XciD/k8s-rmq-autoscaler)
[![Go Report Card](https://goreportcard.com/badge/github.com/XciD/k8s-rmq-autoscaler)](https://goreportcard.com/report/github.com/XciD/k8s-rmq-autoscaler)
[![codecov](https://codecov.io/gh/XciD/k8s-rmq-autoscaler/branch/master/graph/badge.svg)](https://codecov.io/gh/XciD/k8s-rmq-autoscaler)
[![Pulls](https://img.shields.io/docker/pulls/xcid/k8s-rmq-autoscaler.svg)](https://hub.docker.com/r/xcid/k8s-rmq-autoscaler)
[![Layers](https://shields.beevelop.com/docker/image/layers/xcid/k8s-rmq-autoscaler/latest.svg)](https://hub.docker.com/r/xcid/k8s-rmq-autoscaler)
[![Size](https://shields.beevelop.com/docker/image/image-size/xcid/k8s-rmq-autoscaler/latest.svg)](https://hub.docker.com/r/xcid/k8s-rmq-autoscaler)

K8S Autoscaler is a Pod that will run in your k8s cluster and automatically:
  * watch for your deployments that match k8s-rmq-autoscaler annotations
  * watch rabbitmq for messages in queues and consumers
  * choose to scale up / down the deployment

## K8s Configuration

Create the namespace
```
kubectl create namespace k8s-rmq-autoscaler
```

Create a secret with the RMQ Password
```
kubectl create secret generic rmq-credentials --from-literal=RMQ_PASSWORD=test -n k8s-rmq-autoscaler
```

Edit `k8s-rmq-autoscaler.yml` with your RMQ information and then you can apply the k8s configuration

See bellow for ENV configuration
```
kubectl apply -f k8s-rmq-autoscaler.yml
```

You can then watch the logs
```
kubectl logs -f pod/k8s-rmq-autoscaler -n k8s-rmq-autoscaler
```

Now we add annotations to a deployment
```
kubectl annotate deployment/your-deployment -n namespace \
    k8s-rmq-autoscaler/enable=true \ 
    k8s-rmq-autoscaler/max-workers=20 \ 
    k8s-rmq-autoscaler/min-workers=4 \ 
    k8s-rmq-autoscaler/queue=worker-queue \ 
    k8s-rmq-autoscaler/vhost=vhost
```

Now your deployment is watched by the autoscaler

## Annotations

| Config             | Mandatory | Description                                                                                                                                    |
| ------------------ | ------ | -----------------------------------------------------------------------------------------------------------------------------------------------|
| `enable`              | `true`   | enable the autoscaling on this deployment |
| `max-workers`         | `true`   | the maximum amount of worker to scale up |
| `min-workers`         | `true`   | the minimum amount of worker to scale down |
| `queue`               | `true`   | RMQ queue to watch |
| `vhost`               | `true`   | RMQ vhost where the queue can be found |
| `messages-per-worker` | `false`  | Default: `1`, set the number of message per worker |
| `cooldown-delay`      | `false`  | Default: `0s`, How long the autoscaler has to wait before another downscale operation can be performed after the current one has completed. (Duration: `5m0s`) |
| `steps`               | `false`  | Default: `1`, How many workers will be scale up/down if needed |
| `offset`              | `false`  | Default: `0`, The offset will be added if you always want more workers than message in queue. For example, if you set 1 on offset, you will always have 1 worker more than messages  |
| `override`            | `false`  | Default: `false`, Authorize the user to scale more than the max/min limits manually |
| `safe-unscale`        | `false`  | Default: true, Forbid the scaler to scale down when you still have message in queue. Used to avoid to unscale a worker that is processing a message|


## Environnement config

| Config                                               | Description                            |
| ---------------------------------------------------- | ---------------------------------------|
| `RMQ_USER`    | RMQ Username used for authentication with the RabbitMQ API                     |
| `RMQ_PASSWORD`| RMQ Password used for authentication with the RabbitMQ API                     |
| `RMQ_URL`     | RMQ URL with scheme (Ex. https://rmq:15772)                                    |
| `IN_CLUSTER`  | Boolean that indicate if your are inside the cluster or not (default `true`)     |
| `NAMESPACES`  | namespaces to watch separated by commas, (default, watching all namespaces)    |
| `TICK`        | Seconds between checks for autoscaling process (default `10`)                    |