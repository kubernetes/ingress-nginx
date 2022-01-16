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


FROM alpine:3.14.2 as builder

COPY . /

RUN apk update \
	&& apk upgrade \
	&& apk add -U bash \
	&& /build.sh

FROM busybox:latest 

COPY --from=builder init_module.sh /usr/local/bin/init_module.sh
COPY --from=builder /etc/nginx/modules /etc/nginx/modules
