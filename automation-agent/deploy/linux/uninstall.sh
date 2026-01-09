#!/bin/bash
set -e

# Automation Agent Uninstallation Script

BINARY_NAME="automation-agent"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/sysconfig"
DATA_DIR="/var/lib/automation-agent"
LOG_DIR="/var/log/automation-agent"
USER="automation-agent"

# Check for root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Stop and disable service
if systemctl is-active --quiet "$BINARY_NAME.service"; then
    echo "Stopping service..."
    systemctl stop "$BINARY_NAME.service"
fi

if systemctl is-enabled --quiet "$BINARY_NAME.service"; then
    echo "Disabling service..."
    systemctl disable "$BINARY_NAME.service"
fi

# Remove service file
if [ -f "$SERVICE_DIR/$BINARY_NAME.service" ]; then
    echo "Removing service file..."
    rm "$SERVICE_DIR/$BINARY_NAME.service"
    systemctl daemon-reload
fi

# Remove binary
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    echo "Removing binary..."
    rm "$INSTALL_DIR/$BINARY_NAME"
fi

# Remove configuration (optional)
read -p "Remove configuration file? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -f "$CONFIG_DIR/$BINARY_NAME" ]; then
        rm "$CONFIG_DIR/$BINARY_NAME"
    fi
fi

# Remove data/log directories (optional)
read -p "Remove data and log directories? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$DATA_DIR" ]; then
        rm -rf "$DATA_DIR"
    fi
    if [ -d "$LOG_DIR" ]; then
        rm -rf "$LOG_DIR"
    fi
fi

# Remove user (optional)
read -p "Remove user $USER? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if id "$USER" &>/dev/null; then
        userdel "$USER"
    fi
fi

echo "Uninstallation complete!"
