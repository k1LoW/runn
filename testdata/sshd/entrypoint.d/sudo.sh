#!/usr/bin/env bash

set -e

apk add sudo
echo "testuser:testpass" | chpasswd
echo "testuser ALL=(ALL) ALL" >> /etc/sudoers
