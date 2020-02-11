FROM golang:1.13-stretch

RUN apt-get update && apt-get install -y \
  curl \
  gettext \
  g++ \
  git 

WORKDIR /code

RUN GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4

COPY go.mod go.sum ./
RUN go mod download

COPY ./SecretsFromVault.go ./

RUN go build -buildmode plugin -o /opt/kustomize/plugin/kvSources/SecretsFromVault.so ./SecretsFromVault.go 

FROM debian:stretch-slim

RUN apt-get update && apt-get install -y \
  git

COPY --from=0 /opt/kustomize/plugin/kvSources/SecretsFromVault.so /opt/kustomize/plugin/kustomize.config.realgeeks.com/v1beta1/secretsfromvault/SecretsFromVault.so
COPY --from=0 /go/bin/kustomize /usr/bin/kustomize

WORKDIR /working 

ENV XDG_CONFIG_HOME=/opt

ENTRYPOINT ["/usr/bin/kustomize", "build", "--enable_alpha_plugins"]
