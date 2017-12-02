default:

pack: generate
	go-bindata -modtime 1 frontend

generate:
	coffee -c frontend/app.coffee

auto-hook-pre-commit: pack
	git diff --exit-code bindata.go

publish:
	curl -sSLo golang.sh https://raw.githubusercontent.com/Luzifer/github-publish/master/golang.sh
	bash golang.sh
