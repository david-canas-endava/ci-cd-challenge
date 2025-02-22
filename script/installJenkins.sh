#!/bin/bash
sudo apt update && sudo apt upgrade -y
sudo apt autoremove -y
sudo apt install -y software-properties-common apt-transport-https ca-certificates
sudo apt install openjdk-17-jdk -y

sudo wget -O /usr/share/keyrings/jenkins-keyring.asc \
  https://pkg.jenkins.io/debian-stable/jenkins.io-2023.key
echo "deb [signed-by=/usr/share/keyrings/jenkins-keyring.asc]" \
  https://pkg.jenkins.io/debian-stable binary/ | sudo tee \
  /etc/apt/sources.list.d/jenkins.list > /dev/null

wget http://ftp.debian.org/debian/pool/main/i/init-system-helpers/init-system-helpers_1.60_all.deb
sudo dpkg -i init-system-helpers_1.60_all.deb
sudo apt --fix-broken install -y
sudo apt-get update
sudo apt install -y jenkins
