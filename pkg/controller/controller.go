package controller

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/viveksahu26/kubewatchpods/config"
	"github.com/viveksahu26/kubewatchpods/pkg/utils"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// Event indicate the informerEvent
type Event struct {
	key          string
	eventType    string
	namespace    string
	resourceType string
}

// Controller contains 2 Most important things: informer, workqueue, except all these client, and logger
// less important

type Controller struct {
	// Informer: is the controller SharedInformer.
	informer cache.SharedIndexInformer

	// Queue: is the controller Workqueue.
	queue workqueue.RateLimitingInterface

	// Logger: manages the controller logs.
	logger *logrus.Entry

	// clientset: helps the controller interact with Kubernetes API server.
	clientset kubernetes.Interface

	// eventHandler(may be optional):  holds communication to Slack which can extend to other channels.
	eventHandler handlers.Handler
}

func Start(conf *config.Config, eventHandler handlers.Handler) {
	var kubeClient kubernetes.Interface
	// if _, err := rest.InClusterConfig(); err != nil {
	// 	kubeClient = utils.GetClientOutOfCluster()
	// } else {
	kubeClient = utils.GetClient()

	// User Configured Events
	if conf.Resource.Pod {
		// create Informer: for watching pods
		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
					return kubeClient.CoreV1().Pods(conf.Namespace).List(options)
				},
				WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
					return kubeClient.CoreV1().Pods(conf.Namespace).Watch(options)
				},
			},
			&api_v1.Pod{},
			0, // Skip resync
			cache.Indexers{},
		)

		// create resource controller using above informer and workqueue
		c := newResourceController(kubeClient, eventHandler, informer, "pod")
		stopCh := make(chan struct{})
		defer close(stopCh)

		// Run controller
		go c.Run(stopCh)
	}
}

// create a instance of particulr resource controller: contains both informer and workqueue
// Each resource must have their controller
func newResourceController(client kubernetes.Interface, handlers handlers.Handler, informer cache.SharedIndexInformer, resourceType string) *Controller {
	// create a queue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	var newEvent Event
	var err error
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.eventType = "create"
			newEvent.resourceType = resourceType
			logrus.WithField("pkg", "kubewatchpod-"+resourceType).Infof("Processing add to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(old)
			newEvent.eventType = "update"
			newEvent.resourceType = resourceType
			logrus.WithField("pkg", "kubewatchpod-"+resourceType).Infof("Processing update to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newEvent.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			newEvent.eventType = "delete"
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(obj).Namespace
			logrus.WithField("pkg", "kubewatchpod-"+resourceType).Infof("Processing delete to %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
	})

	return &Controller{
		logger:       logrus.WithField("pkg", "kubewatch-"+resourceType),
		clientset:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: eventHandler,
	}
}

func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("Starting kubewatchpod controller")
	serverStartTime := time.Now().Local()
	fmt.Println("serverStartTime: ", serverStartTime)

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}
	c.logger.Info("Kubewatchpod controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}
