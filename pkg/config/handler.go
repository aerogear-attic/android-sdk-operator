package config

import (
	"context"
	b64 "encoding/base64"
	"fmt"

	"github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	configHash = ""
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
	case *v1.AndroidSDK:
		ns := o.Namespace

		if o.Status.Status == v1.Install {
			o.Status.Status = v1.Installing
			err := sdk.Update(o)
			if err != nil {
				return fmt.Errorf("failed to update android sdk resource status: %v", err)
			}

			// Initialise the configmap hash to the b64 encoded value
			// of the configmap's data
			hash, err := h.getHash(o.Spec.ConfigRef, ns)
			if err != nil {
				return fmt.Errorf("unable to update hash for configmap data: %v", err)
			}
			configHash = hash

			// Install the sdk
			name := "android-sdk-pkg-install"
			cmd := []string{"androidctl", "sdk", "install"}
			pod := getSdkPod(cmd, name, o.Spec.ConfigRef, ns)

			err = sdk.Create(pod)
			if err != nil {
				return fmt.Errorf("failed to create sdk installer pod: %v", err)
			}
		}

		if o.Status.Status == v1.Sync {
			o.Status.Status = v1.Syncing
			err := sdk.Update(o)
			if err != nil {
				return fmt.Errorf("failed to update android sdk resource status: %v", err)
			}

			// Update the packages based on the new configmap
			name := "android-sdk-pkg-update"
			cmd := []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}
			pod := getSdkPod(cmd, name, o.Spec.ConfigRef, ns)
			err = sdk.Create(pod)
			if err != nil {
				return fmt.Errorf("failed to create sdk installer pod: %v", err)
			}

			// Update the configmap hash
			hash, err := h.getHash(o.Spec.ConfigRef, ns)
			if err != nil {
				return fmt.Errorf("unable to update hash for configmap data: %v", err)
			}
			configHash = hash
		}

		if o.Status.Status == v1.Done {
			// Verify if the global hash and the current hash are the same
			hash, err := h.getHash(o.Spec.ConfigRef, ns)
			if err != nil {
				return fmt.Errorf("unable to get hash for configmap data: %v", err)
			}

			// Set status to sync if a change is detected
			if hash != configHash {
				logrus.Info("Change in configmap detected")

				o.Status.Status = v1.Sync
				err = sdk.Update(o)
				if err != nil {
					return fmt.Errorf("failed to update android sdk resource status: %v", err)
				}
			}
		}

		if o.Status.Status == v1.Syncing {
			name := "android-sdk-pkg-update"
			pod, err := h.getPod(name, ns)
			if err != nil {
				return err
			}

			// Delete sdk updater pod if completed
			// TODO better/cleaner way to do this?
			if pod.Status.Phase == "Succeeded" {
				logrus.Infof("updater pod finished running, starting its removal..")

				err := sdk.Delete(getSdkPod([]string{""}, name, o.Spec.ConfigRef, ns), sdk.WithDeleteOptions(&metav1.DeleteOptions{}))
				if err != nil {
					return fmt.Errorf("error while deleting pod: %v", err)
				}

				// Set status to 'done'
				o.Status.Status = v1.Done
				err = sdk.Update(o)
				if err != nil {
					return fmt.Errorf("failed to update android sdk resource status: %v", err)
				}
			}
		}

		if o.Status.Status == v1.Installing {
			name := "android-sdk-pkg-install"
			pod, err := h.getPod(name, ns)
			if err != nil {
				return err
			}

			// Delete sdk installer pod if completed
			// TODO better/cleaner way to do this?
			if pod.Status.Phase == "Succeeded" {
				logrus.Infof("installer pod finished running, starting its removal..")

				err := sdk.Delete(getSdkPod([]string{""}, name, o.Spec.ConfigRef, ns), sdk.WithDeleteOptions(&metav1.DeleteOptions{}))
				if err != nil {
					return fmt.Errorf("error while deleting pod: %v", err)
				}

				o.Status.Status = v1.Done
				err = sdk.Update(o)
				if err != nil {
					return fmt.Errorf("failed to update android sdk resource status: %v", err)
				}
			}
		}
	}
	return nil
}

func (h *Handler) getPod(name string, ns string) (*corev1.Pod, error) {
	pods := h.k8c.CoreV1().Pods(ns)
	return pods.Get(name, metav1.GetOptions{})
}

// getHash gets the hash of the configmap's data
func (h *Handler) getHash(name string, ns string) (string, error) {
	configMap, _ := h.k8c.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	config, err := v1.GetConfigData(configMap)
	if err != nil {
		return "", err
	}
	return b64.StdEncoding.EncodeToString([]byte(config)), nil
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

func getSdkPod(cmd []string, name string, configMap string, ns string) *corev1.Pod {
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
								Name: configMap,
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
