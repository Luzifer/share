/* eslint-disable sort-imports */

import 'bootstrap/dist/css/bootstrap.css'
import '@fortawesome/fontawesome-free/css/all.css'

import { createApp, h } from 'vue'
import { createI18n } from 'vue-i18n'

import ContentDisplay from './display.vue'

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

const app = createApp({
  name: 'Share',
  render() {
    return h(ContentDisplay)
  },
})

app.use(createI18n({
  fallbackLocale: 'en',
  locale: new URLSearchParams(window.location.search).get('hl') || navigator.languages?.[0].split('-')[0] || navigator.language?.split('-')[0] || 'en',
  messages,
}))

app.mount('#app')
