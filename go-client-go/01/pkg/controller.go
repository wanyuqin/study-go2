package pkg

import (
	"context"
	v15 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v16 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informer "k8s.io/client-go/informers/core/v1"
	netInformer "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	coreListers "k8s.io/client-go/listers/core/v1"
	v1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"reflect"
	"time"
)

const (
	workNum  = 5
	maxRetry = 10
)

type Controller struct {
	client        kubernetes.Interface
	ingressLister v1.IngressLister
	serviceLister coreListers.ServiceLister
	queue         workqueue.RateLimitingInterface
}

func NewController(client kubernetes.Interface, serviceInformer informer.ServiceInformer, ingressInformer netInformer.IngressInformer) Controller {
	c := Controller{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		serviceLister: serviceInformer.Lister(),
		queue: workqueue.NewRateLimitingQueueWithConfig(workqueue.DefaultControllerRateLimiter(), workqueue.RateLimitingQueueConfig{
			Name: "ingressManager",
		}),
	}

	serviceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.AddService,
		UpdateFunc: c.UpdateService,
		DeleteFunc: c.DeleteService,
	})

	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.AddIngress,
		UpdateFunc: c.UpdateIngress,
		DeleteFunc: c.DeleteIngress,
	})

	return c
}

func (c *Controller) AddService(obj interface{}) {
	c.enQueue(obj)
}

func (c *Controller) UpdateService(oldObj interface{}, newObj interface{}) {
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	// TODO 比较annotations 是否相同

}

func (c *Controller) enQueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
	}

	c.queue.Add(key)
}

func (c *Controller) DeleteService(obj interface{}) {

}

func (c *Controller) AddIngress(obj interface{}) {

}

func (c *Controller) UpdateIngress(oldObj interface{}, newObj interface{}) {

}

func (c *Controller) DeleteIngress(obj interface{}) {
	ingress := obj.(*v15.Ingress)
	service := v16.GetControllerOf(ingress)
	if service == nil {
		return
	}
	if service.Kind != "Service" {
		return
	}

	c.queue.Add(ingress.Namespace + "/" + ingress.Name)
}

func (c *Controller) Run(stopCh chan struct{}) {
	for i := 0; i < workNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCh)
	}
	<-stopCh
}

func (c *Controller) worker() {
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}

	key := item.(string)
	defer c.queue.Done(key)

	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}

func (c *Controller) syncService(item string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(item)
	if err != nil {
		return err
	}
	// 删除
	service, err := c.serviceLister.Services(namespace).Get(name)
	if errors.IsNotFound(err) {
		return nil
	}

	// 新增
	_, ok := service.GetAnnotations()["ingres/http"]
	ingress, err := c.ingressLister.Ingresses(namespace).Get(name)

	if ok && errors.IsNotFound(err) {

		// create ingress
		ig := c.constructIngress(namespace, name)
		_, err := c.client.NetworkingV1().Ingresses(namespace).Create(context.Background(), ig, v16.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		// delete ingress
		c.client.NetworkingV1().Ingresses(namespace).Delete(context.Background(), name, v16.DeleteOptions{})
	}
	return nil
}

func (c *Controller) handlerError(item string, err error) {
	if c.queue.NumRequeues(item) <= maxRetry {
		c.queue.AddRateLimited(item)
		return
	}
	runtime.HandleError(err)
	c.queue.Forget(item)

}

func (c *Controller) constructIngress(namespace string, name string) *v15.Ingress {
	ingress := &v15.Ingress{}
	ingress.Namespace = namespace
	ingress.Name = name
	pathType := v15.PathTypePrefix
	ingress.Spec = v15.IngressSpec{
		Rules: []v15.IngressRule{{
			Host: "example.com",
			IngressRuleValue: v15.IngressRuleValue{
				HTTP: &v15.HTTPIngressRuleValue{
					Paths: []v15.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: v15.IngressBackend{
								Service: &v15.IngressServiceBackend{
									Name: name,
									Port: v15.ServiceBackendPort{
										Number: 80,
									},
								},
							},
						},
					},
				},
			},
		},
		},
	}
	return ingress

}
