#!/bin/bash
set -e
CGROUP_CPU=/sys/fs/cgroup/cpu/cpu.shares
if [ -f "$CGROUP_CPU" ]; then
  SHARES=$(cat $CGROUP_CPU)
  CPUS=$(($SHARES / 1024))
  echo "$SHARES detected in the cgroup, rounds down to $CPUS cpus"
else
  echo "No CGroup shares detected, will use default value of auto"
  CPUS="auto"
fi

