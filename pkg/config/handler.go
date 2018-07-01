package config

import (
	"context"

	"github.com/aerogear/android-sdk-operator/pkg/apis/androidsdk/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewHandler(k8c kubernetes.Interface) sdk.Handler {
	return &Handler {
		k8c:k8c,
	}
}

type Handler struct {
	k8c kubernetes.Interface
	Data string
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	logrus.Infof("Handler is being called")

	switch o := event.Object.(type) {
	case *corev1.ConfigMap:
		isValid := v1.IsValidSdkConfig(o)
		if !isValid {
			logrus.Infof("ConfigMap %s is not a valid android-sdk-config object", o.Name)
			return nil
		}

		cfgStr, cfgStrErr := v1.GetConfigData(o)
		if cfgStrErr != nil {
			return cfgStrErr
		}

		resource := h.updateSdkResource(cfgStr)
		resourceErr := sdk.Create(resource)
		if resourceErr != nil && kerrors.IsAlreadyExists(resourceErr) {
			logrus.Infof("AndroidSDK resource is already created.")
		}

		//TODO: need to persist status for pod execution

		installPod := runSdkPod(h, []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}, "android-sdk-pkg-update")
		installPodErr := sdk.Create(installPod)
		if installPodErr != nil {
			return installPodErr
		}

	}
	return nil
}

func (h *Handler) updateSdkResource(cfg string) *v1.AndroidSDK {
	androidSdk := &v1.AndroidSDK {
		TypeMeta: metav1.TypeMeta {
			Kind:       "AndroidSDK",
			APIVersion: "androidsdk.aerogear.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta {
			Name:      "android-sdk-config-object",
			Namespace: "android",
		},
		Spec: v1.AndroidSDKSpec{
			Data: cfg,
		},
		Status:v1.AndroidSDKStatus{
			Status: v1.Done,
		},
	}
	return androidSdk
}


func runSdkPod(h *Handler, cmd []string, name string) *corev1.Pod {
	pod := &corev1.Pod {
		TypeMeta: metav1.TypeMeta {
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta {
			Name:      name,
			Namespace: "android",
		},
		Spec: corev1.PodSpec {
			Containers: []corev1.Container {
				{
					Name: "android-sdk-pkg",
					Image: "docker.io/aerogear/digger-android-sdk-image:dev",
					Command: cmd,
					VolumeMounts: []corev1.VolumeMount {
						{
							Name: "android-sdk",
							MountPath: "/opt/android-sdk-linux",
						},
						{
							Name: "android-sdk-config",
							MountPath: "/tmp/android-sdk-config",
						},
					},
				},
			},
			Volumes:[]corev1.Volume{
				{
					Name:"android-sdk",
					VolumeSource: corev1.VolumeSource {
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "android-sdk",
							ReadOnly: false,
						},
					},
				},
				{
					Name: "android-sdk-config",
					VolumeSource: corev1.VolumeSource {
						ConfigMap:&corev1.ConfigMapVolumeSource {
							LocalObjectReference: corev1.LocalObjectReference {
								Name: "android-sdk-config",
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
