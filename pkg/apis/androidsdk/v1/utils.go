package v1

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
)

func GetConfigData(configMap *corev1.ConfigMap) (string, error) {
	data, ok := configMap.Data["packages"]

	if !ok {
		return "", errors.New("\"packages\" key not found in configmap")
	}

	return data, nil
}
