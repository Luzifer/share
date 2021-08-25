default: webpack

lint:
	docker run --rm -ti \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npx eslint --ext .js,.vue --fix .

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

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
