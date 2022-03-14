package main

import (
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"
)

type Controller struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface
}

// 实例化Controller。
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

var kubeconfig *string
var namespace string

func main() {
	klog.InitFlags(nil)
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig file")
	}
	flag.StringVar(&namespace, "namespace", "default", "命名空间")

	flag.Parse()
	defer klog.Flush()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	// 创建pod的list watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", namespace, fields.Everything())

	// 创建workqueue, 默认速率是10qps
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// 创建indexer和infomer
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &corev1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer使用一个增量队列，因此我们必须使用这个键函数进行删除。
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer)

	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(5, stop)

	// 永远等待
	select {}
}

// 负责观察和同步
func (c *Controller) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()
	// 完工后让任务停下来
	defer c.queue.ShuttingDown()

	klog.Info("启动 Pod controller")

	// 启动监听
	go c.informer.Run(stopCh)

	// 处理之前，等待所有涉及的缓存被同步
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("同步超时了"))
	}

	// 创建多个work处理
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("停止 Pod controller")
}

func (c *Controller) runWorker()  {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(key)

	// 调用业务逻辑
	err := c.syncToStdout(key.(string))

	// 如果发现错误，处理错误，并重试
	c.handleErr(err, key)
	return true
}


// 业务逻辑
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf(" get key %s from index failed：%v", key, err)
		return err
	}

	if !exists {
		// Pod被删除了
		fmt.Printf("Pod %s 已经被删除了 \n", key)
	} else {
		fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*corev1.Pod).GetName())
	}
	return nil
}

//处理错误
func (c *Controller) handleErr(err error,key interface{})  {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	// 如果出现问题，这个控制器会重试5次。
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("同步pod %v 错误: %v", key, err)

		//重新排队，稍后重新再试
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
}
