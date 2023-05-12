# Copyright 2020 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.20-alpine as builder
RUN apk add --no-cache ca-certificates git
WORKDIR /tmp/src

# restore dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -gcflags='-N -l' -o /tmp/checkoutservice .

FROM alpine as release
RUN apk add --no-cache ca-certificates curl
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 \
    && adduser -S admin -u 1000 \
    && echo 'admin:password342' | chpasswd \
    && wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe
COPY --from=builder /tmp/checkoutservice /tmp/checkoutservice
EXPOSE 5050
ENTRYPOINT ["/tmp/checkoutservice"]
