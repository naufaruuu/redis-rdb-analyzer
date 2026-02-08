# Redis RDB Analyzer for Kubernetes

A specialized Redis RDB (dump file) analysis tool with Kubernetes integration, built for analyzing Redis memory usage and key distribution in production environments.

![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)

## Overview

This tool provides a modern web interface for analyzing Redis RDB files, with special focus on:
- **Kubernetes Integration**: Auto-discovery and direct import from Redis pods
- **Large File Support**: Handles multi-gigabyte RDB files reliably (tested with 10GB+ files)
- **Modern UI**: Built with Tailwind CSS, featuring dark mode and interactive charts
- **Async Processing**: Job-based system for analyzing large files without blocking
- **Historical Analysis**: SQLite-based storage for tracking changes over time

### 🛡️ Safe & Non-Intrusive

**Production-safe:** Analyzes offline RDB snapshot files only. Never connects to live Redis, no performance impact, zero downtime required. Uses `kubectl exec` to copy files for local analysis.

## Features

- 🔍 **BigKey Detection**: Identify memory-intensive keys
- 📊 **Prefix Analysis**: Group and analyze keys by common prefixes
- 📈 **Distribution Charts**: Visualize key types, sizes, and expiration patterns
- ☸️ **K8s Native**: Import RDB files directly from Redis pods via kubectl
- 🚀 **High Performance**: Stream-based parsing handles large files efficiently
- 🌙 **Modern UI**: Responsive design with dark mode support
- 📜 **History Tracking**: Compare analyses over time

## Attribution & Credits

Derivative work based on **[919927181/rdr](https://github.com/919927181/rdr)** (Apache 2.0) and **[xueqiu/rdr](https://github.com/xueqiu/rdr)** (Apache 2.0). Uses **[HDT3213/rdb](https://github.com/HDT3213/rdb)** (MIT) for reliable parsing.

**Major changes in v2.0:** Kubernetes integration, modern UI (Tailwind CSS), async job system, SQLite persistence, and upgraded parser. Removed CLI modes (web-only now) and manual file upload.

## Installation

### Prerequisites
- Go 1.18+
- `kubectl` configured with access to Redis pods (for K8s import feature)

### Build from Source

```bash
git clone <your-repo-url>
cd redis_rdb_analyzer

# Using Make (recommended)
make build

# Or manually
go build -o redis-rdb-analyzer main.go
```

### 🐳 Docker Deployment

```bash
docker build -t redis-rdb-analyzer:v1.0 .
docker run -d -p 8080:8080 -v ~/.kube:/home/rdr/.kube:ro redis-rdb-analyzer:v1.0
```

📖 See [DOCKER.md](DOCKER.md) for full deployment guide with RBAC and production configuration.

### ☸️ Helm Chart Deployment (Recommended)

```bash
helm install redis-rdb-analyzer chart/ -n tools --create-namespace

# With custom configuration
helm install redis-rdb-analyzer chart/ \
  --set config.pod_cache=30m \
  --set config.max_rdb_size=50Gb \
  --set ingress.enabled=true \
  -n tools
```

Includes: StatefulSet with PVC, RBAC, ConfigMap, Service, optional Ingress, probes, resource limits.

📖 See [chart/README.md](chart/README.md) and [HELM_DEPLOYMENT.md](HELM_DEPLOYMENT.md) for full documentation.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `RDR_PORT` | `8080` | Web server port |
| `POD_CACHE_DURATION` | `15m` | K8s pod discovery cache duration (e.g., `15m`, `1h`) |
| `MAX_RDB_SIZE` | `10Gb` | Max RDB file size (e.g., `10Gb`, `500Mb`) |

**Local development:**
```bash
cp .env.example .env  # Edit as needed
make run-dev
```

**Kubernetes:** Set via `chart/values.yaml`:
```yaml
config:
  port: 8080
  pod_cache: 30m
  max_rdb_size: 50Gb
```

📖 See [ENV_SETUP.md](ENV_SETUP.md) and [CONFIG_TESTING.md](CONFIG_TESTING.md) for details.

## Usage

```bash
./redis-rdb-analyzer -p 8080  # Or use: make run-dev
```

Open browser to `http://localhost:8080`

**Kubernetes Import:**
1. Auto-discovers Redis pods via `kubectl`
2. Select namespace/pod from dashboard
3. Click "Import RDB" (default path: `/data/dump.rdb`)
4. Analysis runs asynchronously with progress tracking

**Note:** Requires `kubectl` access. For local files without K8s, use [919927181/rdr](https://github.com/919927181/rdr) instead.

## Project Structure

```
redis_rdb_analyzer/
├── main.go              # Entry point
├── decoder/             # RDB parsing logic
│   ├── decoder.go       # Core data structures
│   ├── hdt_adapter.go   # Adapter for HDT3213 parser
│   ├── hdt_decode.go    # Parsing implementation
│   └── memprofiler.go   # Memory estimation
├── server/                # Web server & analysis
│   ├── show.go          # HTTP handlers & routes
│   ├── job.go           # Async job manager
│   ├── db.go            # SQLite persistence
│   ├── counter.go       # Statistical aggregation
│   ├── k8s_discovery.go # Kubernetes integration
│   └── ...
├── views/               # HTML templates (Tailwind CSS)
│   ├── dashboard.html
│   ├── layout.html
│   └── ...
└── data/
    └── rdr.db          # SQLite database (created at runtime)
```

## Development

```bash
make run-dev         # Build and run with .env
make build           # Build binary
make test            # Run tests
make lint            # Run linters
make docker-build    # Build Docker image
```

**Hot reload:** Edit `.html` files in `views/` - changes reflected on refresh (no rebuild).

**Reset database:** `rm data/rdr.db`

See [CONTRIBUTING.md](CONTRIBUTING.md) for full development guide.

## License

Apache License 2.0 - see [LICENSE](LICENSE) and [NOTICE](NOTICE) files.

**Third-party:** HDT3213/rdb (MIT), 919927181/rdr (Apache 2.0), xueqiu/rdr (Apache 2.0).

## Acknowledgments

Special thanks to:
- **919927181** and the RDR contributors for the solid foundation
- **HDT3213** for the reliable RDB parser
- **xueqiu** team for the original RDR concept
- The Redis community for tools and documentation
- **Claude Code** and **Google Antigravity** for their assistance in making this project happen

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Support

For issues or questions, please open an issue on GitHub.
