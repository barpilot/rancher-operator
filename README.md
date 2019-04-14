# rancher-operator

_rancher-operator_ aim to provide some new features not include directly in product.

## Features

### AutoProject

_AutoProject_ add a new [project](https://rancher.com/docs/rancher/v2.x/en/project-admin/) in each cluster that your Rancher instance manage.

```yaml
apiVersion: rancheroperator.barpilot.io/v1alpha1
kind: AutoProject
metadata:
  name: internal-ops
spec:
  projectSpec:
    displayName: Internal-Ops
    description: Project used by the ops team to give you the best kubernetes UX
```

### AutoMultiClusterApp

_AutoMultiClusterApp_ inject [Multi-Cluster App](https://rancher.com/docs/rancher/v2.x/en/catalog/multi-cluster-apps/) in a project (based on label selector).

```yaml
apiVersion: rancheroperator.barpilot.io/v1alpha1
kind: AutoMultiClusterApp
metadata:
  name: cert-manager
spec:
  multiClusterApp: cert-manager
  projectSelector: "autoproject/displayname==Internal-Ops"
```

Multi-Cluster App should already exists.

## Status: _ALPHA_

Use it after tests and coffee.

## Prerequisites

_rancher-operator_ should be deployed in the *same* kubebernetes cluster where _Rancher_ is deployed (_local_).

## Use-Cases

### For a KaaS team

A Kubernetes as a Service Team can add some default features to a cluster:
- log
- monitoring
- ingress
  - externalDNS
  - cert-manager

This add value to user with default "working" configuration.
