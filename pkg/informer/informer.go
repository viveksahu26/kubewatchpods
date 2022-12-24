package informer


queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
fmt.Println("queue: ", queue)
informer := cache.NewSharedIndexInformer(
	&cache.ListWatch{
		   ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				  return client.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
		   },
		   WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				  return client.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
		   },
	},
	&api_v1.Pod{},
	0, //Skip resync
	cache.Indexers{},
)