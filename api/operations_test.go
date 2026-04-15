package api

import "testing"

func TestValidateResourceIdentityNamespaced(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "demo",
			"namespace": "default",
		},
	}

	if err := validateResourceIdentity("deployment", "default", "demo", obj); err != nil {
		t.Fatalf("expected identity validation to pass, got %v", err)
	}
}

func TestValidateResourceIdentityRejectsMismatchedName(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "other",
			"namespace": "default",
		},
	}

	if err := validateResourceIdentity("deployment", "default", "demo", obj); err == nil {
		t.Fatal("expected mismatched metadata.name to fail validation")
	}
}

func TestValidateResourceIdentityRejectsClusterNamespace(t *testing.T) {
	obj := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "node-1",
			"namespace": "default",
		},
	}

	if err := validateResourceIdentity("node", "", "node-1", obj); err == nil {
		t.Fatal("expected cluster-scoped resource with namespace to fail validation")
	}
}
