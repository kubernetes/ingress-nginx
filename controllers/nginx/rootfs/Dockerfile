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

FROM gcr.io/google_containers/nginx-slim:0.14

RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y \
  diffutils \
  --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

COPY . /

# https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem
RUN curl -sSL -o /sbin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.0/dumb-init_1.2.0_amd64 && \
  chmod +x /sbin/dumb-init

ENTRYPOINT ["/sbin/dumb-init", "--"]

CMD ["/nginx-ingress-controller"]
