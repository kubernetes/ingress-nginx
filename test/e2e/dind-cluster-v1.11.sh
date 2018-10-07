#!/bin/bash
# Copyright 2017 Mirantis
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
set -o errexit
set -o nounset
set -o pipefail
set -o errtrace

if [ $(uname) = Darwin ]; then
  readlinkf(){ perl -MCwd -e 'print Cwd::abs_path shift' "$1";}
else
  readlinkf(){ readlink -f "$1"; }
fi
DIND_ROOT="$(cd $(dirname "$(readlinkf "${BASH_SOURCE}")"); pwd)"

RUN_ON_BTRFS_ANYWAY="${RUN_ON_BTRFS_ANYWAY:-}"
if [[ ! ${RUN_ON_BTRFS_ANYWAY} ]] && docker info| grep -q '^Storage Driver: btrfs'; then
  echo "ERROR: Docker is using btrfs storage driver which is unsupported by kubeadm-dind-cluster" >&2
  echo "Please refer to the documentation for more info." >&2
  echo "Set RUN_ON_BTRFS_ANYWAY to non-empty string to continue anyway." >&2
  exit 1
fi

# In case of linuxkit / moby linux, -v will not work so we can't
# mount /lib/modules and /boot. Also we'll be using localhost
# to access the apiserver.
using_linuxkit=
if ! docker info|grep -s '^Operating System: .*Docker for Windows' > /dev/null 2>&1 ; then
    if docker info|grep -s '^Kernel Version: .*-moby$' >/dev/null 2>&1 ||
         docker info|grep -s '^Kernel Version: .*-linuxkit-' > /dev/null 2>&1 ; then
        using_linuxkit=1
    fi
fi

# Determine when using Linux and docker daemon running locally
using_local_linuxdocker=
if [[ $(uname) == Linux && -z ${DOCKER_HOST:-} ]]; then
    using_local_linuxdocker=1
fi

EMBEDDED_CONFIG=y;DIND_IMAGE=mirantis/kubeadm-dind-cluster:v1.11

# dind::localhost provides the local host IP based on the address family used for service subnet.
function dind::localhost() {
  if [[ ${SERVICE_NET_MODE} = "ipv6" ]]; then
    echo '[::1]'
  else
    echo '127.0.0.1'
  fi
}

# dind::family-for indicates whether the CIDR or IP is for an IPv6 or IPv4 family.
function dind::family-for {
    local addr=$1
    if [[ "$addr" = *":"* ]]; then
	echo "ipv6"
    else
	echo "ipv4"
    fi
}

# dind::cluster-suffix builds a suffix used for resources, based on the DIND_LABEL.
function dind::cluster-suffix {
  if [ "$DIND_LABEL" != "$DEFAULT_DIND_LABEL" ]; then
    echo "-${DIND_LABEL}"
  else
    echo ''
  fi
}

function dind::net-name {
  echo "kubeadm-dind-net$( dind::cluster-suffix )"
}

# dind::add-cluster will inject the cluster ID to the IP address. For IPv4, it is
# assumed that the IP is a /24 with the third part of the address available for cluster ID.
# For IPv6, it is assumed that there is enough space for the cluster to be added, and the
# cluster ID will be added to the 16 bits before the double colon. For example:
#
# 10.192.0.0/24 => 10.192.5.0/24
# fd00:77:20::/64 => fd00:77:20:5::/64
#
# This function is intended to be used for management networks.
#
# TODO: Validate that there is enough space for cluster ID.
# TODO: For IPv6 could get fancy and handle case where cluster ID is placed in upper 8 bits of hextet
# TODO: Consider if want to do /16 for IPv4 management subnet.
#
function dind::add-cluster {
  local cidr=$1
  local ip_mode=$2

  if [[ ${ip_mode} = "ipv4" ]]; then
      echo ${cidr} | sed "s/^\([0-9]*\.[0-9]*\.\).*\/24$/\1${CLUSTER_ID}.0\/24/"
  else  # IPv6
      echo ${cidr} | sed "s/^\(.*\)\(\:\:\/[0-9]*\)$/\1:${CLUSTER_ID}\2/"
  fi
}

# dind::get-and-validate-cidrs takes a list of CIDRs and validates them based on the ip
# mode, returning them. For IPv4 only and IPv6 only modes, only one CIDR is expected. For
# dual stack, two CIDRS are expected. It verifies that the CIDRs are the right family and
# will use the provided defaults, when CIDRs are missing. For dual-stack, the IPv4 address
# will be first.
#
# For the management network, the cluster ID will be injected into the CIDR. Also, if no
# MGMT_CIDRS value is specified, but the legacy DIND_SUBNET/DIND_SUBNET_SIZE is provided,
# that will be used for the (first) CIDR.
#
# NOTE: It is expected that the CIDR size is /24 for IPv4 management networks.
#
# TODO: Generalize so this can be used for pod network to support dual-stack.
#
function dind::get-and-validate-cidrs {
  IFS=', ' read -r -a cidrs <<< "$1"
  IFS=', ' read -r -a defaults <<< "$2"
  case ${IP_MODE} in
    ipv4)
      case ${#cidrs[@]} in
        0)
	  cidrs[0]="${defaults[0]}"
          ;;
        1)
          ;;
        *)
          echo "ERROR! More than one CIDR provided '$1'"
          exit 1
          ;;
      esac
      if [[ $( dind::family-for "${cidrs[0]}" ) != "ipv4" ]]; then
	echo "ERROR! CIDR must be IPv4 value"
	exit 1
      fi
      cidrs[0]="$( dind::add-cluster "${cidrs[0]}" "${IP_MODE}" )"
      ;;

    ipv6)
      case ${#cidrs[@]} in
        0)
          cidrs[0]="${defaults[0]}"
	  ;;
        1)
	  ;;
        *)
          echo "ERROR! More than one CIDR provided '$1'"
          exit 1
          ;;
      esac
      if [[ $( dind::family-for "${cidrs[0]}" ) != "ipv6" ]]; then
	echo "ERROR! CIDR must be IPv6 value"
	exit 1
      fi
      cidrs[0]="$( dind::add-cluster "${cidrs[0]}" "${IP_MODE}" )"
      ;;

    dual-stack)
      case ${#cidrs[@]} in
        0)
          cidrs[0]="${defaults[0]}"
          cidrs[1]="${defaults[1]}"
          ;;
        1)
          if [[ $( dind::family-for "${cidrs[0]}" ) = "ipv6" ]]; then
	    cidrs[1]=${cidrs[0]}
	    cidrs[0]="${defaults[0]}"  # Assuming first default is a V4 address
	  else
	    cidrs[1]="${defaults[1]}"
	  fi
          ;;
        2)
	  # Force ordering to have V4 address first
          if [[ $( dind::family-for "${cidrs[0]}" ) = "ipv6" ]]; then
	    local temp=${cidrs[0]}
	    cidrs[0]=${cidrs[1]}
	    cidrs[1]=${temp}
	  fi
          ;;
        *)
          echo "ERROR! More than two CIDRs provided '$1'"
          exit 1
          ;;
      esac
      local have_v4=""
      local have_v6=""
      for cidr in ${cidrs[@]}; do
        if [[ $( dind::family-for "${cidr}" ) = "ipv6" ]]; then
	  have_v6=1
	else
	  have_v4=1
	fi
      done
      if [[ -z ${have_v4} ]]; then
        echo "ERROR! Missing IPv4 CIDR in '$1'"
        exit 1
      fi
      if [[ -z ${have_v6} ]]; then
        echo "ERROR! Missing IPv6 CIDR in '$1'"
        exit 1
      fi
      cidrs[0]="$( dind::add-cluster "${cidrs[0]}" "${IP_MODE}" )"
      cidrs[1]="$( dind::add-cluster "${cidrs[1]}" "${IP_MODE}" )"
      ;;
  esac
  echo "${cidrs[@]}"
}

# dind::make-ip-from-cidr  strips off the slash and size, and appends the
# interface part to the prefix to form an IP. For IPv4, it strips off the
# fourth part of the prefix, so that it can be replaced. It assumes that the
# resulting prefix will be of sufficient size. It also will use hex for the
# appended part for IPv6, and decimal for IPv4.
#
# fd00:20::/64 -> fd00:20::a
# 10.96.0.0/12 -> 10.96.0.10
#
function dind::make-ip-from-cidr {
  prefix="$(echo $1 | sed 's,/.*,,')"
  if [[ $( dind::family-for ${prefix} ) == "ipv4" ]]; then
    printf "%s%d" $( echo ${prefix} | sed 's/0$//' ) $2
  else
    printf "%s%x" ${prefix} $2
  fi
}

