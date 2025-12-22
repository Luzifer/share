default:

lint:
	docker run --rm -ti \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npx eslint --ext .js --fix frontend/app.js

.PHONY: frontend
frontend: node_modules
	node ci/build.mjs

node_modules:
	npm ci --include=dev

publish: frontend
	bash ci/build.sh
