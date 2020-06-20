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

FROM alpine:3.12

ENV LC_ALL=C.UTF-8
ENV LANG=C.UTF-8

RUN apk update \
  && apk add --no-cache \
    python3 python3-dev \
    musl-dev gcc g++ make \
    libffi libffi-dev libstdc++ \
    py3-gevent py3-gunicorn py3-wheel \
    py3-pip \
 && pip3 install httpbin \
 && apk del python3-dev musl-dev gcc g++ make libffi-dev

EXPOSE 80

CMD ["gunicorn", "-b", "0.0.0.0:80", "httpbin:app", "-k", "gevent"]
