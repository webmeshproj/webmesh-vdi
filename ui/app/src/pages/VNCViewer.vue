<template>
  <q-page flex>
    <div contenteditable="true" id="view" :class="className">
      <div q-gutter-md row v-if="status === 'disconnected' && currentSession">
        <q-spinner-hourglass color="grey" size="4em" />
        <q-space />
        <pre>{{ statusText }}</pre>
      </div>
      <div q-gutter-md row items-center v-if="status === 'disconnected' && !currentSession">
        <q-icon name="warning" class="text-red" style="font-size: 4rem;" />
        <br />
        There are no active desktop sessions
      </div>
      <iframe :class="xpraClassname" v-if="status === 'connected' && currentSession.socketType === 'xpra'" :src="`https://xpra.org/html5/index.html?${xpraArgs}`" />
    </div>
  </q-page>
</template>

<script>
import DisplayManager from '../lib/displayManager.js'

export default {
  name: 'VNCViewer',

  data () {
    return {
      status: 'disconnected',
      statusLines: [],
      className: 'info',
      xpraClassname: 'iframe-container',
      statusText: '',
      currentSession: null
    }
  },

  created () {
    this.displayManager = new DisplayManager(this.displayManagerArgs)
    this.$root.$on('set-fullscreen', this.setFullscreen)
    this.$root.$on('paste-clipboard', this.onPaste)
  },

  beforeDestroy () {
    this.$root.$off('set-fullscreen', this.setFullscreen)
    this.$root.$off('paste-clipboard', this.onPaste)
    this.displayManager.destroy()
  },

  computed: {
    displayManagerArgs () {
      return {
        userStore: this.$userStore,
        sessionStore: this.$desktopSessions,
        onError: this.onError,
        onStatusUpdate: this.onStatusUpdate,
        onDisconnect: this.onDisconnect,
        onConnect: this.onConnect
      }
    },

    xpraArgs () {
      if (this.status !== 'connected') {
        return ''
      }
      const args = this.displayManager.xpraArgs()
      return `\
      server=${args.host}\
      &port=${args.port}\
      &path=${args.path}\
      &ssl=${args.secure}\
      &reconnect=true\
      &dpi=96\
      &notifications=false\
      &clipboard=true\
      &compression_level=1\
      &floating_menu=false\
      `.replace(/\s/g, '')
    }
  },

  methods: {

    onPaste (data) { this.displayManager.syncClipboardData(data) },

    setCurrentSession () { this.currentSession = this.displayManager.getCurrentSession() },

    setFullscreen (val) {
      if (val) {
        this.className = 'no-margin full-screen'
        this.xpraClassname = 'no-margin full-screen iframe-full-screen'
      } else if (this.status === 'connected') {
        this.className = 'no-margin display-container'
        this.xpraClassname = 'iframe-container'
      } else {
        this.className = 'info'
        this.xpraClassname = 'iframe-container'
      }
    },

    onConnect () {
      this.setCurrentSession()
      this.status = 'connected'
      this.className = 'no-margin display-container'
    },

    onDisconnect () {
      this.setCurrentSession()
      this.status = 'disconnected'
      this.className = 'info'
    },

    onStatusUpdate (st) {
      this.setCurrentSession()
      this.statusText = st
    },

    onError (err) {
      this.setCurrentSession()
      this.$root.$emit('notify-error', err)
    }

  },

  mounted () {
    this.$nextTick(() => { this.displayManager.connect() })
  }
}
</script>

<style scoped>
.display-container {
  display: flex;
  width: 100%;
  height: calc(100vh - 100px);
  flex-direction: column;
  background-color: blue;
  overflow: hidden;
}

.iframe-container {
  flex-grow: 1;
  border: none;
  margin: 0;
  padding: 0;
}

.full-screen {
  height: 100vh;
}

.iframe-full-screen {
  position:fixed;
  top:0;
  left:0;
  bottom:0;
  right:0;
  width:100%;
  height:100%;
  border:none;
  margin:0;
  padding:0;
  overflow:hidden;
  z-index:999999;
}

.info {
  position: absolute;
  top: 25%;
  left: 40%;
  margin: 0 auto;
  text-align: center;
  font-size: 16px;
}
</style>