# dind::add-cluster-id-and-validate-nat64-prefix will modify the IPv4 mapping
# subnet prefix, by adding the cluster ID (default 0) to the second octet.
# It will produce an error, if the prefix is not in the 10.0.0.0/8 or
# 172.16.0.0/12 private networks.
function dind::add-cluster-id-and-validate-nat64-prefix {
  local parts
  IFS="." read -a parts <<<${NAT64_V4_SUBNET_PREFIX}
  if [[ ${#parts[@]} -ne 2 ]]; then
    echo "ERROR! NAT64_V4_SUBNET_PREFIX must be two octets (have '${NAT64_V4_SUBNET_PREFIX}')"
    exit 1
  fi
  (( parts[1]+=${CLUSTER_ID} ))
  NAT64_V4_SUBNET_PREFIX="${parts[0]}.${parts[1]}"
  echo "Added cluster ID offset (${CLUSTER_ID}) to NAT64_V4_SUBNET_PREFIX giving prefix '${NAT64_V4_SUBNET_PREFIX}'"
  if [[ ${parts[0]} -eq 10 ]]; then
    if [[ ${parts[1]} > 253 ]]; then
      echo "ERROR! NAT64_V4_SUBNET_PREFIX is too large for 10.0.0.0/8 private net"
      exit 1
    fi
  elif [[ ${parts[0]} -eq 172 ]]; then
    if [[ ${parts[1]} -lt 16 || ${parts[1]} -gt 31 ]]; then
      echo "ERROR! NAT64_V4_SUBNET_PREFIX is outside of range for 172.16.0.0/12 private net"
      exit 1
    fi
  else
      echo "ERROR! NAT64_V4_SUBNET_PREFIX is not in 10.0.0.0/8 or 172.16.0.0/12 private networks"
      exit 1
  fi
  echo "Using NAT64 V4 mapping network prefix: ${NAT64_V4_SUBNET_PREFIX}"
}


# START OF PROCESSING...

IP_MODE="${IP_MODE:-ipv4}"  # ipv4, ipv6, dual-stack
if [[ ! ${EMBEDDED_CONFIG:-} ]]; then
  source "${DIND_ROOT}/config.sh"
fi

# Multicluster support
# Users can specify a cluster ID number from 1..254, represented as a string.
# This will be used to form resource names "cluster-#", and will be used in the
# management subnet to give unique networks for each cluster. If the cluster ID
# is not specified, or zero, it will be considered a single cluster or the first
# in the multi-cluster. This is the recommended usage.
#
# For legacy support, the user can specify DIND_LABEL, which will be used in the
# resource names. If a cluster ID is specified (a hybrid case, where people are
# using the new method, but want custom names), the resourse name will have the
# suffix "-#" with the cluster ID. If no cluster ID is specified (for backward
# compatibility), then the resource name will be just the DIND_LABEL, and a pseudo-
# random number from 1..13 will be generated for the cluster ID to be used in
# management network. The range is limited, because, in IPv6 mode, the cluster ID
# is used in the NAT64 V4 subnet prefix, which must be in a private network.
# The default is 172.18, so the cluster ID cannot be larger than 13 to guarantee
# a valid value.
#
# To get around that limitation, you can set the cluster ID, in addition to the
# DIND_LABEL, and optionally, change the NAT64_V4_SUBNET_PREFIX value.
#
DEFAULT_DIND_LABEL='mirantis.kubeadm_dind_cluster_runtime'
if [[ -z ${DIND_LABEL+x} ]]; then  # No legacy DIND_LABEL set
  if [[ -z ${CLUSTER_ID+x} ]]; then  # No cluster ID set
    DIND_LABEL=${DEFAULT_DIND_LABEL}  # Single cluster mode
    CLUSTER_ID="0"
  else  # Have cluster ID
    if [[ ${CLUSTER_ID} = "0" ]]; then
      DIND_LABEL=${DEFAULT_DIND_LABEL}  # Single cluster mode or first cluster of multi-cluster
    else
      DIND_LABEL="cluster-${CLUSTER_ID}"  # Multi-cluster
    fi
  fi
else  # Legacy DIND_LABEL set for multi-cluster
  if [[ -z ${CLUSTER_ID+x} ]]; then  # No cluster ID set, make one from 1..13, but don't use in resource names
    CLUSTER_ID="$(( ($RANDOM % 12) + 1 ))"
  else
    if [[ ${CLUSTER_ID} = "0" ]]; then
      CLUSTER_ID="$(( ($RANDOM % 12) + 1 ))"  # Force a pseudo-random cluster for additional legacy cluster
    else
      DIND_LABEL="${DIND_LABEL}-${CLUSTER_ID}"
    fi
  fi
fi

CNI_PLUGIN="${CNI_PLUGIN:-bridge}"
GCE_HOSTED="${GCE_HOSTED:-}"
DIND_ALLOW_AAAA_USE="${DIND_ALLOW_AAAA_USE:-}"  # Default is to use DNS64 always for IPv6 mode

# Use legacy DIND_SUBNET/DIND_SUBNET_SIZE, only if MGMT_CIDRS is not set.
legacy_mgmt_cidr=""
if [[ ${DIND_SUBNET:-} && ${DIND_SUBNET_SIZE:-} ]]; then
  legacy_mgmt_cidr="${DIND_SUBNET}/${DIND_SUBNET_SIZE}"
fi

if [[ ${IP_MODE} = "dual-stack" ]]; then
  mgmt_net_defaults="10.192.0.0/24, fd00:20::/64"

  KUBE_RSYNC_ADDR="${KUBE_RSYNC_ADDR:-::1}"
  SERVICE_CIDR="${SERVICE_CIDR:-fd00:30::/110}"  # Will default to IPv6 service net family

  DEFAULT_POD_NETWORK_CIDR="fd00:40::/72"  # TODO: convert for dual-stack

  USE_HAIRPIN="${USE_HAIRPIN:-true}"  # Default is to use hairpin for dual-stack
  if [[ ${DIND_ALLOW_AAAA_USE} && ${GCE_HOSTED} ]]; then
    echo "ERROR! GCE does not support use of IPv6 for external addresses - aborting."
    exit 1
  fi
elif [[ ${IP_MODE} = "ipv6" ]]; then
  mgmt_net_defaults="fd00:20::/64"

  KUBE_RSYNC_ADDR="${KUBE_RSYNC_ADDR:-::1}"
  SERVICE_CIDR="${SERVICE_CIDR:-fd00:30::/110}"

  DEFAULT_POD_NETWORK_CIDR="fd00:40::/72"  # TODO: convert for dual-stack

  USE_HAIRPIN="${USE_HAIRPIN:-true}"  # Default is to use hairpin for IPv6
  if [[ ${DIND_ALLOW_AAAA_USE} && ${GCE_HOSTED} ]]; then
    echo "ERROR! GCE does not support use of IPv6 for external addresses - aborting."
    exit 1
  fi
else  # IPv4 mode
  mgmt_net_defaults="10.192.0.0/24"

  KUBE_RSYNC_ADDR="${KUBE_RSYNC_ADDR:-127.0.0.1}"
  SERVICE_CIDR="${SERVICE_CIDR:-10.96.0.0/12}"

  DEFAULT_POD_NETWORK_CIDR="10.244.0.0/16"  # TODO: convert for dual-stack

  USE_HAIRPIN="${USE_HAIRPIN:-false}"  # Disabled for IPv4, as issue with Virtlet networking
  if [[ ${DIND_ALLOW_AAAA_USE} ]]; then
    echo "WARNING! The DIND_ALLOW_AAAA_USE option is for IPv6 mode - ignoring setting."
    DIND_ALLOW_AAAA_USE=
  fi
  if [[ ${CNI_PLUGIN} = "calico" || ${CNI_PLUGIN} = "calico-kdd" ]]; then
    DEFAULT_POD_NETWORK_CIDR="192.168.0.0/16"  # TODO: convert for dual-stack?
  fi
fi

IFS=' ' read -r -a mgmt_net_cidrs <<<$( dind::get-and-validate-cidrs "${MGMT_CIDRS:-${legacy_mgmt_cidr}}" "${mgmt_net_defaults[@]}" )

REMOTE_DNS64_V4SERVER="${REMOTE_DNS64_V4SERVER:-8.8.8.8}"
if [[ ${IP_MODE} == "ipv6" ]]; then
  # Uses local DNS64 container
  dns_server="$( dind::make-ip-from-cidr ${mgmt_net_cidrs[0]} 0x100 )"
  DNS64_PREFIX="${DNS64_PREFIX:-fd00:10:64:ff9b::}"
  DNS64_PREFIX_SIZE="${DNS64_PREFIX_SIZE:-96}"
  DNS64_PREFIX_CIDR="${DNS64_PREFIX}/${DNS64_PREFIX_SIZE}"

  LOCAL_NAT64_SERVER="$( dind::make-ip-from-cidr ${mgmt_net_cidrs[0]} 0x200 )"
  NAT64_V4_SUBNET_PREFIX="${NAT64_V4_SUBNET_PREFIX:-172.18}"
  dind::add-cluster-id-and-validate-nat64-prefix
else
  dns_server="${REMOTE_DNS64_V4SERVER}"
fi

SERVICE_NET_MODE="$( dind::family-for ${SERVICE_CIDR} )"
DNS_SVC_IP="$( dind::make-ip-from-cidr ${SERVICE_CIDR} 10 )"

ETCD_HOST="${ETCD_HOST:-$( dind::localhost )}"

POD_NETWORK_CIDR="${POD_NETWORK_CIDR:-${DEFAULT_POD_NETWORK_CIDR}}"
if [[ ${IP_MODE} = "ipv6" ]]; then
    # For IPv6 will extract the network part and size from pod cluster CIDR.
    # The size will be increased by eight, as the pod network will be split
    # into subnets for each node. The network part will be converted into a
    # prefix that will get the node ID appended, for each node. In some cases
    # this means padding the prefix. In other cases, the prefix must be
    # trimmed, so that when the node ID is added, it forms a correct prefix.
    POD_NET_PREFIX="$(echo ${POD_NETWORK_CIDR} | sed 's,::/.*,:,')"
    cluster_size="$(echo ${POD_NETWORK_CIDR} | sed 's,.*::/,,')"
    POD_NET_SIZE=$((${cluster_size}+8))

    num_colons="$(grep -o ":" <<< "${POD_NET_PREFIX}" | wc -l)"
    need_zero_pads=$((${cluster_size}/16))

    # Will be replacing lowest byte with node ID, so pull off last byte and colon
    if [[ ${num_colons} -gt ${need_zero_pads} ]]; then
	POD_NET_PREFIX=${POD_NET_PREFIX::-3}
    fi
    # Add in zeros to pad 16 bits at a time, up to the padding needed, which is
    # need_zero_pads - num_colons.
    while [ ${num_colons} -lt ${need_zero_pads} ]; do
	POD_NET_PREFIX+="0:"
        num_colons+=1
    done
elif [[ ${CNI_PLUGIN} = "bridge" || ${CNI_PLUGIN} = "ptp" ]]; then # IPv4, bridge or ptp
    # For IPv4, will assume user specifies /16 CIDR and will use a /24 subnet
    # on each node.
    POD_NET_PREFIX="$(echo ${POD_NETWORK_CIDR} | sed 's/^\([0-9]*\.[0-9]*\.\).*/\1/')"
    POD_NET_SIZE=24
else
    POD_NET_PREFIX=
    POD_NET_SIZE=
fi

DIND_IMAGE="${DIND_IMAGE:-}"
BUILD_KUBEADM="${BUILD_KUBEADM:-}"
BUILD_HYPERKUBE="${BUILD_HYPERKUBE:-}"
KUBEADM_SOURCE="${KUBEADM_SOURCE-}"
HYPERKUBE_SOURCE="${HYPERKUBE_SOURCE-}"
NUM_NODES=${NUM_NODES:-2}
EXTRA_PORTS="${EXTRA_PORTS:-}"
LOCAL_KUBECTL_VERSION=${LOCAL_KUBECTL_VERSION:-}
KUBECTL_DIR="${KUBECTL_DIR:-${HOME}/.kubeadm-dind-cluster}"
DASHBOARD_URL="${DASHBOARD_URL:-https://rawgit.com/kubernetes/dashboard/bfab10151f012d1acc5dfb1979f3172e2400aa3c/src/deploy/kubernetes-dashboard.yaml}"
SKIP_SNAPSHOT="${SKIP_SNAPSHOT:-}"
E2E_REPORT_DIR="${E2E_REPORT_DIR:-}"
DIND_NO_PARALLEL_E2E="${DIND_NO_PARALLEL_E2E:-}"
DNS_SERVICE="${DNS_SERVICE:-kube-dns}"
DIND_STORAGE_DRIVER="${DIND_STORAGE_DRIVER:-overlay2}"

DIND_CA_CERT_URL="${DIND_CA_CERT_URL:-}"
DIND_PROPAGATE_HTTP_PROXY="${DIND_PROPAGATE_HTTP_PROXY:-}"
DIND_HTTP_PROXY="${DIND_HTTP_PROXY:-}"
DIND_HTTPS_PROXY="${DIND_HTTPS_PROXY:-}"
DIND_NO_PROXY="${DIND_NO_PROXY:-}"

DIND_DAEMON_JSON_FILE="${DIND_DAEMON_JSON_FILE:-/etc/docker/daemon.json}"  # can be set to /dev/null
DIND_REGISTRY_MIRROR="${DIND_REGISTRY_MIRROR:-}"  # plain string format
DIND_INSECURE_REGISTRIES="${DIND_INSECURE_REGISTRIES:-}"  # json list format

FEATURE_GATES="${FEATURE_GATES:-MountPropagation=true}"
# you can set special value 'none' not to set any kubelet's feature gates.
KUBELET_FEATURE_GATES="${KUBELET_FEATURE_GATES:-MountPropagation=true,DynamicKubeletConfig=true}"

KUBELET_EXTRA_ARGS="--sync-frequency=30s"

if [[ ! ${LOCAL_KUBECTL_VERSION:-} && ${DIND_IMAGE:-} =~ :(v[0-9]+\.[0-9]+)$ ]]; then
  LOCAL_KUBECTL_VERSION="${BASH_REMATCH[1]}"
fi

ENABLE_CEPH="${ENABLE_CEPH:-}"

# TODO: Test multi-cluster for IPv6, before enabling
if [[ "${DIND_LABEL}" != "${DEFAULT_DIND_LABEL}"  && "${IP_MODE}" == 'dual-stack' ]]; then
    echo "Multiple parallel clusters currently not supported for dual-stack mode" >&2
    exit 1
fi

# not configurable for now, would need to setup context for kubectl _inside_ the cluster
readonly INTERNAL_APISERVER_PORT=8080

function dind::need-source {
  if [[ ! -f cluster/kubectl.sh ]]; then
    echo "$0 must be called from the Kubernetes repository root directory" 1>&2
    exit 1
  fi
}

build_tools_dir="build"
use_k8s_source=y
if [[ ! ${BUILD_KUBEADM} && ! ${BUILD_HYPERKUBE} ]]; then
  use_k8s_source=
fi
if [[ ${use_k8s_source} ]]; then
  dind::need-source
  kubectl=cluster/kubectl.sh
  if [[ ! -f ${build_tools_dir}/common.sh ]]; then
    build_tools_dir="build-tools"
  fi
else
  if [[ ! ${LOCAL_KUBECTL_VERSION:-} ]] && ! hash kubectl 2>/dev/null; then
    echo "You need kubectl binary in your PATH to use prebuilt DIND image" 1>&2
    exit 1
  fi
  kubectl=kubectl
fi

function dind::retry {
  # based on retry function in hack/jenkins/ scripts in k8s source
  for i in {1..10}; do
    "$@" && return 0 || sleep ${i}
  done
  "$@"
}

busybox_image="busybox:1.26.2"
e2e_base_image="golang:1.9.2"
sys_volume_args=()
build_volume_args=()

function dind::set-build-volume-args {
  if [ ${#build_volume_args[@]} -gt 0 ]; then
    return 0
  fi
  build_container_name=
  if [ -n "${KUBEADM_DIND_LOCAL:-}" ]; then
    build_volume_args=(-v "$PWD:/go/src/k8s.io/kubernetes")
  else
    build_container_name="$(KUBE_ROOT=${PWD} ETCD_HOST=${ETCD_HOST} &&
                            . ${build_tools_dir}/common.sh &&
                            kube::build::verify_prereqs >&2 &&
                            echo "${KUBE_DATA_CONTAINER_NAME:-${KUBE_BUILD_DATA_CONTAINER_NAME}}")"
    build_volume_args=(--volumes-from "${build_container_name}")
  fi
}

function dind::volume-exists {
  local name="$1"
  if docker volume inspect "${name}" >& /dev/null; then
    return 0
  fi
  return 1
}

function dind::create-volume {
  local name="$1"
  docker volume create --label "${DIND_LABEL}" --name "${name}" >/dev/null
}

# We mount /boot and /lib/modules into the container
# below to in case some of the workloads need them.
# This includes virtlet, for instance. Also this may be
# useful in future if we want DIND nodes to pass
# preflight checks.
# Unfortunately we can't do this when using Mac Docker
# (unless a remote docker daemon on Linux is used)
# NB: there's no /boot on recent Mac dockers
function dind::prepare-sys-mounts {
  if [[ ! ${using_linuxkit} ]]; then
    sys_volume_args=()
    if [[ -d /boot ]]; then
      sys_volume_args+=(-v /boot:/boot)
    fi
    if [[ -d /lib/modules ]]; then
      sys_volume_args+=(-v /lib/modules:/lib/modules)
    fi
    return 0
  fi
  local dind_sys_vol_name
  dind_sys_vol_name="kubeadm-dind-sys$( dind::cluster-suffix )"
  if ! dind::volume-exists "$dind_sys_vol_name"; then
    dind::step "Saving a copy of docker host's /lib/modules"
    dind::create-volume "$dind_sys_vol_name"
    # Use a dirty nsenter trick to fool Docker on Mac and grab system
    # /lib/modules into sys.tar file on kubeadm-dind-sys volume.
    local nsenter="nsenter --mount=/proc/1/ns/mnt --"
    docker run \
           --rm \
           --privileged \
           -v "$dind_sys_vol_name":/dest \
           --pid=host \
           "${busybox_image}" \
           /bin/sh -c \
           "if ${nsenter} test -d /lib/modules; then ${nsenter} tar -C / -c lib/modules >/dest/sys.tar; fi"
  fi
  sys_volume_args=(-v "$dind_sys_vol_name":/dind-sys)
}

tmp_containers=()

function dind::cleanup {
  if [ ${#tmp_containers[@]} -gt 0 ]; then
    for name in "${tmp_containers[@]}"; do
      docker rm -vf "${name}" 2>/dev/null
    done
  fi
}

trap dind::cleanup EXIT

function dind::check-image {
  local name="$1"
  if docker inspect --format 'x' "${name}" >&/dev/null; then
    return 0
  else
    return 1
  fi
}

function dind::filter-make-output {
  # these messages make output too long and make Travis CI choke
  egrep -v --line-buffered 'I[0-9][0-9][0-9][0-9] .*(parse|conversion|defaulter|deepcopy)\.go:[0-9]+\]'
}

function dind::run-build-command {
    # this is like build/run.sh, but it doesn't rsync back the binaries,
    # only the generated files.
    local cmd=("$@")
    (
        # The following is taken from build/run.sh and build/common.sh
        # of Kubernetes source tree. It differs in
        # --filter='+ /_output/dockerized/bin/**'
        # being removed from rsync
        . ${build_tools_dir}/common.sh
        kube::build::verify_prereqs
        kube::build::build_image
        kube::build::run_build_command "$@"

        kube::log::status "Syncing out of container"

        kube::build::start_rsyncd_container

        local rsync_extra=""
        if (( ${KUBE_VERBOSE} >= 6 )); then
            rsync_extra="-iv"
        fi

        # The filter syntax for rsync is a little obscure. It filters on files and
        # directories.  If you don't go in to a directory you won't find any files
        # there.  Rules are evaluated in order.  The last two rules are a little
        # magic. '+ */' says to go in to every directory and '- /**' says to ignore
        # any file or directory that isn't already specifically allowed.
        #
        # We are looking to copy out all of the built binaries along with various
        # generated files.
        kube::build::rsync \
            --filter='- /vendor/' \
            --filter='- /_temp/' \
            --filter='+ zz_generated.*' \
            --filter='+ generated.proto' \
            --filter='+ *.pb.go' \
            --filter='+ types.go' \
            --filter='+ */' \
            --filter='- /**' \
            "rsync://k8s@${KUBE_RSYNC_ADDR}/k8s/" "${KUBE_ROOT}"

        kube::build::stop_rsyncd_container
    )
}

function dind::make-for-linux {
  local copy="$1"
  shift
  dind::step "Building binaries:" "$*"
  if [ -n "${KUBEADM_DIND_LOCAL:-}" ]; then
    dind::step "+ make WHAT=\"$*\""
    make WHAT="$*" 2>&1 | dind::filter-make-output
  elif [ "${copy}" = "y" ]; then
    dind::step "+ ${build_tools_dir}/run.sh make WHAT=\"$*\""
    "${build_tools_dir}/run.sh" make WHAT="$*" 2>&1 | dind::filter-make-output
  else
    dind::step "+ [using the build container] make WHAT=\"$*\""
    dind::run-build-command make WHAT="$*" 2>&1 | dind::filter-make-output
  fi
}

function dind::check-binary {
  local filename="$1"
  local dockerized="_output/dockerized/bin/linux/amd64/${filename}"
  local plain="_output/local/bin/linux/amd64/${filename}"
  dind::set-build-volume-args
  # FIXME: don't hardcode amd64 arch
  if [ -n "${KUBEADM_DIND_LOCAL:-${force_local:-}}" ]; then
    if [ -f "${dockerized}" -o -f "${plain}" ]; then
      return 0
    fi
  elif docker run --rm "${build_volume_args[@]}" \
              "${busybox_image}" \
              test -f "/go/src/k8s.io/kubernetes/${dockerized}" >&/dev/null; then
    return 0
  fi
  return 1
}

function dind::ensure-downloaded-kubectl {
  local kubectl_url
  local kubectl_sha1
  local kubectl_sha1_linux
  local kubectl_sha1_darwin
  local kubectl_link
  local kubectl_os
  local full_kubectl_version

  case "${LOCAL_KUBECTL_VERSION}" in
    v1.8)
      full_kubectl_version=v1.8.15
      kubectl_sha1_linux=52a1ee321e1e8c0ecfd6e83c38bf972c2c60adf2
      kubectl_sha1_darwin=ac3f823d7aa104237929a1e35ea400c6aa3cc356
      ;;
    v1.9)
      full_kubectl_version=v1.9.9
      kubectl_sha1_linux=c8163a6360119c56d163fbd8cef8727e9841e712
      kubectl_sha1_darwin=09585552eb7616954481789489ec382c633a0162
      ;;
    v1.10)
      full_kubectl_version=v1.10.5
      kubectl_sha1_linux=dbe431b2684f8ff4188335b3b3cea185d5a9ec44
      kubectl_sha1_darwin=08e58440949c71053b45bfadf80532ea3d752d12
      ;;
    v1.11)
      full_kubectl_version=v1.11.0
      kubectl_sha1_linux=e23f251ca0cb848802f3cb0f69a4ba297d07bfc6
      kubectl_sha1_darwin=6eff29a328c4bc00879fd6a0c8b33690c6f75908
      ;;
    "")
      return 0
      ;;
    *)
      echo "Invalid kubectl version" >&2
      exit 1
  esac

  export PATH="${KUBECTL_DIR}:$PATH"

  if [ $(uname) = Darwin ]; then
    kubectl_sha1="${kubectl_sha1_darwin}"
    kubectl_os=darwin
  else
    kubectl_sha1="${kubectl_sha1_linux}"
    kubectl_os=linux
  fi
  local link_target="kubectl-${full_kubectl_version}"
  local link_name="${KUBECTL_DIR}"/kubectl
  if [[ -h "${link_name}" && "$(readlink "${link_name}")" = "${link_target}" ]]; then
    return 0
  fi

  local path="${KUBECTL_DIR}/${link_target}"
  if [[ ! -f "${path}" ]]; then
    mkdir -p "${KUBECTL_DIR}"
    curl -sSLo "${path}" "https://storage.googleapis.com/kubernetes-release/release/${full_kubectl_version}/bin/${kubectl_os}/amd64/kubectl"
    echo "${kubectl_sha1}  ${path}" | sha1sum -c
    chmod +x "${path}"
  fi

  ln -fs "${link_target}" "${KUBECTL_DIR}/kubectl"
}

