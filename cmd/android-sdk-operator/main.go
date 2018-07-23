package main

import (
	"context"
	"runtime"

	configHandler "github.com/aerogear/android-sdk-operator/pkg/androidSdk"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"

	"github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/sirupsen/logrus"
)

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
}

func main() {
	printVersion()

	resource := "androidsdk.aerogear.org/v1"
	kind := "AndroidSDK"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}

	client := k8sclient.GetKubeClient()
	kube := configHandler.NewKube(client)
	sdkHelper := configHandler.DefaultSdkHelper()
	handler := configHandler.NewHandler(&kube, &sdkHelper)
	resyncPeriod := 5

	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)

	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Handle(handler)
	sdk.Run(context.TODO())
}
