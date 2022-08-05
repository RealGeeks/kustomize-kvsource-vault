build: clean compile kustomize
	gzip --best SecretsFromVault.so
	gzip --best kustomize
	sha256sum SecretsFromVault.so.gz kustomize.gz
	mv SecretsFromVault.so.gz SecretsFromVault.v4.5.6.so.amd64.gz
	mv kustomize.gz kustomize.v4.5.6.amd64.gz

compile: kustomize
	GOARCH=amd64 GOOS=linux go build -buildmode plugin -o SecretsFromVault.so ./SecretsFromVault.go

gopath:
	mkdir -p /tmp/kustomize-kvsource-vault/go/
	export GOPATH=/tmp/kustomize-kvsource-vault/go/

install: compile
	mkdir -p ~/.config/kustomize/plugin/mycujoo.tv/v1/secretsfromvault/
	cp ./SecretsFromVault.so ~/.config/kustomize/plugin/mycujoo.tv/v1/secretsfromvault/SecretsFromVault.so

tempdir:
	mkdir /tmp/kustomize-kvsource-vault/ -p

kustomize:
	GOARCH=amd64 GOOS=linux go install sigs.k8s.io/kustomize/kustomize/v4@v4.5.6
	cp ${GOPATH}/bin/kustomize ./

clean-kustomize:
	rm -Rf /tmp/kustomize-kvsource-vault/kustomize
clean:
	go clean
	rm -f *.gz
