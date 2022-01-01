.PHONY: build clean deploy proto proto-clean elm

ELM_DIR=src

PUBLISH_DIR=public

elm:
	elm make $(ELM_DIR)/Main.elm --output=static/main.js

elm-release:
	rm -rf $(PUBLISH_DIR)
	cp -r static $(PUBLISH_DIR)
	elm make $(ELM_DIR)/Main.elm --output=static/main.opt.js --optimize
	uglifyjs --compress --mangle -- static/main.opt.js > public/main.js
	rm -f static/main.opt.js $(PUBLISH_DIR)/config.json
