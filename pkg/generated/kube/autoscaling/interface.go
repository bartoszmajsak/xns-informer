// Code generated by xns-informer-gen. DO NOT EDIT.

package autoscaling

import (
	v1 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v1"
	v2beta1 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v2beta1"
	v2beta2 "github.com/maistra/xns-informer/pkg/generated/kube/autoscaling/v2beta2"
	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
)

type Interface interface {
	V1() v1.Interface
	V2beta1() v2beta1.Interface
	V2beta2() v2beta2.Interface
}

type group struct {
	factory xnsinformers.SharedInformerFactory
}

func New(factory xnsinformers.SharedInformerFactory) Interface {
	return &group{factory: factory}
}

// V1 returns a new v1.Interface.
func (g *group) V1() v1.Interface {
	return v1.New(g.factory)
}

// V2beta1 returns a new v2beta1.Interface.
func (g *group) V2beta1() v2beta1.Interface {
	return v2beta1.New(g.factory)
}

// V2beta2 returns a new v2beta2.Interface.
func (g *group) V2beta2() v2beta2.Interface {
	return v2beta2.New(g.factory)
}