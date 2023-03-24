#!/usr/bin/env bash

set -e

apk add sudo
echo "testuser:testpass" | chpasswd
echo "testuser ALL=NOPASSWD: ALL" >> /etc/sudoers
