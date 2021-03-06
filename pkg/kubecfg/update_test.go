package kubecfg

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ksonnet/kubecfg/utils"
)

func TestStringListContains(t *testing.T) {
	foobar := []string{"foo", "bar"}
	if stringListContains([]string{}, "") {
		t.Error("Empty list was not empty")
	}
	if !stringListContains(foobar, "foo") {
		t.Error("Failed to find foo")
	}
	if stringListContains(foobar, "baz") {
		t.Error("Should not contain baz")
	}
}

func TestEligibleForGc(t *testing.T) {
	const myTag = "my-gctag"
	boolTrue := true
	o := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tests/v1alpha1",
			"kind":       "Dummy",
		},
	}

	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (no tag)", o)
	}

	// [gctag-migration]: Remove annotation in phase2
	utils.SetMetaDataAnnotation(o, AnnotationGcTag, "unknowntag")
	utils.SetMetaDataLabel(o, LabelGcTag, "unknowntag")
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (wrong tag)", o)
	}

	// [gctag-migration]: Remove annotation in phase2
	utils.SetMetaDataAnnotation(o, AnnotationGcTag, myTag)
	utils.SetMetaDataLabel(o, LabelGcTag, myTag)
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible", o)
	}

	// [gctag-migration]: Remove testcase in phase2
	utils.SetMetaDataAnnotation(o, AnnotationGcTag, myTag)
	delete(o.GetLabels(), LabelGcTag) // no label. ie: pre-migration
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible (gctag-migration phase1)", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcStrategy, GcStrategyIgnore)
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (strategy=ignore)", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcStrategy, GcStrategyAuto)
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible (strategy=auto)", o)
	}

	// Unstructured.SetOwnerReferences is broken in apimachinery release-1.6
	// See kubernetes/kubernetes#46817
	setOwnerRef := func(u *unstructured.Unstructured, ref metav1.OwnerReference) {
		// This is not a complete nor robust reimplementation
		c := map[string]interface{}{
			"kind": ref.Kind,
			"name": ref.Name,
		}
		if ref.Controller != nil {
			c["controller"] = *ref.Controller
		}
		u.Object["metadata"].(map[string]interface{})["ownerReferences"] = []interface{}{c}
	}
	setOwnerRef(o, metav1.OwnerReference{Kind: "foo", Name: "bar"})
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible (non-controller ownerref)", o)
	}

	setOwnerRef(o, metav1.OwnerReference{Kind: "foo", Name: "bar", Controller: &boolTrue})
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (controller ownerref)", o)
	}
}
