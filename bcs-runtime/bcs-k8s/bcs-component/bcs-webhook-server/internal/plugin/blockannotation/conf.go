package blockannotation

// BlockCondition block condition for k8s resource
type BlockCondition struct {
	Reference  string `json:"reference"`
	Operator   string `json:"operator"`
	FailPolicy string `json:"failPolicy,omitempty"`
}

// Config config content for annotation blocker
type Config struct {
	Namespaces    []string          `json:"namespaces"`
	AnnotationKey string            `json:"annotationKey"`
	Conditions    []*BlockCondition `json:"conditions,omitempty"`
}
