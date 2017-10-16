#!/bin/bash

# This script removes consecutive empty lines in nginx.conf
# Using sed is more simple than using a go regex

# Sed commands:
# 1. remove the return carrier character/s
# 2. remove empty lines
# 3. replace multiple empty lines
sed -e 's/\r//g' | sed -e 's/^  *$/\'$'\n/g' | sed -e '/^$/{N;/^\n$/D;}'
