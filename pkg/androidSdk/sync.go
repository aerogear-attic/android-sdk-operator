package androidSdk

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	api "github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// syncPackages updates the installed packages to match the desired config
func syncPackages(h *Handler, helper *api.AndroidSDK) error {
	logrus.Info("syncing packages..")

	helper.Status.Phase = api.Syncing
	err := sdk.Update(helper)
	if err != nil {
		return fmt.Errorf("error updating resource status: %v", err)
	}

	// Update the packages based on the new configmap
	cmd := []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}
	pod := h.getSdkPod(cmd, updaterPodName, helper.Spec.ConfigMapName, helper.Namespace)
	err = sdk.Create(pod)
	if err != nil {
		return fmt.Errorf("error creating pod: %v", err)
	}

	// Update the configmap hash
	b64, err := h.encodeData(helper.Spec.ConfigMapName, helper.Namespace)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}
	encodedConfig = b64

	return nil
}
