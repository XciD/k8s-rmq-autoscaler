package main

import (
	"context"
	"os"

	"github.com/namsral/flag"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/klog"
)

func main() {
	ctx := context.Background()

	namespaces := flag.String("namespaces", "", "namespaces to watch separated by commas")
	inCluster := flag.Bool("in_cluster", true, "Boolean that indicate if your are inside the cluster or not")
	rmqURL := flag.String("rmq_url", "", "rmq Host URL")
	rmqUser := flag.String("rmq_user", "", "rmq Username used for authentication with the RabbitMQ API")
	rmqPassword := flag.String("rmq_password", "", "rmq Pasword used for authentication with the RabbitMQ API")
	loopTick := flag.Int("tick", 10, "Seconds between checks for autoscaling scale")
	flag.Parse()

	rmq, err := newRmq(*rmqURL, *rmqUser, *rmqPassword)

	if err != nil {
		klog.Error(err)
		os.Exit(128)
	}

	hub := &Autoscaler{
		rmq:    rmq,
		apps:   make(map[string]*App),
		add:    make(chan *v1beta1.Deployment),
		delete: make(chan *v1beta1.Deployment),
	}

	k8sClient, err := discover(ctx, hub, *inCluster, *namespaces)

	if err != nil {
		klog.Error(err)
		os.Exit(128)
	}

	go hub.Run(ctx, k8sClient, *loopTick)
	<-ctx.Done()
}