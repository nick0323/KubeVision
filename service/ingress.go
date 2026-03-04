package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/nick0323/K8sVision/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.IngressStatus, error) {
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		zap.L().Error("ListIngresses failed", zap.Error(err))
		return nil, err
	}
	result := make([]model.IngressStatus, 0, len(ingresses.Items))
	for _, ing := range ingresses.Items {
		hosts := make([]string, 0)
		paths := make([]string, 0)
		targetSvcs := make([]string, 0)
		for _, rule := range ing.Spec.Rules {
			hosts = append(hosts, rule.Host)
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, path.Path)
					targetSvcs = append(targetSvcs, path.Backend.Service.Name)
				}
			}
		}
		address := ""
		if len(ing.Status.LoadBalancer.Ingress) > 0 {
			address = ing.Status.LoadBalancer.Ingress[0].IP
			if address == "" {
				address = ing.Status.LoadBalancer.Ingress[0].Hostname
			}
		}
		class := ""
		if ing.Spec.IngressClassName != nil {
			class = *ing.Spec.IngressClassName
		}
		result = append(result, model.IngressStatus{
			Namespace:     ing.Namespace,
			Name:          ing.Name,
			Hosts:         hosts,
			Address:       address,
			Ports:         model.DefaultIngressPorts,
			Class:         class,
			Status:        "Active",
			Path:          paths,
			TargetService: targetSvcs,
		})
	}
	return result, nil
}
