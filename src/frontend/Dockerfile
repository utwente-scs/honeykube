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

FROM golang:1.11.5-alpine as builder
RUN apk add --no-cache ca-certificates git build-base gcc abuild binutils binutils-doc gcc-doc

WORKDIR /tmp/src

# restore dependencies
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/frontend .

FROM alpine  as release
RUN apk add --no-cache ca-certificates curl \
    busybox-extras net-tools openrc bind-tools \
    && adduser -D admin -u 1000 \
    && echo 'admin:password1' | chpasswd &&\
    chown -R admin:admin /home/admin

WORKDIR /tmp/frontend
COPY --from=builder /go/bin/frontend /tmp/frontend/server
COPY ./templates ./templates
COPY ./static ./static

EXPOSE 8080
ENTRYPOINT ["/tmp/frontend/server"]