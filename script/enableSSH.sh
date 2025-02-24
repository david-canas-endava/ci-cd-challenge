#!/bin/bash

# Install OpenSSH Server if not installed
sudo apt update && sudo apt install -y openssh-server

# Enable and start the SSH service
sudo systemctl enable ssh
sudo systemctl start ssh

# Allow SSH through the firewall (if UFW is enabled)
sudo ufw allow OpenSSH || echo "UFW not installed, skipping firewall setup"

# Ensure Password Authentication is enabled
sudo sed -i 's/^#PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config
sudo sed -i 's/^PasswordAuthentication no/PasswordAuthentication yes/' /etc/ssh/sshd_config

# Restart SSH to apply changes
sudo systemctl restart ssh

echo "SSH setup completed."
