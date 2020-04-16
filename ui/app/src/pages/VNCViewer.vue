<!-- WIP: I'd like to use a custom viewer component instead of the embedded noVNC html -->

<template>
  <q-page>
    <div id="view" :class="className"></div>
  </q-page>
</template>

<script>
import RFB from '@novnc/novnc/core/rfb'
import { init_logging as initLogging } from '@novnc/novnc/core/util/logging.js'

initLogging('debug')

let currentEndpoint

function getWebsockifyAddr (endpoint) {
  return `${window.location.origin.replace('https', 'wss')}/websockify/${endpoint}`
}

export function setEndpoint (ep) {
  currentEndpoint = ep
}

export default {
  name: 'VNCViewer',

  data () {
    return {
      rfb: null,
      desktopName: '',
      className: 'no-margin to-header-height'
    }
  },

  created () {
    this.$root.$on('set-fullscreen', this.setFullscreen)
  },

  beforeDestroy () {
    this.$root.$off('set-fullscreen', this.setFullscreen)
  },

  methods: {
    setFullscreen (val) {
      if (val) {
        this.className = 'no-margin full-screen'
      } else {
        this.className = 'no-margin to-header-height'
      }
    },
    createConnection () {
      let rfb
      try {
        const url = getWebsockifyAddr(currentEndpoint)
        rfb = new RFB(document.getElementById('view'), url)
        rfb.addEventListener('connect', this.connectedToServer)
        rfb.addEventListener('desktopname', this.updateDesktopName)
        rfb.addEventListener('disconnect', this.disconnectedFromServer)
        rfb.resizeSession = true
      } catch (err) {
        console.error(`Unable to create RFB client: ${err}`)
        this.disconnectedFromServer({ detail: { clean: false } })
        return
      }
      this.rfb = rfb
    },
    updateDesktopName (e) {
      this.desktopName = e.detail.name
    },
    connectedToServer () {
      this.rfb.scaleViewport = true
      this.rfb.resizeSession = true
      this.rfb._requestRemoteResize()
    },
    disconnectedFromServer (e) {
      if (e.detail.clean) {
        console.log('Disconnected')
      } else {
        console.log('Something went wrong, connection is closed')
      }
    },
    disconnect () {
      this.rfb.disconnect()
    }
  },
  mounted () {
    this.$nextTick(() => {
      if (currentEndpoint !== undefined) {
        this.createConnection()
      }
    })
  }
}
</script>

<style scoped>
.to-header-height {
  height: calc(100vh - 50px);
}
.full-screen {
  height: 100vh
}
</style>
