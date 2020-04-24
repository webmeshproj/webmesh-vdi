<template>
  <div id="swagger-ui"/>
</template>

<script>
import SwaggerUIBundle from 'swagger-ui'
import 'swagger-ui/dist/swagger-ui.css'

export default {
  name: 'APIExplorer',
  data () {
    return {
      ui: null
    }
  },
  mounted () {
    this.$nextTick().then(() => {
      this.ui = SwaggerUIBundle({
        url: '/swagger.json',
        dom_id: '#swagger-ui',
        deepLinking: false,
        displayRequestDuration: true,
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
