#!/bin/bash

# Configuration
APP_NAME="thelistbot"
INSTALL_DIR="/opt/thelistbot"
CONFIG_DIR="${INSTALL_DIR}/config"
SERVICE_FILE="thelistbot.service"
USERNAME=$(whoami)  # Use current user or specify a different one

# Build the application
echo "Building application..."
./build_linux.sh

if [ $? -ne 0 ]; then
    echo "Build failed! Exiting."
    exit 1
fi

# Create the installation directory if it doesn't exist
echo "Creating installation directories..."
sudo mkdir -p ${INSTALL_DIR}
sudo mkdir -p ${CONFIG_DIR}

# Copy files
echo "Copying files..."
sudo cp ./bin/${APP_NAME}-linux-amd64 ${INSTALL_DIR}/${APP_NAME}
sudo cp .env ${INSTALL_DIR}/

# Set permissions
echo "Setting permissions..."
sudo chown -R ${USERNAME}:${USERNAME} ${INSTALL_DIR}
sudo chmod +x ${INSTALL_DIR}/${APP_NAME}

# Update service file with correct username
echo "Preparing service file..."
sed "s/User=your_username/User=${USERNAME}/g" ${SERVICE_FILE} | \
sed "s/Group=your_username/Group=${USERNAME}/g" > /tmp/thelistbot.service

# Install service
echo "Installing systemd service..."
sudo cp /tmp/thelistbot.service /etc/systemd/system/${SERVICE_FILE}
sudo systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
sudo systemctl enable ${SERVICE_FILE}
sudo systemctl restart ${SERVICE_FILE}

# Show status
echo "Service status:"
sudo systemctl status ${SERVICE_FILE} --no-pager

echo ""
echo "Deployment complete!"
echo "Use 'sudo systemctl status ${SERVICE_FILE}' to check the status."
echo "Use 'sudo journalctl -u ${SERVICE_FILE} -f' to view logs."
