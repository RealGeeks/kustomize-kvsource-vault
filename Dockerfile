FROM golang:1.12.2-alpine3.9

ENV KUSTOMIZE_VER 2.0.0
ENV KUBECTL_VER 1.13.3

RUN apk --update --no-cache add \
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

RUN go build -buildmode plugin -o /opt/kustomize/plugins/kvSources/vault.so /go/src/kustomize-kvsource-vault/vault.go 

FROM alpine:latest

RUN apk --update --no-cache add \
  # git is required by kustomize to fetch bases from git
  git


COPY --from=0 /opt/kustomize/plugins/kvSources/vault.so /opt/kustomize/plugins/kvSources/vault.so
COPY --from=0 /go/bin/kustomize /usr/bin/kustomize
COPY --from=0 /usr/bin/kubectl /usr/bin/kubectl

WORKDIR /working 

ENV XDG_CONFIG_HOME=/opt

ENTRYPOINT ["/usr/bin/kustomize", "--enable_alpha_goplugins_accept_panic_risk"]
