// Code generated by xns-informer-gen. DO NOT EDIT.
package v1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type serviceInformer struct {
	factory xnsinformers.InformerFactory
}

var _ informers.ServiceInformer = &serviceInformer{}

func (f *serviceInformer) resource() schema.GroupVersionResource {
	return v1.SchemeGroupVersion.WithResource("services")
}

func (f *serviceInformer) Informer() cache.SharedIndexInformer {
	return f.factory.NamespacedResource(f.resource()).Informer()
}

func (f *serviceInformer) Lister() listers.ServiceLister {
	return &serviceLister{lister: f.factory.NamespacedResource(f.resource()).Lister()}
}

type serviceLister struct {
	lister cache.GenericLister
}

var _ listers.ServiceLister = &serviceLister{}

func (l *serviceLister) List(selector labels.Selector) (res []*v1.Service, err error) {
	return listService(l.lister, selector)
}

func (l *serviceLister) Services(namespace string) listers.ServiceNamespaceLister {
	return &serviceNamespaceLister{lister: l.lister.ByNamespace(namespace)}
}

type serviceNamespaceLister struct {
	lister cache.GenericNamespaceLister
}

var _ listers.ServiceNamespaceLister = &serviceNamespaceLister{}

func (l *serviceNamespaceLister) List(selector labels.Selector) (res []*v1.Service, err error) {
	return listService(l.lister, selector)
}

func (l *serviceNamespaceLister) Get(name string) (*v1.Service, error) {
	obj, err := l.lister.Get(name)
	if err != nil {
		return nil, err
	}

	out := &v1.Service{}
	if err := xnsinformers.ConvertUnstructured(obj, out); err != nil {
		return nil, err
	}

	return out, nil
}

func listService(l xnsinformers.SimpleLister, s labels.Selector) (res []*v1.Service, err error) {
	objects, err := l.List(s)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		out := &v1.Service{}
		if err := xnsinformers.ConvertUnstructured(obj, out); err != nil {
			return nil, err
		}

		res = append(res, out)
	}

	return res, nil
}