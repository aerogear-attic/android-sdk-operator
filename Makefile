ORG := aerogear
TAG := dev

deploy:
	oc create -f deploy/rbac.yaml && oc create -f deploy/crd.yaml && oc create -f deploy/operator.yaml

olm:
	oc apply -f extras/olm/

build:
	operator-sdk build docker.io/$(ORG)/android-sdk-operator:$(TAG)
