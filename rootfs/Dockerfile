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

RUN if grep -q -i "ports.ubuntu.com" /etc/apt/sources.list; \		 +RUN clean-install \
  then \
    echo "deb http://ports.ubuntu.com/ubuntu-ports zesty universe" >> /etc/apt/sources.list; \
  else \		
    echo "deb http://archive.ubuntu.com/ubuntu zesty universe" >> /etc/apt/sources.list; \		
  fi \		
  && DEBIAN_FRONTEND=noninteractive apt-get update \		
  && apt-get install -y \		
    diffutils \		
    dumb-init \		
    --no-install-recommends \		
  && apt-get clean \		
  && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/bin/dumb-init"]

COPY . /

CMD ["/nginx-ingress-controller"]
