package androidSdk

import (
	"context"

	"github.com/sirupsen/logrus"
	api "github.com/aerogear/android-sdk-operator/pkg/apis/androidsdk/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func NewHandler(kube HandlerKube, helper HandlerSdk) sdk.Handler {
	logrus.Info("---initializing handler---")
	return &Handler{
		kube: kube,
		sdk: helper,
		encodedConfig: "",
	}
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *api.AndroidSDK:
		o = o.DeepCopy()

		if o.Status.Phase == api.Install {
			return h.installSdk(o)
		}

		if o.Status.Phase == api.Sync {
			return h.syncPackages(o)
		}

		if o.Status.Phase == api.Installing {
			return h.deletePod(o, InstallerPodName)
		}

		if o.Status.Phase == api.Syncing {
			return h.deletePod(o, UpdaterPodName)
		}

		if o.Status.Phase == api.Done {
			return h.watchChanges(o)
		}

		if o.Status.Phase == "" {
			return h.update(o)
		}
	}

	return nil
}

func (h *Handler) installSdk(o *api.AndroidSDK) error {
	logrus.Info("installing Android SDK")

	o.Status.Phase = api.Installing
	err := sdk.Update(o)
	if err != nil {
		return fmt.Errorf("error updating resource status: %v", err)
	}

	configMap, err := h.kube.GetConfigMap(o.Spec.ConfigMapName, o.Namespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve configmap: %v", err)
	}

	// Initialise the encodedConfig variable to the b64 encoded value of the configmap's data.
	// This will be used to detect whether the confimap has been updated or not
	b64, err := h.sdk.EncodeData(configMap)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}
	h.encodedConfig = b64

	// Install the sdk
	cmd := []string{"androidctl", "sdk", "install"}
	pod := h.sdk.GetSdkPod(cmd, InstallerPodName, o.Spec.ConfigMapName, o.Namespace)

	err = sdk.Create(pod)
	if err != nil {
		return fmt.Errorf("failed to create sdk installer pod: %v", err)
	}

	return nil
}

func (h *Handler) syncPackages(o *api.AndroidSDK) error {
	logrus.Info("syncing packages..")

	o.Status.Phase = api.Syncing
	err := sdk.Update(o)
	if err != nil {
		return fmt.Errorf("error updating resource status: %v", err)
	}

	// Update the packages based on the new configmap
	cmd := []string{"/opt/tools/androidctl-sync", "-y", "/tmp/android-sdk-config/packages"}
	pod := h.sdk.GetSdkPod(cmd, UpdaterPodName, o.Spec.ConfigMapName, o.Namespace)
	err = sdk.Create(pod)
	if err != nil {
		return fmt.Errorf("error creating pod: %v", err)
	}

	configMap, err := h.kube.GetConfigMap(o.Spec.ConfigMapName, o.Namespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve configmap: %v", err)
	}

	// Update the configmap hash
	b64, err := h.sdk.EncodeData(configMap)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}
	h.encodedConfig = b64

	return nil
}

func (h *Handler) deletePod(o *api.AndroidSDK, name string) error {
	pod, err := h.kube.GetPod(name, o.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("error getting pod: %v", err)
	}

	if pod.Status.Phase != "Succeeded" {
		return nil
	}

	// TODO better/cleaner way to do this?
	logrus.Infof("pod finished running, starting removal..")

	opts := sdk.WithDeleteOptions(&metav1.DeleteOptions{})
	err = sdk.Delete(h.sdk.GetSdkPod([]string{""}, name, o.Spec.ConfigMapName, o.Namespace), opts)
	if err != nil {
		return fmt.Errorf("error while deleting pod: %v", err)
	}

	if h.sdk.IsInstallerPod(name) {
		// If the Android SDK has finished installing, start syncing packages
		o.Status.Phase = api.Sync
		err = sdk.Update(o)
	} else {
		o.Status.Phase = api.Done
		err = sdk.Update(o)
	}

	if err != nil {
		return fmt.Errorf("error updating resource status: %v", err)
	}

	return nil
}

func (h *Handler) watchChanges(o *api.AndroidSDK) error {
	configMap, err := h.kube.GetConfigMap(o.Spec.ConfigMapName, o.Namespace)
	if err != nil {
		return fmt.Errorf("unable to retrieve configmap: %v", err)
	}

	b64, err := h.sdk.EncodeData(configMap)
	if err != nil {
		return fmt.Errorf("error encoding configmap data: %v", err)
	}

	// Set status to "Sync" if a change is detected
	if b64 != h.encodedConfig {
		logrus.Info("change in configmap detected")

		o.Status.Phase = api.Sync

		err = sdk.Update(o)
		if err != nil {
			return fmt.Errorf("error updating resource status: %v", err)
		}
	}

	return nil
}

func (h *Handler) update(o *api.AndroidSDK) error {
	o.Status.Phase = api.Install
	return sdk.Update(o)
}
