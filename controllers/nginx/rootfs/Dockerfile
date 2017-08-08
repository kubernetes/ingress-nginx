# Copyright 2015 The Kubernetes Authors. All rights reserved.
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

FROM BASEIMAGE

CROSS_BUILD_COPY qemu-QEMUARCH-static /usr/bin/

RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y \
  diffutils \
  --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

RUN curl -sSL -o /tmp/dumb-init.deb http://ftp.us.debian.org/debian/pool/main/d/dumb-init/dumb-init_1.2.0-1_DUMB_ARCH.deb && \
  dpkg -i /tmp/dumb-init.deb && \
  rm /tmp/dumb-init.deb

ENTRYPOINT ["/usr/bin/dumb-init"]

COPY . /

CMD ["/nginx-ingress-controller"]
