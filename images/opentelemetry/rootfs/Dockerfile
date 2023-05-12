# Copyright 2021 The Kubernetes Authors. All rights reserved.
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


FROM alpine:3.17.2 as base

RUN mkdir -p /opt/third_party/install
COPY . /opt/third_party/

# install build tools
RUN apk update \
	&& apk upgrade \
	&& apk add -U bash cmake \
	&& bash /opt/third_party/build.sh -p

# install gRPC
FROM base as grpc
RUN bash /opt/third_party/build.sh -g v1.49.2

# install OpenTelemetry-cpp
FROM base as otel-cpp
COPY --from=grpc /opt/third_party/install/ /usr
RUN bash /opt/third_party/build.sh -o v1.8.1

# install otel_ngx_module.so
FROM base as nginx
COPY --from=grpc /opt/third_party/install/ /usr
COPY --from=otel-cpp /opt/third_party/install/ /usr
RUN bash /opt/third_party/build.sh -n

FROM alpine:3.17.2 as final
COPY --from=base /opt/third_party/init_module.sh /usr/local/bin/init_module.sh
COPY --from=nginx /etc/nginx/modules /etc/nginx/modules
COPY --from=nginx /opt/third_party/install/lib /etc/nginx/modules
