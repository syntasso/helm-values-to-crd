# Helm Values to CRD

A tool to generate a Kubernetes CRD from a Helm values style yaml file.

## Building

### Requirements:
- go 1.20

### Building binary

```
go build main.go -o helm-values-to-crd
```

## Example Usage

```bash
cat <<EOF >>test.yaml
image:
  repository: quay.io/spotahome/redis-operator
monitoring:
  enabled: false
  prometheus:
    name: ""
  serviceMonitor: false
  serviceAnnotations: {}
service:
  type: ClusterIP
  port: 9710
EOF

./helm-values-to-crd test.yaml redis.acme.org/v1alpha1
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: redis.acme.org
spec:
  group: acme.org
  names:
    kind: redis
    plural: redis
    singular: redis
  scope: Namespaced
  versions:
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        properties:
          spec:
            properties:
              image:
                properties:
                  repository:
                    type: string
                type: object
                x-kubernetes-preserve-unknown-fields: true
              monitoring:
                properties:
                  enabled:
                    type: boolean
                  prometheus:
                    properties:
                      name:
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  serviceAnnotations:
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  serviceMonitor:
                    type: boolean
                type: object
                x-kubernetes-preserve-unknown-fields: true
              service:
                properties:
                  port:
                    type: integer
                  type:
                    type: string
                type: object
                x-kubernetes-preserve-unknown-fields: true
            type: object
        type: object
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: null
  storedVersions: null
```
