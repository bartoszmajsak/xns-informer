// Code generated by xns-informer-gen. DO NOT EDIT.

package v1

import (
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
	v1 "k8s.io/api/core/v1"
	informers "k8s.io/client-go/informers/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type limitRangeInformer struct {
	informer cache.SharedIndexInformer
}

var _ informers.LimitRangeInformer = &limitRangeInformer{}

func NewLimitRangeInformer(f xnsinformers.SharedInformerFactory) informers.LimitRangeInformer {
	resource := v1.SchemeGroupVersion.WithResource("limitranges")
	informer := f.NamespacedResource(resource).Informer()

	return &limitRangeInformer{
		informer: xnsinformers.NewInformerConverter(f.GetScheme(), informer, &v1.LimitRange{}),
	}
}

func (i *limitRangeInformer) Informer() cache.SharedIndexInformer {
	return i.informer
}

func (i *limitRangeInformer) Lister() listers.LimitRangeLister {
	return listers.NewLimitRangeLister(i.informer.GetIndexer())
}