function dind::ensure-kubectl {
  if [[ ! ${use_k8s_source} ]]; then
    # already checked on startup
    dind::ensure-downloaded-kubectl
    return 0
  fi
  if [ $(uname) = Darwin ]; then
    if [ ! -f _output/local/bin/darwin/amd64/kubectl ]; then
      dind::step "Building kubectl"
      dind::step "+ make WHAT=cmd/kubectl"
      make WHAT=cmd/kubectl 2>&1 | dind::filter-make-output
    fi
  elif ! force_local=y dind::check-binary kubectl; then
    dind::make-for-linux y cmd/kubectl
  fi
}

function dind::ensure-binaries {
  local -a to_build=()
  for name in "$@"; do
    if ! dind::check-binary "$(basename "${name}")"; then
      to_build+=("${name}")
    fi
  done
  if [ "${#to_build[@]}" -gt 0 ]; then
    dind::make-for-linux n "${to_build[@]}"
  fi
  return 0
}

# dind::ensure-network creates the management network for the cluster. For IPv4
# only it will have the management network CIDR. For IPv6 only, it will have
# the IPv6 management network CIDR and the NAT64 V4 mapping network CIDR. For
# dual stack, it will have the IPv4 and IPv6 management CIDRs. Each of the
# management networks (not the NAT64 network) will have a gateway specified.
#
function dind::ensure-network {
  if ! docker network inspect $(dind::net-name) >&/dev/null; then
    local -a args
    for cidr in "${mgmt_net_cidrs[@]}"; do
      if [[ $( dind::family-for ${cidr} ) = "ipv6" ]]; then
                args+=(--ipv6)
        fi
        args+=(--subnet="${cidr}")
        local gw=$( dind::make-ip-from-cidr ${cidr} 1 )
        args+=(--gateway="${gw}")
      done
        if [[ ${IP_MODE} = "ipv6" ]]; then
            # Need second network for NAT64 V4 mapping network
            args+=(--subnet=${NAT64_V4_SUBNET_PREFIX}.0.0/16)
        fi
        docker network create ${args[@]} $(dind::net-name) >/dev/null
  fi
}

