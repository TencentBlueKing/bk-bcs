package discovery

import (
	"fmt"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/pkg/meta"

	"k8s.io/apimachinery/pkg/labels"
	k8scache "k8s.io/client-go/tools/cache"
)

// AppendFunc is used to add a matching item to whatever list the caller is using
type AppendFunc func(interface{})

func ListAll(store k8scache.Indexer, selector labels.Selector, appendFn AppendFunc) error {
	selectAll := selector.Empty()
	for _, m := range store.List() {
		if selectAll {
			// Avoid computing labels of the objects to speed up common flows
			// of listing all objects.
			appendFn(m)
			continue
		}
		metadata, err := meta.Accessor(m)
		if err != nil {
			return err
		}
		if selector.Matches(labels.Set(metadata.GetLabels())) {
			appendFn(m)
		}
	}
	return nil
}

func ListAllByNamespace(indexer k8scache.Indexer, namespace string, selector labels.Selector, appendFn AppendFunc) error {
	selectAll := selector.Empty()
	if namespace == meta.NamespaceAll {
		for _, m := range indexer.List() {
			if selectAll {
				// Avoid computing labels of the objects to speed up common flows
				// of listing all objects.
				appendFn(m)
				continue
			}
			metadata, err := meta.Accessor(m)
			if err != nil {
				return err
			}
			if selector.Matches(labels.Set(metadata.GetLabels())) {
				appendFn(m)
			}
		}
		return nil
	}

	items, err := indexer.Index(meta.NamespaceIndex, &meta.ObjectMeta{Namespace: namespace})
	if err != nil {
		// Ignore error; do slow search without index.
		blog.Warn("can not retrieve list of objects using index : %v", err)
		for _, m := range indexer.List() {
			metadata, err := meta.Accessor(m)
			if err != nil {
				return err
			}
			if metadata.GetNamespace() == namespace && selector.Matches(labels.Set(metadata.GetLabels())) {
				appendFn(m)
			}

		}
		return nil
	}
	for _, m := range items {
		if selectAll {
			// Avoid computing labels of the objects to speed up common flows
			// of listing all objects.
			appendFn(m)
			continue
		}
		metadata, err := meta.Accessor(m)
		if err != nil {
			return err
		}
		if selector.Matches(labels.Set(metadata.GetLabels())) {
			appendFn(m)
		}
	}

	return nil
}

//only taskgroup has the application index
func ListAllByApplication(indexer k8scache.Indexer, ns, appname string, selector labels.Selector, appendFn AppendFunc) error {
	selectAll := selector.Empty()
	if ns == meta.NamespaceAll {
		for _, m := range indexer.List() {
			if selectAll {
				// Avoid computing labels of the objects to speed up common flows
				// of listing all objects.
				appendFn(m)
				continue
			}
			metadata, err := meta.Accessor(m)
			if err != nil {
				return err
			}
			if selector.Matches(labels.Set(metadata.GetLabels())) {
				appendFn(m)
			}
		}
		return nil
	}

	items, err := indexer.ByIndex(meta.ApplicationIndex, fmt.Sprintf("%s.%s", appname, ns))
	if err != nil {
		// Ignore error; do slow search without index.
		blog.Warn("can not retrieve list of objects using index : %v", err)
		return fmt.Errorf("can not retrieve list of objects using index: %s", err.Error())
	}
	for _, m := range items {
		if selectAll {
			// Avoid computing labels of the objects to speed up common flows
			// of listing all objects.
			appendFn(m)
			continue
		}
		metadata, err := meta.Accessor(m)
		if err != nil {
			return err
		}
		if selector.Matches(labels.Set(metadata.GetLabels())) {
			appendFn(m)
		}
	}

	return nil
}

func EverythingSelector() labels.Selector {
	return labels.Everything()
}
