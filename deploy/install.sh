#!/usr/bin/env bash
set -euo pipefail

# Install awagent as a systemd user service (no root required)
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/awagent"
SYSTEMD_USER_DIR="${HOME}/.config/systemd/user"

echo "==> Building awagent"
"${ROOT_DIR}/build.sh"

echo "==> Installing binary to ${INSTALL_DIR}"
mkdir -p "${INSTALL_DIR}"
cp "${ROOT_DIR}/bin/awagent" "${INSTALL_DIR}/awagent"
chmod +x "${INSTALL_DIR}/awagent"

echo "==> Creating config directory ${CONFIG_DIR}"
mkdir -p "${CONFIG_DIR}"

if [ ! -f "${CONFIG_DIR}/config.json" ]; then
  echo "==> Copying sample config"
  cp "${ROOT_DIR}/config.sample.json" "${CONFIG_DIR}/config.json"
  echo "    Edit ${CONFIG_DIR}/config.json to customize settings"
else
  echo "    Config already exists at ${CONFIG_DIR}/config.json"
fi

echo "==> Installing systemd user service"
mkdir -p "${SYSTEMD_USER_DIR}"

cat > "${SYSTEMD_USER_DIR}/awagent.service" <<EOF
[Unit]
Description=ActivityWatch Git Session Tracker
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/awagent --config ${CONFIG_DIR}/config.json
Restart=on-failure
RestartSec=10

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload

echo ""
echo "Installation complete!"
echo ""
echo "Next steps:"
echo "  1. Edit config: ${CONFIG_DIR}/config.json"
echo "  2. Ensure aw-watcher-window is running"
echo "  3. Start the service: systemctl --user start awagent"
echo "  4. Enable at login:  systemctl --user enable awagent"
echo "  5. Check status:     systemctl --user status awagent"
echo "  6. View logs:        journalctl --user -u awagent -f"
echo ""
