# Copyright 2017 The Kubernetes Authors. All rights reserved.
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

FROM golang:1.20.1-alpine3.17 as builder

WORKDIR /go/src/k8s.io/ingress-nginx/images/fastcgi

COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo \
	-ldflags "-s -w" \
	-o fastcgi-helloserver main.go

# Use distroless as minimal base image to package the binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /go/src/k8s.io/ingress-nginx/images/fastcgi/fastcgi-helloserver /
USER nonroot:nonroot

CMD ["/fastcgi-helloserver"]
