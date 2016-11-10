#!/bin/bash

# This script removes consecutive empty lines in nginx.conf
# Using sed is more simple than using a go regex

# first sed removes empty lines
# second sed command replaces the empty lines
sed -e 's/^  *$/\'$'\n/g' | sed -e '/^$/{N;/^\n$/d;}'
