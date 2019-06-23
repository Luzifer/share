<template>
  <div>

    <b-navbar variant="primary" type="dark">
      <b-navbar-brand href="#"><i class="fas fa-share-alt-square"></i> Share</b-navbar-brand>
    </b-navbar>

    <b-container class="mt-4">
      <b-row>
        <b-col>

          <b-card v-if="loading">
            <b-card-text class="text-center">
              <h2><i class="fas fa-spinner fa-pulse"></i></h2>
              {{ strings.loading }}
            </b-card-text>
          </b-card>

          <template v-else>

            <b-card v-if="error" bg-variant="danger" text-variant="white">
              <b-card-text class="text-center">
                <h2><i class="fas fa-exclamation-circle"></i></h2>
                {{ error }}
              </b-card-text>
            </b-card>

            <b-card v-else-if="fileType.startsWith('image/')">
              <b-card-text class="text-center">
                <a :href="path">
                  <b-img :src="path" fluid></b-img>
                </a>
              </b-card-text>
            </b-card>

            <b-card v-else-if="fileType.startsWith('video/')">
              <b-embed type="video" :src="path" allowfullscreen controls></b-embed>
            </b-card>

            <b-card v-else-if="fileType.startsWith('audio/')">
              <b-card-text class="text-center">
                <audio :src="path" controls></audio>
              </b-card-text>
            </b-card>

            <b-card v-else-if="fileType.startsWith('text/')">
              <pre><code>{{ text }}</code></pre>
            </b-card>

            <b-card v-else>
              <b-card-text class="text-center">
                <h2><i class="fas fa-cloud-download-alt"></i></h2>
                <b-button :href="path" variant="success">{{ fileName }}</b-button>
              </b-card-text>
            </b-card>

          </template>

        </b-col>
      </b-row>
    </b-container>

  </div>
</template>

<script>
import hljs from 'highlight.js'
import rewrites from './mime-rewrite.js'
import strings from './strings.js'

export default {
  name: 'app',

  computed: {
    strings() {
      return strings
    },
  },

  data() {
    return {
      error: null,
      fileType: null,
      fileName: '',
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
        this.error = strings.file_not_found
        this.loading = false
      }
    },
  },

  mounted() {
    window.onhashchange = this.hashChange
    this.hashChange()
  },

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

            if (this.text.length < 200*1024 && v !== 'text/plain') {
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
        this.error = strings.not_permitted
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
          this.fileName = this.path.substring(this.path.lastIndexOf('/')+1)
        })
        .catch(err => {
          switch (err.response.status) {
            case 403:
              this.error = strings.not_permitted
              break
            case 404:
              this.error = strings.file_not_found
              break
            default:
              this.error = `Something went wrong (Status ${err.response.status})`
          }
          this.loading = false
        })
    },
  },
}
</script>

<style scoped>
audio {
  width: 80%;
}
</style>
