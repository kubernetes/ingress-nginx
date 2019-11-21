# Copyright 2019 The Kubernetes Authors. All rights reserved.
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

ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

WORKDIR /httpbin

RUN clean-install python3-pip curl python3-pip git bash gcc libstdc++-8-dev libpython3.7-dev python3-setuptools \
    && pip3 install --no-cache-dir httpbin \
    && pip3 install --no-cache-dir gunicorn \
    && pip3 install --no-cache-dir gevent \
    && apt remove git gcc libstdc++-8-dev libpython3.7-dev python3-setuptools --yes

EXPOSE 80

CMD ["gunicorn", "-b", "0.0.0.0:80", "httpbin:app", "-k", "gevent"]
