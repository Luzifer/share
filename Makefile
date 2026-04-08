default:

lint:
	pnpm eslint frontend-src

.PHONY: frontend
frontend: node_modules
	pnpm node ci/build.mjs

node_modules:
	pnpm i --frozen-lockfile

publish: frontend
	bash ci/build.sh
