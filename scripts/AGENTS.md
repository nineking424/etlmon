<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# scripts

## Purpose

Build, deployment, and maintenance scripts for etlmon.

## Key Files

| File | Description |
|------|-------------|
| `build.sh` | Cross-platform build script |
| `install.sh` | Installation script |
| `etlmon-node.service` | systemd service unit for Node |

## For AI Agents

### Working In This Directory

- Scripts should be POSIX-compatible where possible
- Include error handling and usage messages
- systemd unit should follow best practices

### build.sh Example

```bash
#!/bin/bash
set -e

VERSION=${VERSION:-"dev"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

echo "Building etlmon ${VERSION}..."

# Build node daemon
echo "  Building etlmon-node..."
CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" -o bin/etlmon-node ./cmd/node

# Build UI
echo "  Building etlmon-ui..."
CGO_ENABLED=1 go build -ldflags "${LDFLAGS}" -o bin/etlmon-ui ./cmd/ui

echo "Done. Binaries in bin/"
```

### etlmon-node.service Example

```ini
[Unit]
Description=etlmon Node Daemon
After=network.target

[Service]
Type=simple
User=etlmon
Group=etlmon
ExecStart=/usr/local/bin/etlmon-node -c /etc/etlmon/node.yaml
Restart=on-failure
RestartSec=5

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/etlmon

[Install]
WantedBy=multi-user.target
```

### install.sh Example

```bash
#!/bin/bash
set -e

INSTALL_DIR=${INSTALL_DIR:-"/usr/local/bin"}
CONFIG_DIR=${CONFIG_DIR:-"/etc/etlmon"}
DATA_DIR=${DATA_DIR:-"/var/lib/etlmon"}

# Create directories
sudo mkdir -p "${CONFIG_DIR}" "${DATA_DIR}"

# Copy binaries
sudo cp bin/etlmon-node bin/etlmon-ui "${INSTALL_DIR}/"

# Copy example configs
sudo cp configs/*.yaml "${CONFIG_DIR}/"

# Create user
if ! id etlmon &>/dev/null; then
    sudo useradd -r -s /bin/false etlmon
fi

# Set permissions
sudo chown -R etlmon:etlmon "${DATA_DIR}"

# Install systemd service
sudo cp scripts/etlmon-node.service /etc/systemd/system/
sudo systemctl daemon-reload

echo "Installation complete. Edit ${CONFIG_DIR}/node.yaml and start with:"
echo "  sudo systemctl start etlmon-node"
```

<!-- MANUAL: -->
