package quota

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestQuota(t *testing.T) {
	q1, err := resource.ParseQuantity("10m")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q1.AsApproximateFloat64())

	q2, err := resource.ParseQuantity("1")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q2.AsApproximateFloat64())
}
