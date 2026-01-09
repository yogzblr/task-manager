#!/bin/bash
set -e

# Automation Agent Installation Script
# Supports Debian/Ubuntu and RHEL/CentOS

BINARY_NAME="automation-agent"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/sysconfig"
DATA_DIR="/var/lib/automation-agent"
LOG_DIR="/var/log/automation-agent"
USER="automation-agent"
GROUP="automation-agent"

# Detect distribution
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
else
    echo "Cannot detect Linux distribution"
    exit 1
fi

echo "Detected distribution: $DISTRO"

# Check for root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Create user and group
if ! id "$USER" &>/dev/null; then
    echo "Creating user $USER..."
    useradd -r -s /bin/false -d "$DATA_DIR" "$USER"
fi

# Create directories
echo "Creating directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"
chown "$USER:$GROUP" "$DATA_DIR"
chown "$USER:$GROUP" "$LOG_DIR"

# Install binary
if [ -f "./$BINARY_NAME" ]; then
    echo "Installing binary..."
    cp "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    chown root:root "$INSTALL_DIR/$BINARY_NAME"
else
    echo "Binary not found: ./$BINARY_NAME"
    exit 1
fi

# Install systemd service
echo "Installing systemd service..."
cp "./automation-agent.service" "$SERVICE_DIR/$BINARY_NAME.service"

# Install configuration file
if [ ! -f "$CONFIG_DIR/$BINARY_NAME" ]; then
    echo "Installing configuration file..."
    cp "./automation-agent.sysconfig" "$CONFIG_DIR/$BINARY_NAME"
    echo "Please edit $CONFIG_DIR/$BINARY_NAME with your configuration"
fi

# Reload systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
systemctl enable "$BINARY_NAME.service"
systemctl start "$BINARY_NAME.service"

echo "Installation complete!"
echo "Service status:"
systemctl status "$BINARY_NAME.service" --no-pager
