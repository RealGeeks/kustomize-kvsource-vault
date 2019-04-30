FROM golang:1.12-stretch

ENV KUBECTL_VER 1.13.3

RUN apt-get update && apt-get install -y \
  curl \
  gettext \
  g++ \
  git 

RUN curl -L https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VER}/bin/linux/amd64/kubectl -o /usr/bin/kubectl \
  && chmod +x /usr/bin/kubectl

RUN go get github.com/hashicorp/vault/api
RUN go get sigs.k8s.io/kustomize
RUN go install sigs.k8s.io/kustomize

COPY ./vault.go /go/src/kustomize-kvsource-vault/

RUN go build -buildmode plugin -o /opt/kustomize/plugin/kvSources/vault.so /go/src/kustomize-kvsource-vault/vault.go 

FROM debian:stretch-slim

RUN apt-get update && apt-get install -y \
  git

COPY --from=0 /opt/kustomize/plugin/kvSources/vault.so /opt/kustomize/plugin/kvSources/vault.so
COPY --from=0 /go/bin/kustomize /usr/bin/kustomize
COPY --from=0 /usr/bin/kubectl /usr/bin/kubectl

WORKDIR /working 

ENV XDG_CONFIG_HOME=/opt

ENTRYPOINT ["/usr/bin/kustomize", "--enable_alpha_goplugins_accept_panic_risk"]
