package informers_test

import (
	"reflect"
	"sort"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xnsinformers "github.com/maistra/xns-informer/pkg/informers"
)

func TestNamespaceSet(t *testing.T) {
	testCases := []struct {
		name            string
		namespaceSet    xnsinformers.NamespaceSet
		testFunc        func(ns xnsinformers.NamespaceSet)
		expectedAdds    []string
		expectedRemoves []string
	}{
		{
			name:         "initially empty",
			namespaceSet: xnsinformers.NewNamespaceSet(),
			testFunc: func(ns xnsinformers.NamespaceSet) {
				ns.SetNamespaces([]string{"ns-one", "ns-two"})
			},
			expectedAdds:    []string{"ns-one", "ns-two"},
			expectedRemoves: nil,
		},
		{
			name:         "initially populated",
			namespaceSet: xnsinformers.NewNamespaceSet("ns-one"),
			testFunc: func(ns xnsinformers.NamespaceSet) {
				ns.SetNamespaces([]string{"ns-one", "ns-two", "ns-three"})
				ns.SetNamespaces([]string{"new-ns"})
			},
			expectedAdds:    []string{"ns-one", "ns-two", "ns-three", "new-ns"},
			expectedRemoves: []string{"ns-one", "ns-two", "ns-three"},
		},
		{
			name:         "includes metav1.NamespaceAll",
			namespaceSet: xnsinformers.NewNamespaceSet(),
			testFunc: func(ns xnsinformers.NamespaceSet) {
				// Adding metav1.NamespaceAll means all others are ignored.
				ns.SetNamespaces([]string{metav1.NamespaceAll, "ns-ignored"})
			},
			expectedAdds:    []string{metav1.NamespaceAll},
			expectedRemoves: nil,
		},
		{
			name:         "uninitialized later",
			namespaceSet: xnsinformers.NewNamespaceSet("ns-one"),
			testFunc: func(ns xnsinformers.NamespaceSet) {
				ns.SetNamespaces(nil)
			},
			expectedAdds:    []string{"ns-one"},
			expectedRemoves: []string{"ns-one"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var adds, removes []string

			tc.namespaceSet.AddHandler(xnsinformers.NamespaceSetHandlerFuncs{
				AddFunc: func(ns string) {
					adds = append(adds, ns)
				},
				RemoveFunc: func(ns string) {
					removes = append(removes, ns)
				},
			})

			tc.testFunc(tc.namespaceSet)

			sort.Strings(adds)
			sort.Strings(removes)
			sort.Strings(tc.expectedAdds)
			sort.Strings(tc.expectedRemoves)

			if !reflect.DeepEqual(tc.expectedAdds, adds) {
				t.Errorf("%v ≠ %v", tc.expectedAdds, adds)
			}

			if !reflect.DeepEqual(tc.expectedRemoves, removes) {
				t.Errorf("%v ≠ %v", tc.expectedRemoves, removes)
			}
		})
	}
}

func TestNamespaceSetInitialized(t *testing.T) {
	set := xnsinformers.NewNamespaceSet()
	if set.Initialized() {
		t.Errorf("didn't expect new NamespaceSet to be initialized")
	}

	set.SetNamespaces([]string{"foo"})
	if !set.Initialized() {
		t.Errorf("expected NamespaceSet to be initialized after invoking SetNamespaces()")
	}

	set.SetNamespaces([]string{})
	if !set.Initialized() {
		t.Errorf("expected NamespaceSet to still be initialized after invoking SetNamespaces() with empty namespace slice")
	}

	set.SetNamespaces(nil)
	if set.Initialized() {
		t.Errorf("didn't expect NamespaceSet to still be initialized after invoking SetNamespaces() with nil namespace slice")
	}
}

func TestNamespaceSetList(t *testing.T) {
	testCases := []struct {
		name         string
		namespaceSet xnsinformers.NamespaceSet
		expectedList []string
	}{
		{
			name:         "empty",
			namespaceSet: xnsinformers.NewNamespaceSet(),
			expectedList: []string{},
		},
		{
			name:         "uninitialized",
			namespaceSet: xnsinformers.NewUninitializedNamespaceSet(),
			expectedList: []string{},
		},
		{
			name:         "populated",
			namespaceSet: xnsinformers.NewNamespaceSet("c", "a", "b"),
			expectedList: []string{"a", "b", "c"}, // Should be sorted.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.namespaceSet.List()

			if !reflect.DeepEqual(tc.expectedList, got) {
				t.Errorf("%v ≠ %v", tc.expectedList, got)
			}
		})
	}
}

func TestNamespaceSetContains(t *testing.T) {
	testCases := []struct {
		name         string
		namespaceSet xnsinformers.NamespaceSet
		search       string
		expected     bool
	}{
		{
			name:         "empty",
			namespaceSet: xnsinformers.NewNamespaceSet(),
			search:       "a",
			expected:     false,
		},
		{
			name:         "uninitialized",
			namespaceSet: xnsinformers.NewUninitializedNamespaceSet(),
			search:       "a",
			expected:     false,
		},
		{
			name:         "found",
			namespaceSet: xnsinformers.NewNamespaceSet("c", "a", "b"),
			search:       "b",
			expected:     true,
		},
		{
			name:         "not found",
			namespaceSet: xnsinformers.NewNamespaceSet("e", "f", "g"),
			search:       "z",
			expected:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			found := tc.namespaceSet.Contains(tc.search)

			if tc.expected != found {
				t.Errorf("expected (%t) ≠ found (%t)", tc.expected, found)
			}
		})
	}
}
