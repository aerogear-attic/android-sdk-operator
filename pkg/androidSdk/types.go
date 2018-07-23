package androidSdk

import (
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/api/core/v1"
)

type Handler struct {
	kube HandlerKube
	sdk HandlerSdk
	encodedConfig string
}

type HandlerSdk interface {
	IsInstallerPod(name string) bool
	EncodeData(configMap *corev1.ConfigMap) (string, error)
	GetConfigData(configMap *corev1.ConfigMap) (string, error)
	GetSdkPod(cmd []string, podName string, configMapName string, ns string) *corev1.Pod
}

type HandlerKube interface {
	GetPod(name string, ns string) (*corev1.Pod, error)
	GetConfigMap(name string, ns string) (*corev1.ConfigMap, error)
}

type SdkHelper struct {
	Image string
}

type Kube struct {
	client kubernetes.Interface
}
