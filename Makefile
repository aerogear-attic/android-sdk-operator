ORG := aerogear
TAG := dev
CMD := oc

bootstrap:
	$(CMD) new-app -f extras/android/android-persistent.json

deploy-android-config:
	$(CMD) create -f extras/android/android-sdk-config.yml

deploy-rbac:
	$(CMD) create -f deploy/rbac.yaml

deploy-crd:
	$(CMD) create -f deploy/crd.yaml

deploy-cr:
	$(CMD) create -f deploy/cr.yaml

deploy-operator:
	$(CMD) create -f deploy/operator.yaml

deploy: deploy-android-config deploy-rbac deploy-crd deploy-cr deploy-operator

build:
	operator-sdk build docker.io/$(ORG)/android-sdk-operator:$(TAG)
