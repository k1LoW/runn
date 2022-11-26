#!/usr/bin/env bash

set -e

cp /keys/id_rsa.pub /root/.ssh/authorized_keys
cp /keys/id_rsa.pub /etc/authorized_keys/testuser
chmod 600 /root/.ssh/authorized_keys
chmod 600 /etc/authorized_keys/testuser
chown root: /root/.ssh/authorized_keys
chown testuser: /etc/authorized_keys/testuser
ls -la /root/.ssh
ls -la /etc/authorized_keys
