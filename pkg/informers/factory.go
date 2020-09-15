package informers

import (
	"errors"
	"sync"
	"time"

	"github.com/maistra/xns-informer/pkg/internal/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type InformerFactory interface {
	Start(stopCh <-chan struct{})
	ClusterResource(resource schema.GroupVersionResource) informers.GenericInformer
	NamespacedResource(resource schema.GroupVersionResource) informers.GenericInformer
	WaitForCacheSync(stopCh <-chan struct{})
	SetNamespaces(namespaces []string)
}

// multiNamespaceInformerFactory provides a dynamic informer factory that
// creates informers which track changes across a set of namespaces.
type multiNamespaceInformerFactory struct {
	client       dynamic.Interface
	resyncPeriod time.Duration
	lock         sync.Mutex
	namespaces   sets.Set

	// Map of created informers by resource type.
	informers map[schema.GroupVersionResource]*multiNamespaceGenericInformer
}

var _ InformerFactory = &multiNamespaceInformerFactory{}

// NewInformerFactory returns a new factory for the given namespaces.
func NewInformerFactory(client dynamic.Interface, resync time.Duration, namespaces []string) (InformerFactory, error) {
	if len(namespaces) < 1 {
		return nil, errors.New("must provide at least one namespace")
	}

	factory := &multiNamespaceInformerFactory{
		client:       client,
		resyncPeriod: resync,
		informers:    make(map[schema.GroupVersionResource]*multiNamespaceGenericInformer),
	}

	factory.SetNamespaces(namespaces)

	return factory, nil
}

// SetNamespaces sets the list of namespaces the factory and its informers
// track.  Any new namespaces in the given set will be added to all previously
// created informers, and any namespaces that aren't in the new set will be
// removed.  You must call Start() and WaitForCacheSync() after changing the set
// of namespaces.  These are safe to call multiple times.
func (f *multiNamespaceInformerFactory) SetNamespaces(namespaces []string) {
	f.lock.Lock()
	defer f.lock.Unlock()

	newNamespaceSet := sets.NewSet(namespaces...)

	// If the set of namespaces, includes metav1.NamespaceAll, then it
	// only makes sense to create a single informer for that.
	if newNamespaceSet.Contains(metav1.NamespaceAll) {
		newNamespaceSet = sets.NewSet(metav1.NamespaceAll)
	}

	// Remove any namespaces in the current set which aren't in the
	// new set from the existing informers.
	for namespace := range f.namespaces.Difference(newNamespaceSet) {
		for _, i := range f.informers {
			i.informer.RemoveNamespace(namespace)
		}
	}

	f.namespaces = newNamespaceSet

	// Add any new namespaces to existing informers.
	for namespace := range f.namespaces {
		for _, i := range f.informers {
			i.informer.AddNamespace(namespace)
		}
	}
}

// ClusterResource returns a new cross-namespace informer for the given resource
// type and assumes it is cluster-scoped.  This means the returned informer will
// treat AddNamespace and RemoveNamespace as no-ops.
func (f *multiNamespaceInformerFactory) ClusterResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return f.ForResource(gvr, false)
}

// NamespacedResource returns a new cross-namespace informer for the given
// resource type and assumes it is namespaced.  Requesting a cluster-scoped
// resource via this method will result in errors from the underlying watch and
// will produce no events.
func (f *multiNamespaceInformerFactory) NamespacedResource(gvr schema.GroupVersionResource) informers.GenericInformer {
	return f.ForResource(gvr, true)
}

// ForResource returns a new cross-namespace informer for the given resource
// type.  If an informer for this resource type has been previously requested,
// it will be returned, otherwise a new one will be created.
//
// TODO: Should we use the discovery API to determine resource scope?
func (f *multiNamespaceInformerFactory) ForResource(gvr schema.GroupVersionResource, namespaced bool) informers.GenericInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	// Return existing informer if found.
	if informer, ok := f.informers[gvr]; ok {
		return informer
	}

	newInformerFunc := func(namespace string) informers.GenericInformer {
		// Namespace argument is ignored for cluster-scoped resources.
		if !namespaced {
			namespace = metav1.NamespaceAll
		}

		return dynamicinformer.NewFilteredDynamicInformer(
			f.client,
			gvr,
			namespace,
			f.resyncPeriod,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
			nil,
		)
	}

	informer := NewMultiNamespaceInformer(namespaced, f.resyncPeriod, newInformerFunc)
	lister := dynamiclister.New(informer.GetIndexer(), gvr)

	for namespace := range f.namespaces {
		informer.AddNamespace(namespace)
	}

	f.informers[gvr] = &multiNamespaceGenericInformer{
		informer: informer,
		lister:   dynamiclister.NewRuntimeObjectShim(lister),
	}

	return f.informers[gvr]
}

// Start starts all of the informers the factory has created to this point.
// They will be stopped when stopCh is closed.  Start is safe to call multiple
// times -- only stopped informers will be started.  This is non-blocking.
func (f *multiNamespaceInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for _, i := range f.informers {
		i.informer.NonBlockingRun(stopCh)
	}
}

// WaitForCacheSync waits for all previously started infomers caches to sync.
func (f *multiNamespaceInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	syncFuncs := func() (syncFuncs []cache.InformerSynced) {
		f.lock.Lock()
		defer f.lock.Unlock()

		for _, i := range f.informers {
			syncFuncs = append(syncFuncs, i.informer.HasSynced)
		}

		return
	}

	cache.WaitForCacheSync(stopCh, syncFuncs()...)
}
