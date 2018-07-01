package v1

import(
	"errors"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
)

func IsValidSdkConfig(configMap *corev1.ConfigMap) bool {
	labels := configMap.Labels
	expectedAppLabel := "android-sdk-persistent"
	expectedCtxLabel := "android-sdk-config"
	appLabel, ok := labels["app"]
	if !ok {
		return false
	}

	ctxLabel, ok := labels["context"]
	if !ok {
		return false
	}

	if expectedAppLabel == appLabel && expectedCtxLabel == ctxLabel {
		return true
	}

	return false
}

func GetConfigData(configMap *corev1.ConfigMap) (string, error) {
	data, ok := configMap.Data["packages"]

	if !ok {
		return "", kerrors.NewInternalError(errors.New("Configmap \"packages\" key not found."))
	}

	return data, nil
}
