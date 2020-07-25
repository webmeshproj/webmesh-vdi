<template>
  <q-page flex>
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

import WSAudioPlayer from '../lib/wsaudio.js'

function getWebsockifyAddr (namespace, name, token) {
  return `${window.location.origin.replace('http', 'ws')}/api/desktops/${namespace}/${name}/websockify?token=${token}`
}

function getWebsockifyAudioAddr (namespace, name, token) {
  return `${window.location.origin.replace('http', 'ws')}/api/desktops/${namespace}/${name}/wsaudio?token=${token}`
}

export default {
  name: 'VNCViewer',

  data () {
    return {
      rfb: null,
      player: null,
      currentSession: null,
      connected: false,
      statusLines: [],
      className: 'info',
      audioEnabled: false
    }
  },

  created () {
    this.unsubscribeSessions = this.$desktopSessions.subscribe(this.handleSessionsChange)
    this.$root.$on('set-fullscreen', this.setFullscreen)
  },

  beforeDestroy () {
    this.unsubscribeSessions()
    this.$root.$off('set-fullscreen', this.setFullscreen)
    this.disconnect()
  },

  methods: {

    enableAudio () {
      const audioUrl = getWebsockifyAudioAddr(this.currentSession.namespace, this.currentSession.name, this.$userStore.getters.token)
      console.log(`Connecting to audio stream at ${audioUrl}`)
      const playerCfg = { server: { url: audioUrl } }
      this.player = new WSAudioPlayer(playerCfg)
      this.player.start()
    },

    disableAudio () {
      if (this.player !== null) {
        console.log('Stopping audio stream')
        this.player.stop()
        this.player = null
      }
    },

    handleSessionsChange (mutation, state) {
      if (mutation.type === 'set_active_session') {
        const activeSession = this.$desktopSessions.getters.activeSession
        console.log(`Received a session change to ${JSON.stringify(activeSession)}`)
        if (activeSession === undefined) {
          console.log('There are no more active sessions, disconnecting')
          this.currentSession = null
          this.disconnect()
        } else {
          if (this.currentSession === activeSession) {
            console.log(`${activeSession.namespace}/${activeSession.name} is already the active session`)
            return
          }
          console.log(`Disconnecting from ${this.currentSession.name} and connecting to ${activeSession.name}`)
          this.disconnect().then(() => {
            this.currentSession = activeSession
            this.checkStatusAndConnect()
          })
        }
      }
      if (mutation.type === 'toggle_audio' && this.connected) {
        if (this.$desktopSessions.getters.audioEnabled) {
          this.enableAudio()
        } else {
          this.disableAudio()
        }
      }
    },

    setFullscreen (val) {
      if (val) {
        this.className = 'no-margin full-screen'
      } else if (this.connected) {
        this.className = 'no-margin to-header-height'
      } else {
        this.className = 'info'
      }
    },

    async checkStatusAndConnect () {
      try {
        const doConnect = await this.checkStatusLoop()
        if (doConnect) {
          this.createConnection()
        }
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async checkStatusLoop () {
      let podPhase
      let running
      let loopCount = 0
      const currentSession = this.currentSession
      while (this.sessionIsActiveSession(currentSession) && this.$router.currentRoute.name === 'control') {
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
        } else if (status.podPhase === 'Pending' && loopCount === 20) {
          this.statusLines.push('This is taking a while...the server might be pulling the image for the first time')
        } else if (status.podPhase === 'Running' && status.running !== running) {
          running = status.running
          if (!running) {
            this.statusLines.push('Waiting for desktop to finish booting...')
          } else {
            this.statusLines.push('Desktop has finished booting')
          }
        }
        if (this.statusIsReady(status)) {
          this.statusLines.push('Your desktop is ready')
          break
        }
        loopCount++
        await new Promise((resolve, reject) => setTimeout(resolve, 2000))
      }

      // Extra check to see if we were cancelled eaerly
      if (!this.sessionIsActiveSession(currentSession || this.$router.currentRoute.name !== 'control')) {
        return false
      }

      return true
    },

    sessionIsActiveSession (statusSession) {
      return this.$desktopSessions.getters.activeSession !== undefined && this.$desktopSessions.getters.activeSession === statusSession
    },

    statusIsReady (status) {
      return status.podPhase === 'Running' && status.running
    },

    createConnection () {
      let rfb

      try {
        const url = getWebsockifyAddr(this.currentSession.namespace, this.currentSession.name, this.$userStore.getters.token)
        const target = document.getElementById('view')
        if (target === null || target === undefined) {
          return
        }
        rfb = new RFB(target, url)
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

    async connectedToServer () {
      this.rfb.scaleViewport = true
      this.rfb.resizeSession = true
    },

    disconnectedFromServer (e) {
      if (e.detail.clean) {
        console.log('Disconnected')
      } else {
        this.resetStatus()
        console.log('Something went wrong, connection is closed')
        this.checkStatusAndConnect()
      }
      if (this.player !== null) {
        this.player.stop()
        this.player = null
        this.$desktopSessions.dispatch('toggleAudio', false)
      }
    },

    resetStatus () {
      this.connecting = false
      this.connected = false
      this.statusLines = []
      this.className = 'info'
    },

    async disconnect () {
      this.resetStatus()
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
      this.checkStatusAndConnect()
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
  top: 25%;
  left: 40%;
  margin: 0 auto;
  text-align: center;
  font-size: 16px;
}
</style>
