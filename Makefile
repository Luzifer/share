default:

lint:
	docker run --rm -ti \
		-v "$(CURDIR):/src" \
		-w "/src/src" \
		node:12-alpine \
		npx eslint --ext .js --fix frontend/app.js

assets: frontend/bundle.css
assets: frontend/bundle.js

frontend/bundle.css:
	./ci/combine.sh $@ \
		npm/bootstrap@4/dist/css/bootstrap.min.css \
		npm/bootstrap-vue@2/dist/bootstrap-vue.min.css \
		npm/bootswatch@5/dist/darkly/bootstrap.min.css \
		gh/highlightjs/cdn-release@11.2.0/build/styles/androidstudio.min.css

frontend/bundle.js:
	./ci/combine.sh $@ \
		npm/axios@0.21.1 \
		npm/vue@2 \
		npm/vue-i18n@8.25.0/dist/vue-i18n.min.js \
		npm/bootstrap-vue@2/dist/bootstrap-vue.min.js \
		npm/showdown@1 \
		gh/highlightjs/cdn-release@11.2.0/build/highlight.min.js

publish:
	bash ci/build.sh
