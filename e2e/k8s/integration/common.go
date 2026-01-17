//go:build e2e

package integration

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var secretSantaGVR = schema.GroupVersionResource{
	Group:    "secrets.secret-santa.io",
	Version:  "v1alpha1",
	Resource: "secretsantas",
}
