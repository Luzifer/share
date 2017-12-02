default:

pack: generate
	go-bindata frontend

generate:
	coffee -c frontend/app.coffee
