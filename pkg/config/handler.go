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

		//TODO: there is probably a better way to delete the pod
		runningPod, err := h.getPod("android-sdk-pkg-update")
		if err == nil {
			if runningPod.Status.Phase == "Succeeded" {
				logrus.Infof("Android SDK Pod finished running, starting its removal.")
				delErr := sdk.Delete(getSdkPod(h, []string{""}, "android-sdk-pkg-update"), sdk.WithDeleteOptions(&metav1.DeleteOptions{}))
				if delErr != nil {
					logrus.Infof("Error while deleting pod %s.", runningPod.Name)
					return delErr
				}
				return nil
			}
			return nil
		}

		//TODO: persist config status in AndroidSDK CRD
		//TODO: Maybe we should store a hash string based on the config map content and only run the update if the hash is not a match
		installPod := getSdkPod(h, []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}, "android-sdk-pkg-update")
		installPodErr := sdk.Create(installPod)
		if installPodErr != nil {
			return installPodErr
		}
	}
	return nil
}

func (h *Handler) getPod(name string) (*corev1.Pod, error) {
	pods := h.k8c.CoreV1().Pods("android")

	return pods.Get(name, metav1.GetOptions{})
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


func getSdkPod(h *Handler, cmd []string, name string) *corev1.Pod {
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
