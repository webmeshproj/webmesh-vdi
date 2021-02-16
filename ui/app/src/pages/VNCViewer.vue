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
  <q-page flex>
    <div id="view-area">
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
      </div>
    </div>
  </q-page>
</template>

<script>
import DisplayManager from 'src/lib/displayManager.js'
import { Events } from 'src/lib/events.js'

export default {
  name: 'VNCViewer',

  data () {
    return {
      status: 'disconnected',
      statusLines: [],
      className: 'info',
      statusText: '',
      currentSession: null,
      displayManager: null
    }
  },

  created () {
    this.displayManager = new DisplayManager({
      userStore: this.$userStore,
      sessionStore: this.$desktopSessions
    })
    this.displayManager.on(Events.connected, this.onConnect)
    this.displayManager.on(Events.disconnected, this.onDisconnect)
    this.displayManager.on(Events.update, this.onStatusUpdate)
    this.displayManager.on(Events.error, this.onError)
    this.$root.$on('set-fullscreen', this.setFullscreen)
    this.$root.$on('paste-clipboard', this.onPaste)
    this.setCurrentSession()
  },

  beforeDestroy () {
    this.$root.$off('set-fullscreen', this.setFullscreen)
    this.$root.$off('paste-clipboard', this.onPaste)
    this.displayManager.destroy()
  },

  methods: {

    onPaste (data) { this.displayManager.sendClipboardData(data) },

    setCurrentSession () { this.currentSession = this.displayManager.getCurrentSession() },

    setFullscreen (val) {
      if (val) {
        this.className = 'no-margin full-screen'
      } else if (this.status === 'connected') {
        this.className = 'no-margin display-container'
      } else {
        this.className = 'info'
      }
    },

    onConnect () {
      this.setCurrentSession()
      this.status = 'connected'
      this.className = 'no-margin display-container'
      this.statusText = ''
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
