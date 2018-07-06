# Andrdoid SDK Operator Poc

Android SDK operator poc that watches changes in a configmap object to install/remove Android SDK packages in a Persistent Volume.

|                 | Project Info  |
| --------------- | ------------- |
| License:        | Apache License, Version 2.0                      |
| Google Group:   | https://groups.google.com/forum/#!forum/aerogear |
| IRC             | [#aerogear](https://webchat.freenode.net/?channels=aerogear) channel in the [freenode](http://freenode.net/) network. |

## Building

```sh
$ operator-sdk generate k8s
$ operator-sdk build aerogear/android-sdk-operator:dev
```

The above command will result in a linux container image which can also be pushed to a container registry.

## Deployment

```sh
$ kubectl create -f deploy/rbac.yaml
$ kubectl create -f deploy/crd.yaml
$ kubectl create -f deploy/operator.yaml

# Create a configmap with desired Android SDK package config
$ kubectl create -f deploy/android-sdk-config.yaml

# Create the custom resource which will trigger the operator to sync the packages
$ kubectl create -f deploy/cr.yaml
```
