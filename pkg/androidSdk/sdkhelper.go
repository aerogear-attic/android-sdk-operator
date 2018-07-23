package androidSdk

import (

	"fmt"
	b64 "encoding/base64"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	InstallerPodName = "android-sdk-pkg-install"
	UpdaterPodName   = "android-sdk-pkg-update"
)

func DefaultSdkHelper() SdkHelper {
	return SdkHelper {
		Image: "docker.io/aerogear/digger-android-sdk-image:1.0.0-alpha",
	}
}

func NewSdkhelper(image string) SdkHelper {
	return SdkHelper {
		Image: image,
	}
}

func (helper *SdkHelper) IsInstallerPod(name string) bool {
	return name == InstallerPodName
}

func (helper *SdkHelper) EncodeData(configMap *corev1.ConfigMap) (string, error) {
	config, err := helper.GetConfigData(configMap)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve config data: %v", err)
	}
	return b64.StdEncoding.EncodeToString([]byte(config)), nil
}

func (helper *SdkHelper) GetConfigData(configMap *corev1.ConfigMap) (string, error) {
	data, ok := configMap.Data["packages"]
	if !ok {
		return "", fmt.Errorf("\"packages\" key not found in configmap")
	}

	return data, nil
}

func (helper *SdkHelper) GetSdkPod(cmd []string, podName string, configMapName string, ns string) *corev1.Pod {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "android-sdk-pkg",
					Image:   helper.Image,
					Command: cmd,
					VolumeMounts: []corev1.VolumeMount{
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
			Volumes: []corev1.Volume{
				{
					Name: "android-sdk",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "android-sdk",
							ReadOnly:  false,
						},
					},
				},
				{
					Name: "android-sdk-config",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configMapName,
							},
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	return pod
}

