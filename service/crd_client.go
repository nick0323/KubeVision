package service

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

type CRDManager struct {
	dynamicClient dynamic.Interface
	logger        *zap.Logger
}

func NewCRDManager(restConfig *rest.Config, logger *zap.Logger) (*CRDManager, error) {
	if restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}
	return &CRDManager{dynamicClient: dynamicClient, logger: logger}, nil
}

type CRDSummary struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Version     string `json:"version"`
	Kind        string `json:"kind"`
	Plural      string `json:"plural"`
	Scope       string `json:"scope"`
	InstanceCnt int    `json:"instanceCnt"`
}

type CRDInstance struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Age        string `json:"age"`
}

func (m *CRDManager) ListCRDs(ctx context.Context) ([]CRDSummary, error) {
	if m == nil {
		return nil, fmt.Errorf("CRD manager not initialized")
	}

	list, err := m.dynamicClient.Resource(crdGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}

	result := make([]CRDSummary, 0, len(list.Items))
	for _, crd := range list.Items {
		name := crd.GetName()
		spec, ok := crd.Object["spec"].(map[string]interface{})
		if !ok {
			continue
		}

		group, _ := spec["group"].(string)
		scope, _ := spec["scope"].(string)
		plural := ""
		kind := ""

		versions, ok := spec["versions"].([]interface{})
		if !ok {
			version, _ := spec["version"].(string)
			versions = []interface{}{map[string]interface{}{"name": version, "served": true, "storage": true}}
		}

		names, ok := spec["names"].(map[string]interface{})
		if ok {
			if p, _ := names["plural"].(string); p != "" {
				plural = p
			}
			if k, _ := names["kind"].(string); k != "" {
				kind = k
			}
		}

		storageVersion := ""
		servedVersions := []string{}
		versions, ok = spec["versions"].([]interface{})
		if ok {
			for _, v := range versions {
				ver, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				verName, _ := ver["name"].(string)
				served, _ := ver["served"].(bool)
				stored, _ := ver["storage"].(bool)
				if served {
					servedVersions = append(servedVersions, verName)
				}
				if stored {
					storageVersion = verName
				}
			}
		}
		if storageVersion == "" && len(servedVersions) > 0 {
			storageVersion = servedVersions[0]
		}

		versionToUse := storageVersion
		if versionToUse == "" && len(servedVersions) > 0 {
			versionToUse = servedVersions[0]
		}

		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  versionToUse,
			Resource: plural,
		}

		instanceCnt := 0
		if gvr.Resource != "" && gvr.Version != "" {
			var instanceList *unstructured.UnstructuredList
			if strings.ToLower(scope) == "cluster" {
				instanceList, err = m.dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 1})
			} else {
				instanceList, err = m.dynamicClient.Resource(gvr).Namespace("").List(ctx, metav1.ListOptions{Limit: 1})
			}
			if err == nil && instanceList != nil {
				instanceCnt = len(instanceList.Items)
			}
		}

		result = append(result, CRDSummary{
			Name:        name,
			Group:       group,
			Version:     versionToUse,
			Kind:        kind,
			Plural:      plural,
			Scope:       scope,
			InstanceCnt: instanceCnt,
		})
	}

	return result, nil
}

func (m *CRDManager) ListCRDInstances(ctx context.Context, group, version, plural, namespace string) (*unstructured.UnstructuredList, error) {
	if m == nil {
		return nil, fmt.Errorf("CRD manager not initialized")
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: plural,
	}

	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}

	return m.dynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
}

func (m *CRDManager) GetCRDInstance(ctx context.Context, group, version, plural, namespace, name string) (*unstructured.Unstructured, error) {
	if m == nil {
		return nil, fmt.Errorf("CRD manager not initialized")
	}

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: plural,
	}

	return m.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
}
