//go:build e2e

package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	secretSantaGVR = schema.GroupVersionResource{
		Group:    "secrets.secret-santa.io",
		Version:  "v1alpha1",
		Resource: "secretsanta",
	}

	pollInterval = 2 * time.Second
	pollTimeout  = 90 * time.Second
)

// vaultLocalAddr is used by the test process to verify secrets (via port-forward).
func vaultLocalAddr() string {
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		return addr
	}
	return "http://127.0.0.1:8200"
}

// vaultInClusterAddr is passed to the SecretSanta CR so the controller can reach Vault inside the cluster.
func vaultInClusterAddr() string {
	if addr := os.Getenv("VAULT_IN_CLUSTER_ADDR"); addr != "" {
		return addr
	}
	return "http://vault.vault.svc.cluster.local:8200"
}

func vaultToken() string {
	if tok := os.Getenv("VAULT_TOKEN"); tok != "" {
		return tok
	}
	return "root"
}

func TestVaultSecretStorage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatalf("failed to get kubeconfig: %v", err)
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		t.Fatalf("failed to create dynamic client: %v", err)
	}

	namespace := "default"
	name := "vault-e2e-test"
	secretPath := fmt.Sprintf("%s/%s", namespace, name)

	secretSanta := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "secrets.secret-santa.io/v1alpha1",
			"kind":       "SecretSanta",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"template": `{"password":"{{ .Pass.value }}"}`,
				"generators": []interface{}{
					map[string]interface{}{
						"name": "Pass",
						"type": "random_password",
						"config": map[string]interface{}{
							"length": float64(24),
						},
					},
				},
				"media": map[string]interface{}{
					"type": "hashicorp-vault",
					"config": map[string]interface{}{
						"address":    vaultInClusterAddr(),
						"mount_path": "secret",
						"path":       secretPath,
						"token":      vaultToken(),
					},
				},
			},
		},
	}

	_, err = dynClient.Resource(secretSantaGVR).Namespace(namespace).Create(ctx, secretSanta, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create SecretSanta: %v", err)
	}
	t.Cleanup(func() {
		delCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if delErr := dynClient.Resource(secretSantaGVR).Namespace(namespace).Delete(delCtx, name, metav1.DeleteOptions{}); delErr != nil {
			t.Logf("failed to delete SecretSanta: %v", delErr)
		}
	})

	// Wait for the SecretSanta to reach Ready condition
	err = wait.PollUntilContextTimeout(ctx, pollInterval, pollTimeout, true, func(ctx context.Context) (bool, error) {
		obj, err := dynClient.Resource(secretSantaGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if cond["type"] == "Ready" && cond["status"] == "True" {
				return true, nil
			}
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("SecretSanta did not become Ready: %v", err)
	}

	// Verify the secret exists in Vault via the HTTP API (KV v2)
	data, err := readVaultKVSecret(ctx, vaultLocalAddr(), vaultToken(), "secret", secretPath)
	if err != nil {
		t.Fatalf("failed to read secret from Vault: %v", err)
	}
	if _, ok := data["password"]; !ok {
		t.Errorf("expected 'password' key in Vault secret, got keys: %v", data)
	}
	t.Logf("Vault secret verified successfully")
}

// readVaultKVSecret reads a KV v2 secret using the Vault HTTP API directly,
// avoiding an extra SDK dependency in the e2e module.
func readVaultKVSecret(ctx context.Context, addr, token, mount, path string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", addr, mount, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Vault-Token", token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vault returned HTTP %d for path %s", resp.StatusCode, path)
	}

	var result struct {
		Data struct {
			Data map[string]interface{} `json:"data"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data.Data, nil
}