function dind::ensure-volume {
  local reuse_volume=
  if [[ $1 = -r ]]; then
    reuse_volume=1
    shift
  fi
  local name="$1"
  if dind::volume-exists "${name}"; then
    if [[ ! ${reuse_volume} ]]; then
      docker volume rm "${name}" >/dev/null
    fi
  fi
  dind::create-volume "${name}"
}

function dind::ensure-dns {
    if [[ ${IP_MODE} = "ipv6" ]]; then
	local dns64_name="bind9$( dind::cluster-suffix )"
        if ! docker inspect ${dns64_name} >&/dev/null; then
            local force_dns64_for=""
            if [[ ! ${DIND_ALLOW_AAAA_USE} ]]; then
                # Normally, if have an AAAA record, it is used. This clause tells
                # bind9 to do ignore AAAA records for the specified networks
                # and/or addresses and lookup A records and synthesize new AAAA
                # records. In this case, we select "any" networks that have AAAA
                # records meaning we ALWAYS use A records and do NAT64.
                force_dns64_for="exclude { any; };"
            fi
            read -r -d '' bind9_conf <<BIND9_EOF
options {
    directory "/var/bind";
    allow-query { any; };
    forwarders {
        ${DNS64_PREFIX}${REMOTE_DNS64_V4SERVER};
    };
    auth-nxdomain no;    # conform to RFC1035
    listen-on-v6 { any; };
    dns64 ${DNS64_PREFIX_CIDR} {
        ${force_dns64_for}
    };
};
BIND9_EOF
            docker run -d --name ${dns64_name} --hostname ${dns64_name} --net "$(dind::net-name)" --label "dind-support$( dind::cluster-suffix )" \
               --sysctl net.ipv6.conf.all.disable_ipv6=0 --sysctl net.ipv6.conf.all.forwarding=1 \
               --privileged=true --ip6 ${dns_server} --dns ${dns_server} \
               -e bind9_conf="${bind9_conf}" \
               diverdane/bind9:latest /bin/sh -c 'echo "${bind9_conf}" >/named.conf && named -c /named.conf -g -u named' >/dev/null
            ipv4_addr="$(docker exec ${dns64_name} ip addr list eth0 | grep "inet" | awk '$1 == "inet" {print $2}')"
            docker exec ${dns64_name} ip addr del ${ipv4_addr} dev eth0
            docker exec ${dns64_name} ip -6 route add ${DNS64_PREFIX_CIDR} via ${LOCAL_NAT64_SERVER}
        fi
    fi
}

function dind::ensure-nat {
    if [[  ${IP_MODE} = "ipv6" ]]; then
	local nat64_name="tayga$( dind::cluster-suffix )"
        if ! docker ps | grep ${nat64_name} >&/dev/null; then
            docker run -d --name ${nat64_name} --hostname ${nat64_name} --net "$(dind::net-name)" --label "dind-support$( dind::cluster-suffix )" \
		   --sysctl net.ipv6.conf.all.disable_ipv6=0 --sysctl net.ipv6.conf.all.forwarding=1 \
		   --privileged=true --ip ${NAT64_V4_SUBNET_PREFIX}.0.200 --ip6 ${LOCAL_NAT64_SERVER} --dns ${REMOTE_DNS64_V4SERVER} --dns ${dns_server} \
		   -e TAYGA_CONF_PREFIX=${DNS64_PREFIX_CIDR} -e TAYGA_CONF_IPV4_ADDR=${NAT64_V4_SUBNET_PREFIX}.0.200 \
		   -e TAYGA_CONF_DYNAMIC_POOL=${NAT64_V4_SUBNET_PREFIX}.0.128/25 danehans/tayga:latest >/dev/null
	    # Need to check/create, as "clean" may remove route
	    local route="$(ip route | egrep "^${NAT64_V4_SUBNET_PREFIX}.0.128/25")"
	    if [[ -z "${route}" ]]; then
	        docker run --net=host --rm --privileged ${busybox_image} ip route add ${NAT64_V4_SUBNET_PREFIX}.0.128/25 via ${NAT64_V4_SUBNET_PREFIX}.0.200
	    fi
	fi
    fi
}

