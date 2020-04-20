<!-- WIP: I'd like to use a custom viewer component instead of the embedded noVNC html -->

<template>
  <q-page>
    <div id="view" :class="className">
      <div q-gutter-md row v-if="!connected && currentSession !== null">
        <q-spinner-hourglass color="grey" size="4em" />
        <q-space />
        <div v-for="line in statusLines" :key="line">
          {{ line }}
        </div>
      </div>
      <div q-gutter-md row items-center v-if="!connected && currentSession === null">
        <q-icon name="warning" class="text-red" style="font-size: 4rem;" />
        <br />
        There are no active desktop sessions
      </div>
    </div>
  </q-page>
</template>

<script>
import RFB from '@novnc/novnc/core/rfb'
import { init_logging as initLogging } from '@novnc/novnc/core/util/logging.js'

initLogging('info')

function getWebsockifyAddr (endpoint, token) {
  return `${window.location.origin.replace('http', 'ws')}/api/websockify/${endpoint}?token=${token}`
}

export default {
  name: 'VNCViewer',

  data () {
    return {
      rfb: null,
      currentSession: null,
      connected: false,
      statusLines: [],
      className: 'info'
    }
  },

  created () {
    this.unsubscribeSessions = this.$desktopSessions.subscribe(this.handleSessionsChange)
    this.$root.$on('set-fullscreen', this.setFullscreen)
  },

  beforeDestroy () {
    this.unsubscribeSessions()
    this.$root.$off('set-fullscreen', this.setFullscreen)
  },

  methods: {
    handleSessionsChange (mutation, state) {
      const currentSession = this.$desktopSessions.getters.activeSession
      if (currentSession === undefined) {
        this.currentSession = null
        this.disconnect()
      } else {
        if (currentSession.endpoint !== this.currentSession.endpoint) {
          this.disconnect().then(() => {
            this.currentSession = currentSession
            this.checkStatusLoop()
              .then((cont) => {
                if (cont) {
                  this.createConnection()
                }
              })
              .catch((err) => {
                this.$root.$emit('notify-error', err)
              })
          })
        }
      }
    },
    setFullscreen (val) {
      if (val) {
        this.className = 'no-margin full-screen'
      } else {
        this.className = 'no-margin to-header-height'
      }
    },
    async checkStatusLoop () {
      let podPhase
      let running
      let resolvable
      let loopCount = 0
      const currentSession = this.currentSession
      while (this.$desktopSessions.getters.activeSession !== undefined && this.$desktopSessions.getters.activeSession === currentSession) {
        const status = await this.$desktopSessions.getters.sessionStatus(this.currentSession)
        console.log(status)
        if (this.statusIsReady(status) && loopCount === 0) {
          break
        }
        if (status.podPhase === '') {
          await new Promise((resolve, reject) => setTimeout(resolve, 2000))
          continue
        } else if (status.podPhase !== podPhase) {
          podPhase = status.podPhase
          if (status.podPhase === 'Pending' || status.podPhase === 'ContainerCreating') {
            this.statusLines.push('Waiting for container to start...')
          } else if (status.podPhase === 'Running') {
            this.statusLines.push('Container has started')
          }
        } else if (status.running !== running) {
          running = status.running
          if (!running) {
            this.statusLines.push('Waiting for desktop to finish booting...')
          } else {
            this.statusLines.push('Desktop has finished booting')
          }
        } else if (status.resolvable !== resolvable) {
          resolvable = status.resolvable
          if (!resolvable) {
            this.statusLines.push('Waiting for desktop to be reachable...')
          } else {
            this.statusLines.push('Desktop is reachable')
          }
        }
        if (this.statusIsReady(status)) {
          this.statusLines.push('Your desktop is ready')
          break
        }
        loopCount++
        await new Promise((resolve, reject) => setTimeout(resolve, 2000))
      }

      // Extra check to see if we were cancelled wrongly
      if (this.$desktopSessions.getters.activeSession === undefined || this.$desktopSessions.getters.activeSession !== currentSession) {
        return false
      }

      return true
    },
    statusIsReady (status) {
      return status.podPhase === 'Running' && status.running && status.resolvable
    },
    createConnection () {
      let rfb
      try {
        const url = getWebsockifyAddr(this.currentSession.endpoint, this.$userStore.getters.token)
        rfb = new RFB(document.getElementById('view'), url)
        rfb.addEventListener('connect', this.connectedToServer)
        rfb.addEventListener('disconnect', this.disconnectedFromServer)
        rfb.resizeSession = true
      } catch (err) {
        console.error(`Unable to create RFB client: ${err}`)
        this.disconnectedFromServer({ detail: { clean: false } })
        return
      }
      this.connected = true
      this.className = 'no-margin to-header-height'
      this.rfb = rfb
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
      this.disconnect()
    },
    async disconnect () {
      this.connecting = false
      this.connected = false
      this.statusLines = []
      this.className = 'info'
      if (this.rfb !== null) {
        this.rfb.disconnect()
        this.rfb = null
      }
    }
  },

  mounted () {
    this.$nextTick(() => {
      const currentSession = this.$desktopSessions.getters.activeSession
      if (currentSession === undefined) {
        return
      }
      this.currentSession = currentSession
      this.checkStatusLoop()
        .then((cont) => {
          if (cont) {
            this.createConnection()
          }
        })
        .catch((err) => {
          this.$root.$emit('notify-error', err)
        })
    })
  }
}
</script>

<style scoped>
.to-header-height {
  height: calc(100vh - 100px);
}
.full-screen {
  height: 100vh;
}
.info {
  position: absolute;
  top: 45%;
  left: 45%;
  margin: 0 auto;
  text-align: center;
  font-size: 16px;
}
</style>
