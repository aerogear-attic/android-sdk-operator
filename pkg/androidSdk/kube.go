package androidSdk

import (
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKube(k8c kubernetes.Interface) Kube {
	return Kube {
		client: k8c,
	}
}

func (kube *Kube) GetPod(name string, ns string) (*v1.Pod, error) {
	pods := kube.client.CoreV1().Pods(ns)
	return pods.Get(name, metav1.GetOptions{})
}

func (kube *Kube) GetConfigMap(name string, ns string) (*v1.ConfigMap, error) {
	configMap, err := kube.client.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configMap, nil
}
