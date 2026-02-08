# Redis RDB Analyzer Helm Chart

A Kubernetes Helm chart for deploying Redis RDB Analyzer with StatefulSet, persistent storage, RBAC, and ingress support.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- PersistentVolume provisioner support in the underlying infrastructure
- kubectl access configured for the cluster

## Installing the Chart

### Quick Install

```bash
# Install with default values
helm install redis-rdb-analyzer ./chart

# Install in specific namespace
helm install redis-rdb-analyzer ./chart -n tools --create-namespace
```

### Install with Custom Values

```bash
# Using values file
helm install redis-rdb-analyzer ./chart -f custom-values.yaml

# Override specific values
helm install redis-rdb-analyzer ./chart \
  --set image.tag=v1.0 \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=rdb.example.com
```

## Configuration

The following table lists the configurable parameters of the Redis RDB Analyzer chart and their default values.

### Image Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Image repository | `redis-rdb-analyzer` |
| `image.tag` | Image tag | `v1.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets | `[]` |

### Deployment Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `podAnnotations` | Pod annotations | `{}` |
| `podSecurityContext` | Pod security context | See values.yaml |
| `securityContext` | Container security context | See values.yaml |

### Service Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Service type | `ClusterIP` |
| `service.port` | Service port | `80` |
| `service.targetPort` | Container port | `8080` |
| `service.annotations` | Service annotations | `{}` |

### Ingress Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.annotations` | Ingress annotations | See values.yaml |
| `ingress.hosts` | Ingress hosts | `rdb-analyzer.example.com` |
| `ingress.tls` | Ingress TLS configuration | See values.yaml |

### Persistence Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `persistence.enabled` | Enable persistent storage | `true` |
| `persistence.storageClass` | Storage class | `""` (default) |
| `persistence.accessMode` | Access mode | `ReadWriteOnce` |
| `persistence.size` | Volume size | `10Gi` |
| `persistence.existingClaim` | Use existing PVC | `""` |

### Resources

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.limits.cpu` | CPU limit | `1000m` |
| `resources.limits.memory` | Memory limit | `2Gi` |
| `resources.requests.cpu` | CPU request | `200m` |
| `resources.requests.memory` | Memory request | `512Mi` |

### Health Checks

| Parameter | Description | Default |
|-----------|-------------|---------|
| `livenessProbe` | Liveness probe configuration | See values.yaml |
| `readinessProbe` | Readiness probe configuration | See values.yaml |

### Node Selection & Tolerations

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Affinity rules | `{}` |

### RBAC

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rbac.create` | Create RBAC resources | `true` |
| `rbac.rules` | RBAC rules for kubectl access | See values.yaml |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` (auto-generated) |

## Example Configurations

### Production with Ingress

```yaml
# prod-values.yaml
replicaCount: 1

image:
  tag: v1.0
  pullPolicy: Always

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 500m
    memory: 1Gi

persistence:
  enabled: true
  size: 20Gi
  storageClass: "fast-ssd"

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
  hosts:
    - host: rdb-analyzer.production.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: rdb-analyzer-tls
      hosts:
        - rdb-analyzer.production.example.com

nodeSelector:
  node-role.kubernetes.io/worker: "true"
  kubernetes.io/arch: amd64

tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "tools"
    effect: "NoSchedule"
```

Deploy:
```bash
helm install redis-rdb-analyzer ./chart -f prod-values.yaml -n tools
```

### Development with NodePort

```yaml
# dev-values.yaml
replicaCount: 1

image:
  tag: latest
  pullPolicy: Always

resources:
  limits:
    cpu: 500m
    memory: 1Gi
  requests:
    cpu: 100m
    memory: 256Mi

persistence:
  enabled: true
  size: 5Gi

service:
  type: NodePort
  port: 80
```

Deploy:
```bash
helm install redis-rdb-analyzer ./chart -f dev-values.yaml -n dev
```

### With Specific Node Selection

```yaml
nodeSelector:
  kubernetes.io/arch: amd64
  node-role.kubernetes.io/worker: "true"
  node-type: tools

tolerations:
  - key: "tools-only"
    operator: "Equal"
    value: "true"
    effect: "NoSchedule"
  - key: "node.kubernetes.io/not-ready"
    operator: "Exists"
    effect: "NoExecute"
    tolerationSeconds: 300
```

## Upgrading

```bash
# Upgrade with new values
helm upgrade redis-rdb-analyzer ./chart -f new-values.yaml

# Upgrade with specific parameters
helm upgrade redis-rdb-analyzer ./chart --set image.tag=v1.1
```

## Uninstalling

```bash
# Uninstall the release
helm uninstall redis-rdb-analyzer -n tools

# Clean up PVCs (if needed)
kubectl delete pvc -l app.kubernetes.io/name=redis-rdb-analyzer -n tools
```

## Verifying the Deployment

```bash
# Check StatefulSet status
kubectl get statefulset -n tools

# Check pods
kubectl get pods -n tools -l app.kubernetes.io/name=redis-rdb-analyzer

# Check PVC
kubectl get pvc -n tools

# Check service
kubectl get svc -n tools

# View logs
kubectl logs -n tools -l app.kubernetes.io/name=redis-rdb-analyzer -f

# Port forward for local access
kubectl port-forward -n tools svc/redis-rdb-analyzer 8080:80
```

## Troubleshooting

### Pod not starting

```bash
# Check pod status
kubectl describe pod -n tools redis-rdb-analyzer-0

# Check events
kubectl get events -n tools --sort-by='.lastTimestamp'
```

### PVC not binding

```bash
# Check PVC status
kubectl describe pvc -n tools data-redis-rdb-analyzer-0

# Check available PVs
kubectl get pv
```

### RBAC issues

```bash
# Check service account
kubectl get sa -n tools

# Check cluster role binding
kubectl get clusterrolebinding | grep redis-rdb-analyzer

# Test kubectl access from pod
kubectl exec -n tools redis-rdb-analyzer-0 -- kubectl get pods -A
```

### Ingress not working

```bash
# Check ingress
kubectl describe ingress -n tools redis-rdb-analyzer

# Check ingress controller logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx
```

## Notes

- The chart uses StatefulSet for stable network identity and persistent storage
- Each replica gets its own PVC for data persistence
- RBAC is configured for kubectl access to discover and exec into Redis pods
- The application runs as non-root user (UID 1000)
- Health checks are configured with sensible defaults
- Use `nodeSelector` and `tolerations` to control pod placement

## License

Apache 2.0
