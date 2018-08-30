# Copyright 2018 The Kubernetes Authors. All rights reserved.
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

FROM quay.io/kubernetes-ingress-controller/nginx-ingress-controller:debug

RUN clean-install \
  dstat \
  ethstatus \
  file \
  nano \
  xz-utils \
  iperf \
  jq \
  perl-modules \
  build-essential cmake bc curl diffutils git kmod libcurl4-openssl-dev wget \
  zlib1g-dev elfutils libdw-dev gettext \
  libnss3-dev libavahi-core-dev \
  fakeroot crash makedumpfile kernel-wedge

RUN wget https://sourceware.org/systemtap/ftp/releases/systemtap-3.3.tar.gz \
  && tar -xvf systemtap-3.3.tar.gz \
  && cd systemtap-3.3/ \
  && ./configure --prefix=/opt/stap --disable-docs --disable-publican --disable-refdocs CFLAGS="-g -O2" \
  && make \
  && make install

RUN cd /opt \
  && curl -sL https://github.com/openresty/stapxx/archive/master.tar.gz | tar zxv \
  && curl -sL https://github.com/openresty/openresty-systemtap-toolkit/archive/master.tar.gz | tar zxv \
  && chmod +x /opt/openresty-systemtap-toolkit-master/*

RUN mkdir -p /var/log/nginx \
  && chmod -R 777 /var/log

ENV PATH=$PATH:/opt/stapxx-master:/opt/openresty-systemtap-toolkit-master:/opt/stap/bin
