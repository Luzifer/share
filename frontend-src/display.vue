<template>
  <div>
    <nav class="navbar navbar-expand-lg bg-body-tertiary">
      <div class="container-fluid">
        <a
          class="navbar-brand"
          href="#"
          @click.prevent=""
        >
          <i class="fas fa-fw fa-share-alt-square mr-1" /> Share
        </a>
      </div>
    </nav>

    <div class="container my-4">
      <div class="row">
        <div class="col">
          <div
            v-if="loading"
            class="card"
          >
            <div class="card-body text-center">
              <h2><i class="fas fa-spinner fa-pulse" /></h2>
              {{ $t('loading') }}
            </div>
          </div>

          <template v-else>
            <div
              v-if="error"
              class="card text-bg-danger"
            >
              <div class="card-body text-center">
                <h2><i class="fas fa-exclamation-circle" /></h2>
                {{ error }}
              </div>
            </div>

            <div
              v-else-if="fileType.startsWith('application/pdf') && canViewPDF"
              class="card position-relative pdf-frame"
            >
              <div class="card-body text-center">
                <iframe
                  class="h-100 position-absolute start-0 top-0 w-100"
                  :src="`${path}#page=1&view=fit`"
                />
              </div>
            </div>

            <div
              v-else-if="fileType.startsWith('image/')"
              class="card"
            >
              <div class="card-body text-center">
                <a :href="path">
                  <img
                    :src="path"
                    class="img-fluid"
                  >
                </a>
              </div>
            </div>

            <div
              v-else-if="fileType.startsWith('video/')"
              class="card"
            >
              <div class="card-body text-center">
                <div class="ratio ratio-16x9">
                  <video controls>
                    <source :src="path">
                  </video>
                </div>
              </div>
            </div>

            <div
              v-else-if="fileType.startsWith('audio/')"
              class="card"
            >
              <div class="card-body text-center">
                <audio
                  controls
                  :src="path"
                />
              </div>
            </div>

            <div
              v-else-if="fileType.startsWith('text/markdown')"
              class="card"
            >
              <div
                class="card-body"
                v-html="renderMarkdown(text)"
              />
            </div>

            <div
              v-else-if="fileType.startsWith('text/')"
              class="card"
            >
              <div class="card-body">
                <pre><code>{{ text }}</code></pre>
              </div>
            </div>

            <div
              v-else
              class="card"
            >
              <div class="card-body text-center">
                <h2><i class="fas fa-cloud-download-alt" /></h2>
                <a
                  class="btn btn-success"
                  :href="path"
                >
                  {{ fileName }}
                </a>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import hljs from 'highlight.js'
import showdown from 'showdown'

const rewrites = {
  'application/javascript': 'text/javascript',
}

const markdownClassMap = {
  table: 'table table-sm',
}

const markdownBindings = Object.keys(markdownClassMap)
  .map(key => ({
    regex: new RegExp(`<${key}(.*)>`, 'g'),
    replace: `<${key} class="${markdownClassMap[key]}" $1>`,
    type: 'output',
  }))

export default defineComponent({
  computed: {
    canViewPDF(): boolean {
      /*
       * iOS reports they can display PDFs but they'll display only
       * the first page without any means of navigation, therefore we
       * disable PDF viewing on iOS
       */
      const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) || navigator.platform === 'MacIntel' && navigator.maxTouchPoints > 1

      return navigator.pdfViewerEnabled && !isIOS
    },
  },

  data() {
    return {
      error: null,
      fileName: '',
      fileType: '',
      loading: true,
      path: '',
      text: '',
    }
  },

  methods: {
    hashChange() {
      const hash = window.location.hash

      if (hash.length > 0) {
        this.path = hash.substring(1)
      } else {
        this.error = this.$i18n.t('fileNotFound')
        this.loading = false
      }
    },

    renderMarkdown(text) {
      const sd = new showdown.Converter({
        backslashEscapesHTMLTags: true,
        disableForced4SpacesIndentedSublists: true,
        emoji: true,
        extensions: [...markdownBindings],
        ghCodeBlocks: true,
        ghCompatibleHeaderId: true,
        ghMentions: false,
        literalMidWordUnderscores: true,
        omitExtraWLInCodeBlocks: true,
        requireSpaceBeforeHeadingText: true,
        simpleLineBreaks: true,
        simplifiedAutoLink: true,
        splitAdjacentBlockquotes: true,
        strikethrough: true,
        tables: true,
        tablesHeaderId: true,
        tasklists: true,
      })

      return sd.makeHtml(text)
    },
  },

  mounted() {
    window.onhashchange = this.hashChange
    this.hashChange()
  },

  name: 'ShareContentDisplay',

  watch: {
    fileType(v) {
      // Rewrite known file types not matching the expectations above
      if (rewrites[v]) {
        this.fileType = rewrites[v]
        return
      }

      // Load text files directly and highlight them
      if (v.startsWith('text/')) {
        this.loading = true
        fetch(this.path)
          .then(resp => resp.text())
          .then(text => {
            this.text = text

            if (this.text.length < 200 * 1024 && v !== 'text/plain') {
              // Only highlight up to 200k and not on text/plain
              window.setTimeout(() => hljs.initHighlighting(), 100)
            }
            this.loading = false
          })
          .catch(err => console.log(err))
      }
    },

    path() {
      if (this.path.indexOf('://') >= 0) {
        // Strictly disallow loading files having any protocol in them
        this.error = this.$i18n.t('notPermitted')
        this.loading = false
        return
      }

      fetch(this.path, {
        method: 'HEAD',
      })
        .then(resp => {
          this.loading = false

          switch (resp.status) {
          case 200:
            break
          case 403:
            this.error = this.$i18n.t('notPermitted')
            return
          case 404:
            this.error = this.$i18n.t('fileNotFound')
            return
          default:
            this.error = this.$i18n.t('genericError', { status: resp.status })
            return
          }

          this.fileType = resp?.headers?.get('content-type') || 'application/octet-stream'
          this.fileName = this.path.substring(this.path.lastIndexOf('/') + 1)
        })
    },
  },
})
</script>

<style scoped>
.pdf-frame {
  height: 0;
  padding-top: calc(100vh - 70px - 3rem);
}
</style>
