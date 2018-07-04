package androidSdk

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	api "github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func installSdk(h *Handler, helper *api.AndroidSDK) error {
	logrus.Info("installing Android SDK")

	helper.Status.Phase = api.Installing
	err := sdk.Update(helper)
	if err != nil {
		return fmt.Errorf("error updating resource status: %v", err)
	}

	// Initialise the encodedConfig variable to the b64 encoded value of the configmap's data.
	// This will be used to detect whether the confimap has been updated or not
	b64, err := h.encodeData(helper.Spec.ConfigMapName, helper.Namespace)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}
	encodedConfig = b64

	// Install the sdk
	cmd := []string{"androidctl", "sdk", "install"}
	pod := h.getSdkPod(cmd, installerPodName, helper.Spec.ConfigMapName, helper.Namespace)

	err = sdk.Create(pod)
	if err != nil {
		return fmt.Errorf("failed to create sdk installer pod: %v", err)
	}
	return nil
}
