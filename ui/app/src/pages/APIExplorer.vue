<!--
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
-->

<template>
  <div id="swagger-ui"/>
</template>

<script>
import SwaggerUIBundle from 'swagger-ui'

export default {
  name: 'APIExplorer',
  data () {
    return {
      ui: null
    }
  },
  created () {
    this.unsubscribeTokens = this.$userStore.subscribe(this.onTokenRefresh)
  },
  beforeDestroy () {
    this.unsubscribeTokens()
  },
  methods: {
    onTokenRefresh (mutation, state) {
      if (mutation.type === 'auth_success') {
        this.ui.authActions.authorize({
          api_key: {
            name: 'api_key',
            schema: {
              type: 'apiKey',
              in: 'header',
              name: 'X-Session-Token',
              description: ''
            },
            value: state.token
          }
        })
      }
    }
  },
  mounted () {
    this.$nextTick().then(() => {
      this.ui = SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui',
        deepLinking: false,
        displayRequestDuration: true,
        responseInterceptor: (response) => {
          if (response.status === 401) {
            return this.$userStore.dispatch('refreshToken')
              .then(() => {
                response.text = '{"internal": "KVDI: Your access token has been refreshed, please try again"}'
                response.ok = true
                return response
              })
              .catch((err) => {
                console.error(err)
                response.text = `{"internal": "KVDI: Failure while trying to refresh access token: ${err}"}`
                return response
              })
          }
          return response
        },
        presets: [
          SwaggerUIBundle.presets.apis
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ]
      })
      if (this.$userStore.getters.isLoggedIn) {
        this.ui.authActions.authorize({
          api_key: {
            name: 'api_key',
            schema: {
              type: 'apiKey',
              in: 'header',
              name: 'X-Session-Token',
              description: ''
            },
            value: this.$userStore.getters.token
          }
        })
      }
    })
  }
}
</script>

<style lang="sass">
.scheme-container
  background-color: $blue-grey-3
</style>