function dind::run {
  local reuse_volume=
  if [[ $1 = -r ]]; then
    reuse_volume="-r"
    shift
  fi
  local container_name="${1:-}"
  local node_id=${2:-0}
  local portforward="${3:-}"
  if [[ $# -gt 3 ]]; then
    shift 3
  else
    shift $#
  fi

  local -a opts=("$@")
  local ip_mode="--ip"
  for cidr in "${mgmt_net_cidrs[@]}"; do
    if [[ $( dind::family-for ${cidr} ) = "ipv6" ]]; then
      ip_mode="--ip6"
    fi
    opts+=("${ip_mode}" "$( dind::make-ip-from-cidr ${cidr} $((${node_id}+1)) )")
  done
  opts+=("$@")
  
  local -a args=("systemd.setenv=CNI_PLUGIN=${CNI_PLUGIN}")
  args+=("systemd.setenv=IP_MODE=${IP_MODE}")
  args+=("systemd.setenv=DIND_STORAGE_DRIVER=${DIND_STORAGE_DRIVER}")

  if [[ ${IP_MODE} != "ipv4" ]]; then
    opts+=(--sysctl net.ipv6.conf.all.disable_ipv6=0)
    opts+=(--sysctl net.ipv6.conf.all.forwarding=1)
  fi

  if [[ ${IP_MODE} = "ipv6" ]]; then
    opts+=(--dns ${dns_server})
    args+=("systemd.setenv=DNS64_PREFIX_CIDR=${DNS64_PREFIX_CIDR}")
    args+=("systemd.setenv=LOCAL_NAT64_SERVER=${LOCAL_NAT64_SERVER}")

    # For prefix, if node ID will be in the upper byte, push it over
    if [[ $((${POD_NET_SIZE} % 16)) -ne 0 ]]; then
      node_id=$(printf "%02x00\n" "${node_id}")
    else
      if [[ "${POD_NET_PREFIX: -1}" = ":" ]]; then
	node_id=$(printf "%x\n" "${node_id}")
      else
        node_id=$(printf "%02x\n" "${node_id}")  # In lower byte, so ensure two chars
      fi
    fi
  fi

  if [[ ${POD_NET_PREFIX} ]]; then
    args+=("systemd.setenv=POD_NET_PREFIX=${POD_NET_PREFIX}${node_id}")
  fi
  args+=("systemd.setenv=POD_NET_SIZE=${POD_NET_SIZE}")
  args+=("systemd.setenv=USE_HAIRPIN=${USE_HAIRPIN}")
  args+=("systemd.setenv=DNS_SVC_IP=${DNS_SVC_IP}")
  args+=("systemd.setenv=DNS_SERVICE=${DNS_SERVICE}")
  if [[ ! "${container_name}" ]]; then
    echo >&2 "Must specify container name"
    exit 1
  fi

  # remove any previously created containers with the same name
  docker rm -vf "${container_name}" >&/dev/null || true

  if [[ "${portforward}" ]]; then
    IFS=';' read -ra array <<< "${portforward}"
    for element in "${array[@]}"; do
      opts+=(-p "${element}")
    done
  fi

  opts+=(${sys_volume_args[@]+"${sys_volume_args[@]}"})

  dind::step "Starting DIND container:" "${container_name}"

  if [[ ! ${using_linuxkit} ]]; then
    opts+=(-v /boot:/boot -v /lib/modules:/lib/modules)
  fi

  if [[ ${ENABLE_CEPH} ]]; then
    opts+=(-v /dev:/dev
           -v /sys/bus:/sys/bus
           -v /var/run/docker.sock:/opt/outer-docker.sock)
  fi

  local volume_name="kubeadm-dind-${container_name}"
  dind::ensure-network
  dind::ensure-volume ${reuse_volume} "${volume_name}"
  dind::ensure-nat
  dind::ensure-dns

  # TODO: create named volume for binaries and mount it to /k8s
  # in case of the source build

  # Start the new container.
  docker run \
	 -e IP_MODE="${IP_MODE}" \
         -e KUBEADM_SOURCE="${KUBEADM_SOURCE}" \
         -e HYPERKUBE_SOURCE="${HYPERKUBE_SOURCE}" \
         -d --privileged \
         --net "$(dind::net-name)" \
         --name "${container_name}" \
         --hostname "${container_name}" \
         -l "${DIND_LABEL}" \
         -v "${volume_name}:/dind" \
         ${opts[@]+"${opts[@]}"} \
         "${DIND_IMAGE}" \
         ${args[@]+"${args[@]}"}
}

function dind::kubeadm {
  local container_id="$1"
  shift
  dind::step "Running kubeadm:" "$*"
  status=0
  # See image/bare/wrapkubeadm.
  # Capturing output is necessary to grab flags for 'kubeadm join'
  kubelet_feature_gates="-e KUBELET_FEATURE_GATES=${KUBELET_FEATURE_GATES}"
  if ! docker exec ${kubelet_feature_gates} "${container_id}" /usr/local/bin/wrapkubeadm "$@" 2>&1 | tee /dev/fd/2; then
    echo "*** kubeadm failed" >&2
    return 1
  fi
  return ${status}
}

# function dind::bare {
#   local container_name="${1:-}"
#   if [[ ! "${container_name}" ]]; then
#     echo >&2 "Must specify container name"
#     exit 1
#   fi
#   shift
#   run_opts=(${@+"$@"})
#   dind::run "${container_name}"
# }

function dind::configure-kubectl {
  dind::step "Setting cluster config"
  local host="$(dind::localhost)"
  if [[ -z "$using_local_linuxdocker" ]]; then
    host="127.0.0.1"
  fi
  local context_name cluster_name
  context_name="$(dind::context-name)"
  cluster_name="$(dind::context-name)"
  "${kubectl}" config set-cluster "$cluster_name" \
    --server="http://${host}:$(dind::apiserver-port)" \
    --insecure-skip-tls-verify=true
  "${kubectl}" config set-context "$context_name" --cluster="$cluster_name"
  if [[ ${DIND_LABEL} = ${DEFAULT_DIND_LABEL} ]]; then
      # Single cluster mode
      "${kubectl}" config use-context "$context_name"
  fi
}

force_make_binaries=
function dind::set-master-opts {
  master_opts=()
  if [[ ${BUILD_KUBEADM} || ${BUILD_HYPERKUBE} ]]; then
    # share binaries pulled from the build container between nodes
    local dind_k8s_bin_vol_name
    dind_k8s_bin_vol_name="dind-k8s-binaries$(dind::cluster-suffix)"
    dind::ensure-volume -r "${dind_k8s_bin_vol_name}"
    dind::set-build-volume-args
    master_opts+=("${build_volume_args[@]}" -v "${dind_k8s_bin_vol_name}:/k8s")
    local -a bins
    if [[ ${BUILD_KUBEADM} ]]; then
      master_opts+=(-e KUBEADM_SOURCE=build://)
      bins+=(cmd/kubeadm)
    else
      master_opts+=(-e ${KUBEADM_SOURCE})
    fi
    if [[ ${BUILD_HYPERKUBE} ]]; then
      master_opts+=(-e HYPERKUBE_SOURCE=build://)
      bins+=(cmd/hyperkube)
    fi
    if [[ ${force_make_binaries} ]]; then
      dind::make-for-linux n "${bins[@]}"
    else
      dind::ensure-binaries "${bins[@]}"
    fi
  fi
  if [[ ${MASTER_EXTRA_OPTS:-} ]]; then
    master_opts+=( ${MASTER_EXTRA_OPTS} )
  fi
}

function dind::ensure-dashboard-clusterrolebinding {
  local ctx
  ctx="$(dind::context-name)"
  # 'create' may cause etcd timeout, yet create the clusterrolebinding.
  # So use 'apply' to actually create it
  "${kubectl}" --context "$ctx" create clusterrolebinding add-on-cluster-admin \
               --clusterrole=cluster-admin \
               --serviceaccount=kube-system:default \
               -o json --dry-run |
    docker exec -i "$(dind::master-name)" jq '.apiVersion="rbac.authorization.k8s.io/v1beta1"|.kind|="ClusterRoleBinding"' |
    "${kubectl}" --context "$ctx" apply -f -
}

function dind::deploy-dashboard {
  dind::step "Deploying k8s dashboard"
  dind::retry "${kubectl}" --context "$(dind::context-name)" apply -f "${DASHBOARD_URL}"
  # https://kubernetes-io-vnext-staging.netlify.com/docs/admin/authorization/rbac/#service-account-permissions
  # Thanks @liggitt for the hint
  dind::retry dind::ensure-dashboard-clusterrolebinding
}

function dind::kubeadm-version {
  if [[ ${use_k8s_source} ]]; then
    (cluster/kubectl.sh version --short 2>/dev/null || true) |
      grep Client |
      sed 's/^.*: v\([0-9.]*\).*/\1/'
  else
    docker exec "$(dind::master-name)" \
           /bin/bash -c 'kubeadm version -o json | jq -r .clientVersion.gitVersion' |
      sed 's/^v\([0-9.]*\).*/\1/'
  fi
}

function dind::kubeadm-skip-checks-flag {
  kubeadm_version="$(dind::kubeadm-version)"
  if [[ ${kubeadm_version} =~ 1\.8\. ]]; then
    echo -n "--skip-preflight-checks"
  else
    echo -n "--ignore-preflight-errors=all"
  fi
}

function dind::init {
  local -a opts
  dind::set-master-opts
  local local_host master_name container_id
  master_name="$(dind::master-name)"
  local_host="$( dind::localhost )"
  container_id=$(dind::run "${master_name}" 1 "${local_host}:$(dind::apiserver-port):${INTERNAL_APISERVER_PORT}" ${master_opts[@]+"${master_opts[@]}"})
  # FIXME: I tried using custom tokens with 'kubeadm ex token create' but join failed with:
  # 'failed to parse response as JWS object [square/go-jose: compact JWS format must have three parts]'
  # So we just pick the line from 'kubeadm init' output
  # Using a template file in the image to build a kubeadm.conf file and to customize
  # it based on CNI plugin, IP mode, and environment settings. User can add additional
  # customizations to template and then rebuild the image used (build/build-local.sh).
  local pod_subnet_disable="# "
  # TODO: May want to specify each of the plugins that require --pod-network-cidr
  if [[ ${CNI_PLUGIN} != "bridge" && ${CNI_PLUGIN} != "ptp" ]]; then
    pod_subnet_disable=""
  fi
  local bind_address="0.0.0.0"
  if [[ ${IP_MODE} = "ipv6" ]]; then
    bind_address="::"
  fi
  dind::proxy "$master_name"
  dind::custom-docker-opts "$master_name"

  # HACK: Indicating mode, so that wrapkubeadm will not set a cluster CIDR for kube-proxy
  # in IPv6 (only) mode.
  if [[ ${IP_MODE} = "ipv6" ]]; then
    docker exec --privileged -i "$master_name" touch /v6-mode
  fi

  feature_gates="{}"
  if [[ ${DNS_SERVICE} == "coredns" ]]; then
    # can't just use 'CoreDNS: false' because
    # it'll break k8s 1.8. FIXME: simplify
    # after 1.8 support is removed
    feature_gates="{CoreDNS: true}"
  elif docker exec "$master_name" kubeadm init --help 2>&1 | grep -q CoreDNS; then
    # FIXME: CoreDNS should be the default in 1.11
    feature_gates="{CoreDNS: false}"
  fi

  component_feature_gates=""
  if [ "${FEATURE_GATES}" != "none" ]; then
    component_feature_gates="feature-gates: \\\"${FEATURE_GATES}\\\""
  fi

  apiserver_extra_args=""
  for e in $(set -o posix ; set | grep -E "^APISERVER_[a-z_]+=" | cut -d'=' -f 1); do
    opt_name=$(echo ${e#APISERVER_} | sed 's/_/-/g')
    apiserver_extra_args+="  ${opt_name}: \\\"$(eval echo \$$e)\\\"\\n"
  done

  controller_manager_extra_args=""
  for e in $(set -o posix ; set | grep -E "^CONTROLLER_MANAGER_[a-z_]+=" | cut -d'=' -f 1); do
    opt_name=$(echo ${e#CONTROLLER_MANAGER_} | sed 's/_/-/g')
    controller_manager_extra_args+="  ${opt_name}: \\\"$(eval echo \$$e)\\\"\\n"
  done

  scheduler_extra_args=""
  for e in $(set -o posix ; set | grep -E "^SCHEDULER_[a-z_]+=" | cut -d'=' -f 1); do
    opt_name=$(echo ${e#SCHEDULER_} | sed 's/_/-/g')
    scheduler_extra_args+="  ${opt_name}: \\\"$(eval echo \$$e)\\\"\\n"
  done

  kubeadm_version="$(dind::kubeadm-version)"
  api_version="kubeadm.k8s.io/v1alpha3"
  kind="ClusterConfiguration"
  api_endpoint="apiEndpoint:"
  if [[ ${kubeadm_version} =~ 1\.(8|9|10)\. ]]; then
    api_version="kubeadm.k8s.io/v1alpha1"
    kind="MasterConfiguration"
    api_endpoint="api:"
  elif [[ ${kubeadm_version} =~ 1\.(11|12)\. ]]; then
    api_version="kubeadm.k8s.io/v1alpha2"
    kind="MasterConfiguration"
    api_endpoint="api:"
  fi
  local mgmt_cidr=${mgmt_net_cidrs[0]}
  if [[ ${IP_MODE} = "dual-stack" && $( dind::family-for ${SERVICE_CIDR} ) = "ipv6" ]]; then
      mgmt_cidr=${mgmt_net_cidrs[1]}
  fi
  local master_ip=$( dind::make-ip-from-cidr ${mgmt_cidr} 2 )
  docker exec -i "$master_name" bash <<EOF
sed -e "s|{{API_VERSION}}|${api_version}|" \
    -e "s|{{ADV_ADDR}}|${master_ip}|" \
    -e "s|{{POD_SUBNET_DISABLE}}|${pod_subnet_disable}|" \
    -e "s|{{POD_NETWORK_CIDR}}|${POD_NETWORK_CIDR}|" \
    -e "s|{{SVC_SUBNET}}|${SERVICE_CIDR}|" \
    -e "s|{{BIND_ADDR}}|${bind_address}|" \
    -e "s|{{BIND_PORT}}|${INTERNAL_APISERVER_PORT}|" \
    -e "s|{{FEATURE_GATES}}|${feature_gates}|" \
    -e "s|{{KUBEADM_VERSION}}|${kubeadm_version}|" \
    -e "s|{{COMPONENT_FEATURE_GATES}}|${component_feature_gates}|" \
    -e "s|{{APISERVER_EXTRA_ARGS}}|${apiserver_extra_args}|" \
    -e "s|{{CONTROLLER_MANAGER_EXTRA_ARGS}}|${controller_manager_extra_args}|" \
    -e "s|{{SCHEDULER_EXTRA_ARGS}}|${scheduler_extra_args}|" \
    -e "s|{{KUBE_MASTER_NAME}}|${master_name}|" \
    -e "s|{{KUBEADM_KIND}}|${kind}|" \
    -e "s|{{API_ENDPOINT}}|${api_endpoint}|" \
    /etc/kubeadm.conf.tmpl > /etc/kubeadm.conf
EOF
  init_args=(--config /etc/kubeadm.conf)
  skip_preflight_arg="$(dind::kubeadm-skip-checks-flag)"
  # required when building from source
  if [[ ${BUILD_KUBEADM} || ${BUILD_HYPERKUBE} ]]; then
    docker exec "$master_name" mount --make-shared /k8s
  fi
  kubeadm_join_flags="$(dind::kubeadm "${container_id}" init "${init_args[@]}" "${skip_preflight_arg}" "$@" | grep '^ *kubeadm join' | sed 's/^ *kubeadm join //')"
  dind::configure-kubectl
  dind::start-port-forwarder
}

function dind::create-node-container {
  local reuse_volume next_node_index node_name
  reuse_volume=''
  if [[ ${1:-} = -r ]]; then
    reuse_volume="-r"
    shift
  fi
  # if there's just one node currently, it's master, thus we need to use
  # kube-node-1 hostname, if there are two nodes, we should pick
  # kube-node-2 and so on
  next_node_index=${1:-$(docker ps -q --filter=label="${DIND_LABEL}" | wc -l | sed 's/^ *//g')}
  local -a opts
  if [[ ${BUILD_KUBEADM} || ${BUILD_HYPERKUBE} ]]; then
    opts+=(-v "dind-k8s-binaries$(dind::cluster-suffix)":/k8s)
    if [[ ${BUILD_KUBEADM} ]]; then
      opts+=(-e KUBEADM_SOURCE=build://)
    fi
    if [[ ${BUILD_HYPERKUBE} ]]; then
      opts+=(-e HYPERKUBE_SOURCE=build://)
    fi
  fi
  node_name="$(dind::node-name ${next_node_index})"
  dind::run ${reuse_volume} "$node_name" $((next_node_index + 1)) "${EXTRA_PORTS}" ${opts[@]+"${opts[@]}"}
}

function dind::join {
  local container_id="$1"
  shift
  dind::proxy "${container_id}"
  dind::custom-docker-opts "${container_id}"
  skip_preflight_arg="$(dind::kubeadm-skip-checks-flag)"
  dind::kubeadm "${container_id}" join "${skip_preflight_arg}" "$@" >/dev/null
}

function dind::escape-e2e-name {
    sed 's/[]\$*.^()[]/\\&/g; s/\s\+/\\s+/g' <<< "$1" | tr -d '\n'
}

function dind::accelerate-kube-dns {
  if [[ ${DNS_SERVICE} == "kube-dns" ]]; then
     dind::step "Patching kube-dns deployment to make it start faster"
     # Could do this on the host, too, but we don't want to require jq here
     # TODO: do this in wrapkubeadm
     docker exec "$(dind::master-name)" /bin/bash -c \
        "kubectl get deployment kube-dns -n kube-system -o json | jq '.spec.template.spec.containers[0].readinessProbe.initialDelaySeconds = 3|.spec.template.spec.containers[0].readinessProbe.periodSeconds = 3' | kubectl apply --force -f -"
 fi
}

function dind::component-ready {
  local label="$1"
  local out
  if ! out="$("${kubectl}" --context "$(dind::context-name)" get pod -l "${label}" -n kube-system \
                           -o jsonpath='{ .items[*].status.conditions[?(@.type == "Ready")].status }' 2>/dev/null)"; then
    return 1
  fi
  if ! grep -v False <<<"${out}" | grep -q True; then
    return 1
  fi
  return 0
}

function dind::kill-failed-pods {
  local pods ctx
  ctx="$(dind::context-name)"
  # workaround for https://github.com/kubernetes/kubernetes/issues/36482
  if ! pods="$(kubectl --context "$ctx" get pod -n kube-system -o jsonpath='{ .items[?(@.status.phase == "Failed")].metadata.name }' 2>/dev/null)"; then
    return
  fi
  for name in ${pods}; do
    kubectl --context "$ctx" delete pod --now -n kube-system "${name}" >&/dev/null || true
  done
}

function dind::create-static-routes {
  echo "Creating static routes for bridge/PTP plugin"
  for ((i=0; i <= NUM_NODES; i++)); do
    if [[ ${i} -eq 0 ]]; then
      node="$(dind::master-name)"
    else
      node="$(dind::node-name $i)"
    fi
    for ((j=0; j <= NUM_NODES; j++)); do
      if [[ ${i} -eq ${j} ]]; then
	continue
      fi
      if [[ ${j} -eq 0 ]]; then
        dest_node="$(dind::master-name)"
      else
        dest_node="$(dind::node-name $j)"
      fi
      id=$((${j}+1))
      if [[ ${IP_MODE} = "ipv4" ]]; then
	# Assuming pod subnets will all be /24
        dest="${POD_NET_PREFIX}${id}.0/24"
        gw=`docker exec ${dest_node} ip addr show eth0 | grep -w inet | awk '{ print $2 }' | sed 's,/.*,,'`
      else
	instance=$(printf "%02x" ${id})
	if [[ $((${POD_NET_SIZE} % 16)) -ne 0 ]]; then
	  instance+="00" # Move node ID to upper byte
	fi
	dest="${POD_NET_PREFIX}${instance}::/${POD_NET_SIZE}"
        gw=`docker exec ${dest_node} ip addr show eth0 | grep -w inet6 | grep -i global | head -1 | awk '{ print $2 }' | sed 's,/.*,,'`
      fi
      docker exec "${node}" ip route add "${dest}" via "${gw}"
    done
  done
}

# If we are allowing AAAA record use, then provide SNAT for IPv6 packets from
# node containers, and forward packets to bridge used for $(dind::net-name).
# This gives pods access to external IPv6 sites, when using IPv6 addresses.
function dind::setup_external_access_on_host {
  if [[ ! ${DIND_ALLOW_AAAA_USE} ]]; then
    return
  fi
  local main_if=`ip route | grep default | awk '{print $5}'`
  local bridge_if=`ip route | grep ${NAT64_V4_SUBNET_PREFIX}.0.0 | awk '{print $3}'`
  dind::ip6tables-on-hostnet -t nat -A POSTROUTING -o $main_if -j MASQUERADE
  if [[ -n "$bridge_if" ]]; then
    dind::ip6tables-on-hostnet -A FORWARD -i $bridge_if -j ACCEPT
  else
    echo "WARNING! No $(dind::net-name) bridge - unable to setup forwarding/SNAT"
  fi
}

# Remove ip6tables rules for SNAT and forwarding, if they exist.
function dind::remove_external_access_on_host {
  if [[ ! ${DIND_ALLOW_AAAA_USE} ]]; then
    return
  fi
  local have_rule
  local main_if="$(ip route | grep default | awk '{print $5}')"
  local bridge_if="$(ip route | grep ${NAT64_V4_SUBNET_PREFIX}.0.0 | awk '{print $3}')"

  have_rule="$(dind::ip6tables-on-hostnet -S -t nat | grep "\-o $main_if" || true)"
  if [[ -n "$have_rule" ]]; then
    dind::ip6tables-on-hostnet -t nat -D POSTROUTING -o $main_if -j MASQUERADE
  else
    echo "Skipping delete of ip6tables rule for SNAT, as rule non-existent"
  fi
  if [[ -n "$bridge_if" ]]; then
    have_rule="$(dind::ip6tables-on-hostnet -S | grep "\-i $bridge_if" || true)"
    if [[ -n "$have_rule" ]]; then
      dind::ip6tables-on-hostnet -D FORWARD -i $bridge_if -j ACCEPT
    else
      echo "Skipping delete of ip6tables rule for forwarding, as rule non-existent"
    fi
  else
    echo "Skipping delete of ip6tables rule for forwarding, as no bridge interface"
  fi
}

function dind::ip6tables-on-hostnet {
  local mod_path='/lib/modules'
  docker run -v "${mod_path}:${mod_path}" --entrypoint /sbin/ip6tables --net=host --rm --privileged "${DIND_IMAGE}" "$@"
}

function dind::wait-for-ready {
  local app="kube-proxy"
  if [[ ${CNI_PLUGIN} = "kube-router" ]]; then
    app=kube-router
  fi
  dind::step "Waiting for ${app} and the nodes"
  local app_ready
  local nodes_ready
  local n=3
  local ntries=200
  local ctx
  ctx="$(dind::context-name)"
  while true; do
    dind::kill-failed-pods
    if "${kubectl}" --context "$ctx" get nodes 2>/dev/null | grep -q NotReady; then
      nodes_ready=
    else
      nodes_ready=y
    fi
    if dind::component-ready k8s-app=${app}; then
      app_ready=y
    else
      app_ready=
    fi
    if [[ ${nodes_ready} && ${app_ready} ]]; then
      if ((--n == 0)); then
        echo "[done]" >&2
        break
      fi
    else
      n=3
    fi
    if ((--ntries == 0)); then
      echo "Error waiting for ${app} and the nodes" >&2
      exit 1
    fi
    echo -n "." >&2
    sleep 1
  done

  dind::step "Bringing up ${DNS_SERVICE} and kubernetes-dashboard"
  # on Travis 'scale' sometimes fails with 'error: Scaling the resource failed with: etcdserver: request timed out; Current resource version 442' here
  dind::retry "${kubectl}" --context "$ctx" scale deployment --replicas=1 -n kube-system ${DNS_SERVICE}
  dind::retry "${kubectl}" --context "$ctx" scale deployment --replicas=1 -n kube-system kubernetes-dashboard

  ntries=200
  while ! dind::component-ready k8s-app=kube-dns || ! dind::component-ready app=kubernetes-dashboard; do
    if ((--ntries == 0)); then
      echo "Error bringing up ${DNS_SERVICE} and kubernetes-dashboard" >&2
      exit 1
    fi
    echo -n "." >&2
    dind::kill-failed-pods
    sleep 1
  done
  echo "[done]" >&2

  dind::retry "${kubectl}" --context "$ctx" get nodes >&2

  local local_host
  local_host="$( dind::localhost )"
  dind::step "Access dashboard at:" "http://${local_host}:$(dind::apiserver-port)/api/v1/namespaces/kube-system/services/kubernetes-dashboard:/proxy"
}

function dind::up {
  dind::down
  dind::init
  local ctx
  ctx="$(dind::context-name)"
  # pre-create node containers sequentially so they get predictable IPs
  local -a node_containers
  for ((n=1; n <= NUM_NODES; n++)); do
    dind::step "Starting node container:" ${n}
    if ! container_id="$(dind::create-node-container ${n})"; then
      echo >&2 "*** Failed to start node container ${n}"
      exit 1
    else
      node_containers+=(${container_id})
      dind::step "Node container started:" ${n}
    fi
  done
  dind::fix-mounts
  status=0
  local -a pids
  for ((n=1; n <= NUM_NODES; n++)); do
    (
      dind::step "Joining node:" ${n}
      container_id="${node_containers[${n}-1]}"
      if ! dind::join ${container_id} ${kubeadm_join_flags}; then
        echo >&2 "*** Failed to start node container ${n}"
        exit 1
      else
        dind::step "Node joined:" ${n}
      fi
    )&
    pids[${n}]=$!
  done
  if ((NUM_NODES > 0)); then
    for pid in ${pids[*]}; do
      wait ${pid}
    done
  else
    # FIXME: this may fail depending on k8s/kubeadm version
    # FIXME: check for taint & retry if it's there
    "${kubectl}" --context "$ctx" taint nodes $(dind::master-name) node-role.kubernetes.io/master- || true
  fi
  case "${CNI_PLUGIN}" in
    bridge | ptp)
      dind::create-static-routes
      dind::setup_external_access_on_host
      ;;
    flannel)
      # without --validate=false this will fail on older k8s versions
      dind::retry "${kubectl}" --context "$ctx" apply --validate=false -f "https://github.com/coreos/flannel/blob/master/Documentation/kube-flannel.yml?raw=true"
      ;;
    calico)
      dind::retry "${kubectl}" --context "$ctx" apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml
      ;;
    calico-kdd)
      dind::retry "${kubectl}" --context "$ctx" apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/rbac-kdd.yaml
      dind::retry "${kubectl}" --context "$ctx" apply -f https://docs.projectcalico.org/v2.6/getting-started/kubernetes/installation/hosted/kubernetes-datastore/calico-networking/1.7/calico.yaml
      ;;
    weave)
      dind::retry "${kubectl}" --context "$ctx" apply -f "https://github.com/weaveworks/weave/blob/master/prog/weave-kube/weave-daemonset-k8s-1.6.yaml?raw=true"
      ;;
    kube-router)
      dind::retry "${kubectl}" --context "$ctx" apply -f "https://raw.githubusercontent.com/cloudnativelabs/kube-router/master/daemonset/kubeadm-kuberouter-all-features.yaml"
      dind::retry "${kubectl}" --context "$ctx" -n kube-system delete ds kube-proxy
      docker run --privileged --net=host k8s.gcr.io/kube-proxy-amd64:v1.10.2 kube-proxy --cleanup
      ;;
    *)
      echo "Unsupported CNI plugin '${CNI_PLUGIN}'" >&2
      ;;
  esac
  dind::deploy-dashboard
  dind::accelerate-kube-dns
  if [[ (${CNI_PLUGIN} != "bridge" && ${CNI_PLUGIN} != "ptp") || ${SKIP_SNAPSHOT} ]]; then
    # This is especially important in case of Calico -
    # the cluster will not recover after snapshotting
    # (at least not after restarting from the snapshot)
    # if Calico installation is interrupted
    dind::wait-for-ready
  fi
  dind::step "Cluster Info"
  echo "Network Mode: ${IP_MODE}"
  echo "Cluster context: $( dind::context-name )"
  echo "Cluster ID: ${CLUSTER_ID}"
  echo "Management CIDR(s): ${mgmt_net_cidrs[@]}"
  echo "Service CIDR/mode: ${SERVICE_CIDR}/${SERVICE_NET_MODE}"
  echo "Pod CIDR(s): ${POD_NETWORK_CIDR}"
}

function dind::fix-mounts {
  local node_name
  for ((n=0; n <= NUM_NODES; n++)); do
    node_name="$(dind::master-name)"
    if ((n > 0)); then
      node_name="$(dind::node-name $n)"
    fi
    docker exec "${node_name}" mount --make-shared /run
    if [[ ! ${using_linuxkit} ]]; then
      docker exec "${node_name}" mount --make-shared /lib/modules/
    fi
    # required when building from source
    if [[ ${BUILD_KUBEADM} || ${BUILD_HYPERKUBE} ]]; then
      docker exec "${node_name}" mount --make-shared /k8s
    fi
    # docker exec "${node_name}" mount --make-shared /sys/kernel/debug
  done
}

function dind::snapshot_container {
  local container_name="$1"
  docker exec -i ${container_name} /usr/local/bin/snapshot prepare
  # remove the hidden *plnk directories
  docker diff ${container_name} | grep -v plnk | docker exec -i ${container_name} /usr/local/bin/snapshot save
}

function dind::snapshot {
  dind::step "Taking snapshot of the cluster"
  dind::snapshot_container "$(dind::master-name)"
  for ((n=1; n <= NUM_NODES; n++)); do
    dind::snapshot_container "$(dind::node-name $n)"
  done
  dind::wait-for-ready
}

restore_cmd=restore
function dind::restore_container {
  local container_id="$1"
  docker exec ${container_id} /usr/local/bin/snapshot "${restore_cmd}"
}

function dind::restore {
  local apiserver_port local_host pid pids
  dind::down
  dind::step "Restoring master container"
  dind::set-master-opts
  local_host="$( dind::localhost )"
  apiserver_port="$( dind::apiserver-port )"
  for ((n=0; n <= NUM_NODES; n++)); do
    (
      if [[ n -eq 0 ]]; then
        dind::step "Restoring master container"
        dind::restore_container "$(dind::run -r "$(dind::master-name)" 1 "${local_host}:${apiserver_port}:${INTERNAL_APISERVER_PORT}" ${master_opts[@]+"${master_opts[@]}"})"
        dind::step "Master container restored"
      else
        dind::step "Restoring node container:" ${n}
        if ! container_id="$(dind::create-node-container -r ${n})"; then
          echo >&2 "*** Failed to start node container ${n}"
          exit 1
        else
          dind::restore_container "${container_id}"
          dind::step "Node container restored:" ${n}
        fi
      fi
    )&
    pids[${n}]=$!
  done
  for pid in ${pids[*]}; do
    wait ${pid}
  done
  if [[ ${CNI_PLUGIN} = "bridge" || ${CNI_PLUGIN} = "ptp" ]]; then
    dind::create-static-routes
    dind::setup_external_access_on_host
  fi
  dind::fix-mounts
  # Recheck kubectl config. It's possible that the cluster was started
  # on this docker from different host
  dind::configure-kubectl
  dind::start-port-forwarder
  dind::wait-for-ready
}

function dind::down {
  dind::remove-images "${DIND_LABEL}"
  if [[ ${CNI_PLUGIN} = "bridge" || ${CNI_PLUGIN} = "ptp" ]]; then
    dind::remove_external_access_on_host
  elif [[ "${CNI_PLUGIN}" = "kube-router" ]]; then
    docker run --privileged --net=host cloudnativelabs/kube-router --cleanup-config
  fi
}

function dind::apiserver-port {
  # APISERVER_PORT is explicitely set
  if [ -n "${APISERVER_PORT:-}" ]
  then
    echo "$APISERVER_PORT"
    return
  fi

  # Get the port from the master
  local master port
  master="$(dind::master-name)"
  # 8080/tcp -> 127.0.0.1:8082  =>  8082
  port="$( docker port "$master" 2>/dev/null | awk -F: "/^${INTERNAL_APISERVER_PORT}/{ print \$NF }" )"
  if [ -n "$port" ]
  then
    APISERVER_PORT="$port"
    echo "$APISERVER_PORT"
    return
  fi

  # get a random free port
  APISERVER_PORT=0
  echo "$APISERVER_PORT"
}

function dind::master-name {
  echo "kube-master$( dind::cluster-suffix )"
}

function dind::node-name {
  local nr="$1"
  echo "kube-node-${nr}$( dind::cluster-suffix )"
}

function dind::context-name {
  echo "dind$( dind::cluster-suffix )"
}

function dind::remove-volumes {
  # docker 1.13+: docker volume ls -q -f label="${DIND_LABEL}"
  local nameRE
  nameRE="^kubeadm-dind-(sys|kube-master|kube-node-[0-9]+)$(dind::cluster-suffix)$"
  docker volume ls -q | (grep -E "$nameRE" || true) | while read -r volume_id; do
    dind::step "Removing volume:" "${volume_id}"
    docker volume rm "${volume_id}"
  done
}

function dind::remove-images {
  local which=$1
  docker ps -a -q --filter=label="${which}" | while read container_id; do
    dind::step "Removing container:" "${container_id}"
    docker rm -fv "${container_id}"
  done
}

function dind::remove-cluster {
  cluster_name="dind$(dind::cluster-suffix)"
  if ${kubectl} config get-clusters | grep -qE "^${cluster_name}$"; then
    dind::step "Removing cluster from config:" "${cluster_name}"
    ${kubectl} config delete-cluster ${cluster_name} 2>/dev/null || true
  fi
}

function dind::remove-context {
  context_name="$(dind::context-name)"
  if ${kubectl} config get-contexts | grep -qE "${context_name}\\s"; then
    dind::step "Removing context from config:" "${context_name}"
    ${kubectl} config delete-context ${context_name} 2>/dev/null || true
  fi
}

function dind::start-port-forwarder {
  local fwdr port
  fwdr="${DIND_PORT_FORWARDER:-}"

  [ -n "$fwdr" ] || return 0

  [ -x "$fwdr" ] || {
    echo "'${fwdr}' is not executable." >&2
    return 1
  }

  port="$( dind::apiserver-port )"
  dind::step "+ Setting up port-forwarding for :${port}"
  "$fwdr" "$port"
}

function dind::check-for-snapshot {
  if ! dind::volume-exists "kubeadm-dind-$(dind::master-name)"; then
    return 1
  fi
  for ((n=1; n <= NUM_NODES; n++)); do
    if ! dind::volume-exists "kubeadm-dind-$(dind::node-name ${n})"; then
      return 1
    fi
  done
}

function dind::do-run-e2e {
  local parallel="${1:-}"
  local focus="${2:-}"
  local skip="${3:-}"
  local host="$(dind::localhost)"
  if [[ -z "$using_local_linuxdocker" ]]; then
    host="127.0.0.1"
  fi  
  dind::need-source
  local kubeapi test_args term=
  local -a e2e_volume_opts=()
  kubeapi="http://${host}:$(dind::apiserver-port)"
  test_args="--host=${kubeapi}"
  if [[ ${focus} ]]; then
    test_args="--ginkgo.focus=${focus} ${test_args}"
  fi
  if [[ ${skip} ]]; then
    test_args="--ginkgo.skip=${skip} ${test_args}"
  fi
  if [[ ${E2E_REPORT_DIR} ]]; then
    test_args="--report-dir=/report ${test_args}"
    e2e_volume_opts=(-v "${E2E_REPORT_DIR}:/report")
  fi
  dind::make-for-linux n "cmd/kubectl test/e2e/e2e.test vendor/github.com/onsi/ginkgo/ginkgo"
  dind::step "Running e2e tests with args:" "${test_args}"
  dind::set-build-volume-args
  if [ -t 1 ] ; then
    term="-it"
    test_args="--ginkgo.noColor --num-nodes=2 ${test_args}"
  fi
  docker run \
         --rm ${term} \
         --net=host \
         "${build_volume_args[@]}" \
         -e KUBERNETES_PROVIDER=dind \
         -e KUBE_MASTER_IP="${kubeapi}" \
         -e KUBE_MASTER=local \
         -e KUBERNETES_CONFORMANCE_TEST=y \
         -e GINKGO_PARALLEL=${parallel} \
         ${e2e_volume_opts[@]+"${e2e_volume_opts[@]}"} \
         -w /go/src/k8s.io/kubernetes \
         "${e2e_base_image}" \
         bash -c "cluster/kubectl.sh config set-cluster dind --server='${kubeapi}' --insecure-skip-tls-verify=true &&
         cluster/kubectl.sh config set-context dind --cluster=dind &&
         cluster/kubectl.sh config use-context dind &&
         go run hack/e2e.go -- --v 6 --test --check-version-skew=false --test_args='${test_args}'"
}

function dind::clean {
  dind::down
  dind::remove-images "dind-support$( dind::cluster-suffix )"
  dind::remove-volumes
  local net_name
  net_name="$(dind::net-name)"
  if docker network inspect "$net_name" >&/dev/null; then
    docker network rm "$net_name"
  fi
  dind::remove-cluster
  dind::remove-context
}

function dind::copy-image {
  local image="${2:-}"
  local image_path="/tmp/save_${image//\//_}"
  if [[ -f "${image_path}" ]]; then
    rm -fr "${image_path}"
  fi
  docker save "${image}" -o "${image_path}"
  docker ps -a -q --filter=label="${DIND_LABEL}" | while read container_id; do
    cat "${image_path}" | docker exec -i "${container_id}" docker load
  done
  rm -fr "${image_path}"
}

function dind::run-e2e {
  local focus="${1:-}"
  local skip="${2:-[Serial]}"
  skip="$(dind::escape-e2e-name "${skip}")"
  if [[ "$focus" ]]; then
    focus="$(dind::escape-e2e-name "${focus}")"
  else
    focus="\[Conformance\]"
  fi
  local parallel=y
  if [[ ${DIND_NO_PARALLEL_E2E} ]]; then
    parallel=
  fi
  dind::do-run-e2e "${parallel}" "${focus}" "${skip}"
}

function dind::run-e2e-serial {
  local focus="${1:-}"
  local skip="${2:-}"
  skip="$(dind::escape-e2e-name "${skip}")"
  dind::need-source
  if [[ "$focus" ]]; then
    focus="$(dind::escape-e2e-name "${focus}")"
  else
    focus="\[Serial\].*\[Conformance\]"
  fi
  dind::do-run-e2e n "${focus}" "${skip}"
  # TBD: specify filter
}

function dind::step {
  local OPTS=""
  if [ "$1" = "-n" ]; then
    shift
    OPTS+="-n"
  fi
  GREEN="$1"
  shift
  if [ -t 2 ] ; then
    echo -e ${OPTS} "\x1B[97m* \x1B[92m${GREEN}\x1B[39m $*" 1>&2
  else
    echo ${OPTS} "* ${GREEN} $*" 1>&2
  fi
}

function dind::dump {
  set +e
  echo "*** Dumping cluster state ***"
  for node in $(docker ps --format '{{.Names}}' --filter label="${DIND_LABEL}"); do
    for service in kubelet.service dindnet.service criproxy.service dockershim.service; do
      if docker exec "${node}" systemctl is-enabled "${service}" >&/dev/null; then
        echo "@@@ service-${node}-${service}.log @@@"
        docker exec "${node}" systemctl status "${service}"
        docker exec "${node}" journalctl -xe -n all -u "${service}"
      fi
    done
    echo "@@@ psaux-${node}.txt @@@"
    docker exec "${node}" ps auxww
    echo "@@@ dockerps-a-${node}.txt @@@"
    docker exec "${node}" docker ps -a
    echo "@@@ ip-a-${node}.txt @@@"
    docker exec "${node}" ip a
    echo "@@@ ip-r-${node}.txt @@@"
    docker exec "${node}" ip r
  done
  local ctx master_name
  master_name="$(dind::master-name)"
  ctx="$(dind::context-name)"
  docker exec "$master_name" kubectl get pods --all-namespaces \
          -o go-template='{{range $x := .items}}{{range $x.spec.containers}}{{$x.spec.nodeName}}{{" "}}{{$x.metadata.namespace}}{{" "}}{{$x.metadata.name}}{{" "}}{{.name}}{{"\n"}}{{end}}{{end}}' |
    while read node ns pod container; do
      echo "@@@ pod-${node}-${ns}-${pod}--${container}.log @@@"
      docker exec "$master_name" kubectl logs -n "${ns}" -c "${container}" "${pod}"
    done
  echo "@@@ kubectl-all.txt @@@"
  docker exec "$master_name" kubectl get all --all-namespaces -o wide
  echo "@@@ describe-all.txt @@@"
  docker exec "$master_name" kubectl describe all --all-namespaces
  echo "@@@ nodes.txt @@@"
  docker exec "$master_name" kubectl get nodes -o wide
}

function dind::dump64 {
  echo "%%% start-base64 %%%"
  dind::dump | docker exec -i "$(dind::master-name)" /bin/sh -c "lzma | base64 -w 100"
  echo "%%% end-base64 %%%"
}

function dind::split-dump {
  mkdir -p cluster-dump
  cd cluster-dump
  awk '!/^@@@ .* @@@$/{print >out}; /^@@@ .* @@@$/{out=$2}' out=/dev/null
  ls -l
}

function dind::split-dump64 {
  decode_opt=-d
  if base64 --help | grep -q '^ *-D'; then
    # Mac OS X
    decode_opt=-D
  fi
  sed -n '/^%%% start-base64 %%%$/,/^%%% end-base64 %%%$/p' |
    sed '1d;$d' |
    base64 "${decode_opt}" |
    lzma -dc |
    dind::split-dump
}

function dind::proxy {
  local container_id="$1"
  if [[ ${DIND_CA_CERT_URL} ]] ; then
    dind::step "+ Adding certificate on ${container_id}"
    docker exec ${container_id} /bin/sh -c "cd /usr/local/share/ca-certificates; curl -sSO ${DIND_CA_CERT_URL}"
    docker exec ${container_id} update-ca-certificates
  fi
  if [[ "${DIND_PROPAGATE_HTTP_PROXY}" || "${DIND_HTTP_PROXY}" || "${DIND_HTTPS_PROXY}" || "${DIND_NO_PROXY}" ]]; then
    dind::step "+ Setting *_PROXY for docker service on ${container_id}"
    local proxy_env="[Service]"$'\n'"Environment="
    if [[ "${DIND_PROPAGATE_HTTP_PROXY}" ]]; then
      # take *_PROXY values from container environment
      proxy_env+=$(docker exec ${container_id} env | grep -i _proxy | awk '{ print "\""$0"\""}' | xargs -d'\n')
    else
      if [[ "${DIND_HTTP_PROXY}" ]] ;  then proxy_env+="\"HTTP_PROXY=${DIND_HTTP_PROXY}\" "; fi
      if [[ "${DIND_HTTPS_PROXY}" ]] ; then proxy_env+="\"HTTPS_PROXY=${DIND_HTTPS_PROXY}\" "; fi
      if [[ "${DIND_NO_PROXY}" ]] ;    then proxy_env+="\"NO_PROXY=${DIND_NO_PROXY}\" "; fi
    fi
    docker exec -i ${container_id} /bin/sh -c "cat > /etc/systemd/system/docker.service.d/30-proxy.conf" <<< "${proxy_env}"
    docker exec ${container_id} systemctl daemon-reload
    docker exec ${container_id} systemctl restart docker
  fi
}

function dind::custom-docker-opts {
  local container_id="$1"
  local -a jq=()
  if [[ ! -f ${DIND_DAEMON_JSON_FILE} ]] ; then
    jq[0]="{}"
  else
    jq+=("$(cat ${DIND_DAEMON_JSON_FILE})")
  fi
  if [[ ${DIND_REGISTRY_MIRROR} ]] ; then
    dind::step "+ Setting up registry mirror on ${container_id}"
    jq+=("{\"registry-mirrors\": [\"${DIND_REGISTRY_MIRROR}\"]}")
  fi
  if [[ ${DIND_INSECURE_REGISTRIES} ]] ; then
    dind::step "+ Setting up insecure-registries on ${container_id}"
    jq+=("{\"insecure-registries\": ${DIND_INSECURE_REGISTRIES}}")
  fi
  if [[ ${jq} ]] ; then
    local json=$(IFS="+"; echo "${jq[*]}")
    docker exec -i ${container_id} /bin/sh -c "mkdir -p /etc/docker && jq -n '${json}' > /etc/docker/daemon.json"
    docker exec ${container_id} systemctl daemon-reload
    docker exec ${container_id} systemctl restart docker
  fi
}

case "${1:-}" in
  up)
    if [[ ! ( ${DIND_IMAGE} =~ local ) && ! ${DIND_SKIP_PULL:-} ]]; then
      dind::step "Making sure DIND image is up to date"
      docker pull "${DIND_IMAGE}" >&2
    fi

    dind::prepare-sys-mounts
    dind::ensure-kubectl
    if [[ ${SKIP_SNAPSHOT} ]]; then
      force_make_binaries=y dind::up
    elif ! dind::check-for-snapshot; then
      force_make_binaries=y dind::up
      dind::snapshot
    else
      dind::restore
    fi
    ;;
  reup)
    dind::prepare-sys-mounts
    dind::ensure-kubectl
    if [[ ${SKIP_SNAPSHOT} ]]; then
      force_make_binaries=y dind::up
    elif ! dind::check-for-snapshot; then
      force_make_binaries=y dind::up
      dind::snapshot
    else
      force_make_binaries=y
      restore_cmd=update_and_restore
      dind::restore
    fi
    ;;
  down)
    dind::down
    ;;
  init)
    shift
    dind::prepare-sys-mounts
    dind::ensure-kubectl
    dind::init "$@"
    ;;
  join)
    shift
    dind::prepare-sys-mounts
    dind::ensure-kubectl
    dind::join "$(dind::create-node-container)" "$@"
    ;;
  # bare)
  #   shift
  #   dind::bare "$@"
  #   ;;
  snapshot)
    shift
    dind::snapshot
    ;;
  restore)
    shift
    dind::restore
    ;;
  clean)
    dind::clean
    ;;
  copy-image)
    dind::copy-image "$@"
    ;;
  e2e)
    shift
    dind::run-e2e "$@"
    ;;
  e2e-serial)
    shift
    dind::run-e2e-serial "$@"
    ;;
  dump)
    dind::dump
    ;;
  dump64)
    dind::dump64
    ;;
  split-dump)
    dind::split-dump
    ;;
  split-dump64)
    dind::split-dump64
    ;;
  apiserver-port)
    dind::apiserver-port
    ;;
  *)
    echo "usage:" >&2
    echo "  $0 up" >&2
    echo "  $0 reup" >&2
    echo "  $0 down" >&2
    echo "  $0 init kubeadm-args..." >&2
    echo "  $0 join kubeadm-args..." >&2
    # echo "  $0 bare container_name [docker_options...]"
    echo "  $0 clean"
    echo "  $0 copy-image [image_name]" >&2
    echo "  $0 e2e [test-name-substring]" >&2
    echo "  $0 e2e-serial [test-name-substring]" >&2
    echo "  $0 dump" >&2
    echo "  $0 dump64" >&2
    echo "  $0 split-dump" >&2
    echo "  $0 split-dump64" >&2
    exit 1
    ;;
esac
