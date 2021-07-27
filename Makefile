.PHONY: build clean deploy proto proto-clean elm

SERVICES=auth kifu
PROTOBUF=document api lambdakifu
ELM_DIR=elm

TARGETS=$(addprefix bin/, $(SERVICES))
PROTO_TARGETS=$(addprefix proto/, $(PROTOBUF))
PROTO_ELM_TARGETS=$(ELM_DIR)/Proto/Api.proto

PUBLISH_DIR=public

build: proto
	sam build

bin/%: %/main.go
	env GOOS=linux go build -ldflags="-s -w" -o $@ $^

clean:
	rm -rf ./bin

deploy: build
	sam deploy

proto: $(PROTO_TARGETS) $(PROTO_ELM_TARGETS)

proto/%: proto/%.proto
	protoc --go_out=. $<

$(PROTO_ELM_TARGETS): proto/api.proto
	protoc --elm_out=$(ELM_DIR) $< 2> /dev/null

elm:
	elm make $(ELM_DIR)/Main.elm --output=static/main.js

elm-release:
	rm -rf $(PUBLISH_DIR)
	cp -r static $(PUBLISH_DIR)
	elm make $(ELM_DIR)/Main.elm --output=static/main.opt.js --optimize
	uglifyjs --compress --mangle -- static/main.opt.js > public/main.js
	rm -f static/main.opt.js $(PUBLISH_DIR)/config.json
