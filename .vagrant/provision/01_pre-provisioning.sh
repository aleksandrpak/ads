#!/usr/bin/env bash

echo "Setting the Timezone to Asia/Samarkand"
echo "Asia/Samarkand" | sudo tee /etc/timezone && dpkg-reconfigure -f noninteractive tzdata

echo "Installing russian locale"
sudo locale-gen ru_RU.UTF-8

# Stop errors like 'Cannot allocate memory'
if grep -qF "swapfile" /etc/fstab; then
    echo "Swap file found. Do nothing."
else
    echo "Creating swapfile of 1GB with block size 1MB"
    # Exit on any error (non-zero return code)
    set -e
    # Create swapfile of 1GB with block size 1MB
    /bin/dd if=/dev/zero of=/swapfile bs=1024 count=1048576
    # Set up the swap file
    /sbin/mkswap /swapfile
    # Enable swap file immediately
    /sbin/swapon /swapfile
    # Enable swap file on every boot
    /bin/echo "/swapfile          swap            swap    defaults        0 0" >> /etc/fstab
fi
