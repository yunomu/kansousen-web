.PHONY: build clean deploy proto proto-clean elm

SERVICES=auth kifu
PROTOBUF=document api
ELM_DIR=elm

TARGETS=$(addprefix bin/, $(SERVICES))
PROTO_TARGETS=$(addprefix proto/, $(PROTOBUF))
PROTO_ELM_TARGETS=$(ELM_DIR)/Proto/Api.proto


build: clean $(TARGETS)

bin/%: %/main.go
	env GOOS=linux go build -ldflags="-s -w" -o $@ $^

clean:
	rm -rf ./bin

deploy: build
	sls deploy --verbose

proto: $(PROTO_TARGETS) $(PROTO_ELM_TARGETS)

proto/%: proto/%.proto
	protoc --go_out=. $<

$(PROTO_ELM_TARGETS): proto/api.proto
	protoc --elm_out=$(ELM_DIR) $< 2> /dev/null

elm:
	elm make elm/Main.elm --output=static/main.js

elm-release:
	elm make elm/Main.elm --output=static/main.opt.js --optimize
	uglifyjs --compress --mangle -- static/main.opt.js > public/main.min.js
