package androidSdk

import (
	b64 "encoding/base64"
	"fmt"

	api "github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getPod returns the pod object given its name
func (h *Handler) getPod(name string, ns string) (*v1.Pod, error) {
	pods := h.k8c.CoreV1().Pods(ns)
	return pods.Get(name, metav1.GetOptions{})
}

// getConfigMap returns the configmap object given its name
func (h *Handler) getConfigMap(name string, ns string) (*v1.ConfigMap, error) {
	configMap, err := h.k8c.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// encodeData returns a b64 string of the configmap's data
func (h *Handler) encodeData(configMapName string, ns string) (string, error) {
	configMap, err := h.getConfigMap(configMapName, ns)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve configmap: %v", err)
	}

	config, err := api.GetConfigData(configMap)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve config data: %v", err)
	}
	return b64.StdEncoding.EncodeToString([]byte(config)), nil
}

// getSdkPod returns a pod based on the android sdk image
func (h *Handler) getSdkPod(cmd []string, podName string, configMapName string, ns string) *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "android-sdk-pkg",
					Image:   "docker.io/aerogear/digger-android-sdk-image:1.0.0-alpha",
					Command: cmd,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "android-sdk",
							MountPath: "/opt/android-sdk-linux",
						},
						{
							Name:      "android-sdk-config",
							MountPath: "/tmp/android-sdk-config",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "android-sdk",
					VolumeSource: v1.VolumeSource{
						PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
							ClaimName: "android-sdk",
							ReadOnly:  false,
						},
					},
				},
				{
					Name: "android-sdk-config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{
							LocalObjectReference: v1.LocalObjectReference{
								Name: configMapName,
							},
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}

	return pod
}
