#!/usr/bin/env bash
set -euo pipefail

# Install awagent as a systemd user service (no root required)
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.config/awagent"
SYSTEMD_USER_DIR="${HOME}/.config/systemd/user"

# Resolve machine name: first CLI arg, then git global user.name, then primary non-loopback IP, then hostname
NAME=""
if [ "$#" -ge 1 ] && [ -n "${1:-}" ]; then
  NAME="$1"
else
  if command -v git >/dev/null 2>&1 ; then
    NAME=$(git config --global user.name || true)
  fi
  NAME=$(echo "$NAME" | tr -d '\r')
  if [ -z "$NAME" ]; then
    # Try to get primary non-loopback IP
    if command -v ip >/dev/null 2>&1; then
      NAME=$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src"){print $(i+1); exit}}')
    fi
    if [ -z "$NAME" ]; then
      NAME=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    if [ -z "$NAME" ]; then
      NAME=$(hostname)
    fi
  fi
fi

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

# Update machine name in config.json to the resolved NAME
if [ -n "$NAME" ]; then
  # Escape for sed
  esc=$(printf '%s' "$NAME" | sed -e 's/[\/&]/\\&/g')
  if command -v jq >/dev/null 2>&1; then
    tmp=$(mktemp)
    jq --arg m "$NAME" '.activityWatch.machine = $m' "${CONFIG_DIR}/config.json" > "$tmp" && mv "$tmp" "${CONFIG_DIR}/config.json"
  else
    # Fallback to sed replace of the machine field
    sed -i "s/\"machine\"[[:space:]]*:[[:space:]]*\"[^\"]*\"/\"machine\": \"${esc}\"/" "${CONFIG_DIR}/config.json" || true
  fi
  echo "==> Configured machine name: ${NAME}"
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
