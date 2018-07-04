package androidSdk

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	api "github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deletePod cleans up suceeded pods and sets the resource's status to "Done"
// once complete
func deletePod(h *Handler, helper *api.AndroidSDK, podName string) error {
	pod, err := h.getPod(podName, helper.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("error getting pod: %v", err)
	}

	// TODO better/cleaner way to do this?
	if pod.Status.Phase == "Succeeded" {
		logrus.Infof("pod finished running, starting removal..")

		opts := sdk.WithDeleteOptions(&metav1.DeleteOptions{})
		err := sdk.Delete(h.getSdkPod([]string{""}, podName, helper.Spec.ConfigMapName, helper.Namespace), opts)
		if err != nil {
			return fmt.Errorf("error while deleting pod: %v", err)
		}

		helper.Status.Phase = api.Done
		err = sdk.Update(helper)
		if err != nil {
			return fmt.Errorf("error updating resource status: %v", err)
		}
	}

	return nil
}

// watchChanges compares whether the configmap has been changed or not
// and sets the resource's status to "Sync" if it has
func watchChanges(h *Handler, helper *api.AndroidSDK) error {
	b64, err := h.encodeData(helper.Spec.ConfigMapName, helper.Namespace)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}

	// Set status to "Sync" if a change is detected
	if b64 != encodedConfig {
		logrus.Info("change in configmap detected")

		helper.Status.Phase = api.Sync
		err = sdk.Update(helper)
		if err != nil {
			return fmt.Errorf("error updating resource status: %v", err)
		}
	}

	return nil
}
