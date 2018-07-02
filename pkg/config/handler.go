package config

import (
	"context"
	"fmt"

	"github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func NewHandler(k8c kubernetes.Interface) sdk.Handler {
	return &Handler{
		k8c: k8c,
	}
}

type Handler struct {
	k8c  kubernetes.Interface
	Data string
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	logrus.Infof("Handler is being called")

	switch o := event.Object.(type) {
	case *corev1.ConfigMap:
		ns := o.Namespace

		// Ignore invalid config maps
		isValid := v1.IsValidSdkConfig(o)
		if !isValid {
			return nil
		}

		config, err := v1.GetConfigData(o)
		if err != nil {
			return fmt.Errorf("failed to get config: %v", err)
		}

		// Create the custom resource if it doesn't already exist
		resource := updateSdkResource(config, ns)
		err = sdk.Create(resource)
		if err != nil && !kerrors.IsAlreadyExists(err) {
			return fmt.Errorf("failed to create sdk resource: %v", err)
		}

		// TODO: there is probably a better way to delete the pod
		// Cleans up the completed pod
		name := "android-sdk-pkg-update"
		err = h.cleanUp(name, ns)
		if err != nil {
			return fmt.Errorf("failed to clean up installer pod: %v", err)
		}

		// TODO: persist config status in AndroidSDK CRD
		// TODO: Maybe we should store a hash string based on the config map content and only run the update if the hash is not a match
		cmd := []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}
		pod := getSdkPod(cmd, name, ns)
		err = sdk.Create(pod)
		if err != nil {
			return fmt.Errorf("failed to create sdk installer pod: %v", err)
		}
	}
	return nil
}

func (h *Handler) getPod(name string, ns string) (*corev1.Pod, error) {
	pods := h.k8c.CoreV1().Pods(ns)
	return pods.Get(name, metav1.GetOptions{})
}

func (h *Handler) cleanUp(name string, ns string) error {
	pod, err := h.getPod(name, ns)
	if err == nil {
		if pod.Status.Phase == "Succeeded" {
			logrus.Infof("Android SDK Pod finished running, starting its removal.")

			err = sdk.Delete(getSdkPod([]string{""}, name, ns), sdk.WithDeleteOptions(&metav1.DeleteOptions{}))
			if err != nil {
				logrus.Infof("Error while deleting pod %s.", pod.Name)
				return err
			}
			return nil
		}
		return nil
	}
	return err
}

func updateSdkResource(cfg string, ns string) *v1.AndroidSDK {
	androidSdk := &v1.AndroidSDK{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AndroidSDK",
			APIVersion: "androidsdk.aerogear.org/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "android-sdk-config-object",
			Namespace: ns,
		},
		Spec: v1.AndroidSDKSpec{
			Data: cfg,
		},
		Status: v1.AndroidSDKStatus{
			Status: v1.Done,
		},
	}

	return androidSdk
}

func getSdkPod(cmd []string, name string, ns string) *corev1.Pod {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "android-sdk-pkg",
					Image:   "docker.io/aerogear/digger-android-sdk-image:1.0.0-alpha",
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
