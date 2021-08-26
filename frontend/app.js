/* global axios, hljs, showdown, Vue */

const rewrites = {
  'application/javascript': 'text/javascript',
}

const messages = {
  en: {
    fileNotFound: 'The requested file has not been found.',
    genericError: 'Something went wrong (Status {status})',
    loading: 'Loading file details...',
    notPermitted: 'Access to this file was denied.',
  },
  de: {
    fileNotFound: 'Die angegebene Datei wurde nicht gefunden.',
    genericError: 'Irgendwas lief schief... (Status {status})',
    loading: 'Lade Datei-Informationen...',
    notPermitted: 'Der Zugriff auf diese Datei wurde verweigert.',
  },
}

new Vue({
  data: {
    error: null,
    fileName: '',
    fileType: null,
    loading: true,
    path: '',
    text: '',
  },

  el: '#app',

  i18n: new VueI18n({
    fallbackLocale: 'en',
    locale: new URLSearchParams(window.location.search).get('hl') || navigator.languages?.[0].split('-')[0] || navigator.language?.split('-')[0] || 'en',
    messages,
  }),

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
      return new showdown.Converter().makeHtml(text)
    },
  },

  mounted() {
    window.onhashchange = this.hashChange
    this.hashChange()
  },

  name: 'App',

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
        axios.get(this.path)
          .then(resp => {
            this.text = resp.data

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

      axios.head(this.path)
        .then(resp => {
          let contentType = 'application/octet-stream'
          if (resp && resp.headers && resp.headers['content-type']) {
            contentType = resp.headers['content-type']
          }

          this.loading = false
          this.fileType = contentType
          this.fileName = this.path.substring(this.path.lastIndexOf('/') + 1)
        })
        .catch(err => {
          switch (err.response.status) {
          case 403:
            this.error = this.$i18n.t('notPermitted')
            break
          case 404:
            this.error = this.$i18n.t('fileNotFound')
            break
          default:
            this.error = this.$i18n.t('genericError', { status: err.response.status })
          }
          this.loading = false
        })
    },
  },
})
