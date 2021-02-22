package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

type controller struct {
	indexer  cache.Indexer
	queue    workqueue.Interface
	informer cache.Controller
	hub      *Autoscaler
}

func newController(queue workqueue.Interface, indexer cache.Indexer, informer cache.Controller, hub *Autoscaler) *controller {
	return &controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
		hub:      hub,
	}
}

func createClient(inCluster bool) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if inCluster {
		config, err = rest.InClusterConfig()
	} else {
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func discover(ctx context.Context, hub *Autoscaler, inCluster bool, namespacesToWatch string) (*kubernetes.Clientset, error) {
	// create the clientset
	client, err := createClient(inCluster)

	if err != nil {
		return nil, err
	}

	namespaceToWatch := getNamespacesSet(namespacesToWatch)

	namespaces, err := client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces.Items {
		klog.Infof("Scanning namespace %s", namespace.Name)

		// If we need to watch some namespace, skik others now
		if len(namespacesToWatch) > 0 {
			if _, ok := namespaceToWatch[namespace.Name]; !ok {
				klog.Infof("Skipping namespace %s", namespace.Name)
				continue
			}
		}

		listWatch := createWatch(client, namespace.Name)
		queue := workqueue.New()

		indexer, informer := cache.NewIndexerInformer(listWatch, &appsv1.Deployment{}, 0, cache.ResourceEventHandlerFuncs{
			AddFunc: func(o interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(o)
				if err == nil {
					queue.Add(key)
				}
			},
			DeleteFunc: func(o interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(o)
				if err == nil {
					queue.Add(key)
				}
			},
			UpdateFunc: func(p, o interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(o)
				if err == nil {
					queue.Add(key)
				}
			},
		}, cache.Indexers{})

		controller := newController(queue, indexer, informer, hub)

		go controller.run(ctx)
	}

	return client, nil
}

func createWatch(client *kubernetes.Clientset, namespace string) *cache.ListWatch {
	return &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return client.AppsV1().Deployments(namespace).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return client.AppsV1().Deployments(namespace).Watch(options)
		},
	}
}

func getNamespacesSet(namespacesToWatch string) map[string]bool {
	namespaceToWatchSet := make(map[string]bool)
	for _, namespacesToWatch := range strings.Split(namespacesToWatch, ",") {
		namespaceToWatchSet[namespacesToWatch] = true
	}
	return namespaceToWatchSet
}

func (c *controller) run(ctx context.Context) {
	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Info("Starting Service controller")

	go c.informer.Run(ctx.Done())

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		klog.Error("Timed out waiting for caches to sync")
		return
	}

	go wait.Until(c.runWorker, time.Second, ctx.Done())

	<-ctx.Done()
	klog.Info("Stopping Service controller")
}

func (c *controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	obj, exists, err := c.indexer.GetByKey(key.(string))
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return true
	}

	if obj == nil {
		klog.Errorf("Object is nil %s", key)
		return true
	}

	if !exists {
		fmt.Printf("Deployment %s does not exist anymore\n", key)
		c.hub.delete <- obj.(*appsv1.Deployment)
	} else {
		c.hub.add <- obj.(*appsv1.Deployment)
	}
	return true
}
