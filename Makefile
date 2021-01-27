default:

lint:
	docker run --rm -ti \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npx eslint --ext .js,.vue --fix .

pack: webpack
	go-bindata \
		-modtime 1 \
		frontend/...

webpack: src/node_modules
	docker run --rm -i \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npm run build

src/node_modules:
	docker run --rm -i \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npm ci

auto-hook-pre-commit: pack
	git diff --exit-code bindata.go

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
