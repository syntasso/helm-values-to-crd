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
```
