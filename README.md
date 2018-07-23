# Andrdoid SDK Operator Poc

Android SDK operator that watches changes in a configmap object to install/remove Android SDK packages in a Persistent Volume.

|                 | Project Info  |
| --------------- | ------------- |
| License:        | Apache License, Version 2.0                      |
| Google Group:   | https://groups.google.com/forum/#!forum/aerogear |
| IRC             | [#aerogear](https://webchat.freenode.net/?channels=aerogear) channel in the [freenode](http://freenode.net/) network. |

## Building

```sh
$ operator-sdk generate k8s
$ operator-sdk build quay.io/aerogear/android-sdk-operator:dev
```

The above command will result in a linux container image which can also be pushed to an external container registry.

## Deployment

Deploying the operator and related resources:

```sh
#Deploy the android-sdk persistent volume
oc new-app -f extras/android/android-persistent.json

#Deploy a configmap with desired Android SDK package config
$ oc create -f extras/android/android-sdk-config.yaml

#Deploy the required resource definitions
$ oc create -f deploy/rbac.yaml
$ oc create -f deploy/crd.yaml
$ oc create -f deploy-cr.yaml

#Deploy the operator itself
$ oc create -f deploy/operator.yaml
```
