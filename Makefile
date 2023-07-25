.PHONY: protoc-install
protoc-install:
	brew install protobuf
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

.PHONY: generate
generate:
	protoc -I=proto --go_out=pkg/streamspb --go_opt=paths=source_relative \
					--go-grpc_out=pkg/streamspb --go-grpc_opt=paths=source_relative \
					proto/streams.proto

.gitignore:
	@wget https://www.toptal.com/developers/gitignore/api/go,goland+all,visualstudiocode -q -O .gitignore

.PHONY: skaffold-install
skaffold-install:
	brew install skaffold

