package stub

import (
	"context"
	"errors"

	//"github.com/aerogear/android-sdk-operator/pkg/apis/androidsdk/v1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"fmt"
	"strings"
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
		isValid := isValidSdkConfig(o)
		if !isValid {
			logrus.Infof("ConfigMap %s is not a valid android-sdk-config object", o.Name)
			return nil
		}

		cfgStr, cfgStrErr := getConfigData(o)
		if cfgStrErr != nil {
			return cfgStrErr
		}

		if !h.hasConfigChanged(cfgStr) {
			logrus.Info("Config string did not change.")
			return nil
		}

		logrus.Infof("Config string has changed, updating package cfg definition...")
		h.updateConfig(cfgStr)
		logrus.Infof("Package definition changed")

		//TODO: need to persist status for pod execution

		cmd := []string{"/opt/tools/write-to-file.sh", fmt.Sprintf("'%s'", strings.Replace(cfgStr, "\n", "\\n", -1)), "/opt/android-sdk-linux/sdk.cfg"}
		updatePod := runSdkPod(h, cmd, "android-sdk-config-update")
		updatePodErr := sdk.Create(updatePod)
		if updatePodErr != nil {
			return updatePodErr
		}

		installPod := runSdkPod(h, []string{"/opt/tools/androidctl-sync", "-y", "/opt/android-sdk-linux/sdk.cfg"}, "android-sdk-pkg-update")
		installPodErr := sdk.Create(installPod)
		if installPodErr != nil {
			return installPodErr
		}
	}
	return nil
}

func isValidSdkConfig(configMap *corev1.ConfigMap) bool {
	labels := configMap.Labels
	expectedAppLabel := "android-sdk-persistent"
	expectedCtxLabel := "android-sdk-config"
	appLabel, appLabelOk := labels["app"]
	ctxLabel, ctxLabelOk := labels["context"]

	if appLabelOk && ctxLabelOk {
		if expectedAppLabel == appLabel && expectedCtxLabel == ctxLabel {
			return true
		}
	}

	return false
}

func getConfigData(configMap *corev1.ConfigMap) (string, error) {
	data, ok := configMap.Data["packages"]

	if !ok {
		return "", kerrors.NewInternalError(errors.New("Configmap \"packages\" key not found."))
	}

	return data, nil
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
			},
			RestartPolicy: "Never",
		},
	}

	return pod
}

func (h *Handler) hasConfigChanged(cfg string) bool {
	if len(h.Data) == 0 {
		return true
	}

	return h.Data != cfg
}

func (h *Handler) updateConfig(cfg string) {
	h.Data = cfg
}
