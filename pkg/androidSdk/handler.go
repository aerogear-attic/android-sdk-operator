package androidSdk

import (
	"context"

	"github.com/Sirupsen/logrus"
	api "github.com/aerogear/android-sdk-operator-poc/pkg/apis/androidsdk/v1"
	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"

	"k8s.io/client-go/kubernetes"
)

var encodedConfig = ""

const (
	installerPodName = "android-sdk-pkg-install"
	updaterPodName   = "android-sdk-pkg-update"
)

func NewHandler(k8c kubernetes.Interface) sdk.Handler {
	return &Handler{
		k8c: k8c,
	}
}

type Handler struct {
	k8c kubernetes.Interface
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	logrus.Info("handler is being called")

	switch o := event.Object.(type) {
	case *api.AndroidSDK:
		o = o.DeepCopy()

		if o.Status.Phase == api.Install {
			return installSdk(h, o)
		}

		if o.Status.Phase == api.Sync {
			return syncPackages(h, o)
		}

		if o.Status.Phase == api.Installing {
			return deletePod(h, o, installerPodName)
		}

		if o.Status.Phase == api.Syncing {
			return deletePod(h, o, updaterPodName)
		}

		if o.Status.Phase == api.Done {
			return watchChanges(h, o)
		}
	}
	return nil
}
