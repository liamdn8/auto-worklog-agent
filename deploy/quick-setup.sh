#!/bin/bash
# Quick deployment script for Linux desktop

set -e

echo "=== Auto Worklog Agent - Quick Setup ==="
echo ""

# Check if running on Linux
if [ "$(uname)" != "Linux" ]; then
    echo "Error: This script is for Linux only"
    exit 1
fi

# 1. Check for window detection tools
echo "Step 1: Checking window detection tools..."
HAS_TOOL=false
for tool in xdotool xprop wmctrl gdbus qdbus; do
    if command -v $tool >/dev/null 2>&1; then
        echo "✓ Found: $tool"
        HAS_TOOL=true
    fi
done

if [ "$HAS_TOOL" = false ]; then
    echo "⚠ No window detection tool found!"
    echo "  Please install one of: xdotool (recommended), xprop, or wmctrl"
    echo ""
    echo "  Ubuntu/Debian: sudo apt-get install xdotool"
    echo "  Fedora/RHEL:   sudo dnf install xdotool"
    echo "  Arch Linux:    sudo pacman -S xdotool"
    echo ""
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 2. Create config directory
echo ""
echo "Step 2: Creating config directory..."
mkdir -p ~/.config/awagent
echo "✓ Config directory created: ~/.config/awagent"

# 3. Copy config file
echo ""
echo "Step 3: Setting up configuration..."
if [ ! -f ~/.config/awagent/config.json ]; then
    cp config.sample.json ~/.config/awagent/config.json
    # Replace /home/liamdn with actual home
    sed -i "s|/home/liamdn|$HOME|g" ~/.config/awagent/config.json
    echo "✓ Config file created: ~/.config/awagent/config.json"
    echo "  Edit this file to customize your settings"
else
    echo "✓ Config file already exists"
fi

# 4. Make binary executable
echo ""
echo "Step 4: Making binary executable..."
chmod +x awagent
echo "✓ Binary is executable"

# 5. Check ActivityWatch
echo ""
echo "Step 5: Checking ActivityWatch server..."
if curl -s http://localhost:5600/api/0/info >/dev/null 2>&1; then
    echo "✓ ActivityWatch server is running"
else
    echo "⚠ ActivityWatch server not detected at http://localhost:5600"
    echo "  You need to start ActivityWatch server first!"
    echo ""
    echo "  Option 1 - Using Docker:"
    echo "    docker run -d --name activitywatch \\"
    echo "      -p 5600:5600 \\"
    echo "      -v ~/.local/share/activitywatch:/data \\"
    echo "      activitywatch/activitywatch:latest"
    echo ""
    echo "  Option 2 - Install ActivityWatch native:"
    echo "    https://activitywatch.net/downloads/"
fi

# 6. Test the agent
echo ""
echo "Step 6: Testing the agent..."
echo "Running agent for 5 seconds..."
timeout 5 ./awagent --verbose 2>&1 | head -20 || true

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "Next steps:"
echo "  1. Review config: nano ~/.config/awagent/config.json"
echo "  2. Start the agent: ./awagent --verbose"
echo "  3. Install as service: ./install.sh"
echo ""
echo "View logs: journalctl --user -u awagent -f"
echo "View dashboard: http://localhost:5600"
echo ""
echo "For detailed instructions, see DEPLOYMENT.md"
