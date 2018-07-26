default:

lint:
	docker run --rm -ti -v $(CURDIR):$(CURDIR) -w $(CURDIR) luzifer/eslint src/*.js

pack: webpack
	go-bindata -modtime 1 frontend/...

webpack: src/node_modules
	cd src && npm run build

src/node_modules:
	cd src && npm install

auto-hook-pre-commit: pack
	git diff --exit-code bindata.go

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
