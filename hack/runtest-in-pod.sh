#!/bin/sh

cat ${CFG_DIR:-/testconfig}/dns-entries >> /etc/hosts

ptptests -test.v


