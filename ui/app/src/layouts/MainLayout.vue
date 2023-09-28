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
  <q-layout view="hHh lpR fFf">

    <q-header :value="revealHeader"  class="bg-grey-9 text-white" height-hint="98">
      <q-toolbar class="flat">
        <q-btn dense flat round icon="menu" @click="openDrawer = !openDrawer" />

        <q-toolbar-title>
          <q-avatar>
            <img src="../assets/logo.png">
          </q-avatar>
          kVDI
        </q-toolbar-title>

        <q-space />

        <q-btn type="a" href="https://discord.gg/vpkFjGuwYC" target="_blank" stretch flat icon-right="img:https://assets-global.website-files.com/6257adef93867e50d84d30e2/636e0a6ca814282eca7172c6_icon_clyde_white_RGB.svg" label="Get help at Discord" />
        <q-btn type="a" href="https://github.com/webmeshproj/webmesh-vdi" target="_blank" stretch flat label="Github"    >
          <svg style="margin-left: 12px" height="32" aria-hidden="true" viewBox="0 0 16 16" version="1.1" width="32" data-view-component="true" class="octicon octicon-mark-github v-align-middle color-fg-default">
            <path d="M8 0c4.42 0 8 3.58 8 8a8.013 8.013 0 0 1-5.45 7.59c-.4.08-.55-.17-.55-.38 0-.27.01-1.13.01-2.2 0-.75-.25-1.23-.54-1.48 1.78-.2 3.65-.88 3.65-3.95 0-.88-.31-1.59-.82-2.15.08-.2.36-1.02-.08-2.12 0 0-.67-.22-2.2.82-.64-.18-1.32-.27-2-.27-.68 0-1.36.09-2 .27-1.53-1.03-2.2-.82-2.2-.82-.44 1.1-.16 1.92-.08 2.12-.51.56-.82 1.28-.82 2.15 0 3.06 1.86 3.75 3.64 3.95-.23.2-.44.55-.51 1.07-.46.21-1.61.55-2.33-.66-.15-.24-.6-.83-1.23-.82-.67.01-.27.38.01.53.34.19.73.9.82 1.13.16.45.68 1.31 2.69.94 0 .67.01 1.3.01 1.49 0 .21-.15.45-.55.38A7.995 7.995 0 0 1 0 8c0-4.42 3.58-8 8-8Z"></path>
         </svg>
        </q-btn>
       
      </q-toolbar>


      <q-tabs align="center" v-if="controlSessions.length != 0">
        <SessionTab v-for="tab in controlSessions" v-bind="tab" :key="(tab as any).name" />
      </q-tabs>
    </q-header>

    <q-drawer  v-model="openDrawer" side="left" behavior="desktop">
      <q-list>
        <q-item-label
          header
          class="text-grey-8"
        >
          Menu
        </q-item-label>

        <!-- Desktop Templates  -->
        <q-item clickable tag="a" href="#/templates" :active="desktopTemplatesActive" @click="onClickDesktopTemplates">

          <q-item-section avatar>
            <q-icon name="devices" />
          </q-item-section>

          <q-item-section>
            <q-item-label>Desktop Templates</q-item-label>
            <q-item-label caption>Containerized workspace environments</q-item-label>
          </q-item-section>

        </q-item>

        <!-- Desktop session controls  -->
        <q-expansion-item
          v-model="controlActive"
          label="Control"
          caption="Interact with a desktop session"
          icon="cast"
          to="control"
          :active="controlActive"
          @click="onClickControl"
          group="control"
          :content-inset-level="0.2"
        >

          <q-list>
            <q-item dense clickable @click="() => { this.$q.fullscreen.request() }">

              <q-item-section avatar>
                <q-icon name="fullscreen" />
              </q-item-section>

              <q-item-section>
                <q-item-label caption>Enter fullscreen mode</q-item-label>
              </q-item-section>

            </q-item>

            <q-expansion-item
              dense dense-toggle
              v-model="audioExpanded"
              :caption="audioCaption"
              :icon="audioIcon"
              :active="audioEnabled"
              :header-class="audioHeaderClass"
              :content-inset-level="0.2"
              @click="onClickAudio"
            >

              <q-list v-if="audioExpanded">

                <q-item dense :active="recordingEnabled" clickable @click="onClickRecord">

                  <q-item-section avatar>
                    <q-icon :name="recordingIcon" />
                  </q-item-section>

                  <q-item-section>
                    <q-item-label caption>{{ recordingCaption }}</q-item-label>
                  </q-item-section>

                </q-item>

              </q-list>

            </q-expansion-item>

            <!-- <q-item dense :active="audioEnabled" clickable @click="onClickAudio">

              <q-item-section avatar>
                <q-icon :name="audioIcon" />
              </q-item-section>

              <q-item-section>
                <q-item-label caption>{{ audioCaption }}</q-item-label>
              </q-item-section>

            </q-item> -->

            <q-item dense clickable @click="onPaste">

              <q-item-section avatar>
                <q-icon name="content_copy" />
              </q-item-section>

              <q-item-section>
                <q-item-label caption>Sync clipboard to remote</q-item-label>
              </q-item-section>

            </q-item>

            <q-item dense clickable @click="onFileTransfer">

              <q-item-section avatar>
                <q-icon name="cloud_download" />
              </q-item-section>

              <q-item-section>
                <q-item-label caption>Transfer files to/from desktop</q-item-label>
              </q-item-section>

            </q-item>

          </q-list>

        </q-expansion-item>

        <!-- Settings  -->
        <q-item clickable tag="a" href="#/settings" :active="settingsActive" @click="onClickSettings">

          <q-item-section avatar>
            <q-icon name="settings" />
          </q-item-section>

          <q-item-section>
            <q-item-label>Settings</q-item-label>
            <q-item-label caption>Configure roles and users</q-item-label>
          </q-item-section>

        </q-item>

        <!-- API Explorer  -->
        <q-item clickable tag="a" href="#/swagger" :active="apiExplorerActive" @click="onClickAPIExplorer">

          <q-item-section avatar>
            <q-icon name="code" />
          </q-item-section>

          <q-item-section>
            <q-item-label>API Explorer</q-item-label>
            <q-item-label caption>Interact with the kVDI API</q-item-label>
          </q-item-section>

        </q-item>

        <!-- Metrics  -->
        <q-item v-if="grafanaEnabled" clickable tag="a" href="#/metrics" :active="metricsActive" @click="onClickMetrics">

          <q-item-section avatar>
            <q-icon name="calculate" />
          </q-item-section>

          <q-item-section>
            <q-item-label>Metrics</q-item-label>
            <q-item-label caption>Visualize kVDI Performance</q-item-label>
          </q-item-section>

        </q-item>

      </q-list>

      <q-separator />

       <!-- Link to login when user is logged out -->
      <q-item clickable tag="a" href="#/login" :active="loginActive" @click="onClickLogin" v-if="!isLoggedIn">
        <q-item-section avatar>
          <q-icon name="meeting_room" />
        </q-item-section>
        <q-item-section>
          <q-item-label>Login</q-item-label>
        </q-item-section>
      </q-item>

      <!-- User controls when logged in  -->
      <q-expansion-item group="menu" :content-inset-level="0.2" v-if="isLoggedIn">

        <template v-slot:header>

          <q-item-section avatar>
            <q-avatar color="teal" text-color="white">{{ userInitial }}</q-avatar>
          </q-item-section>
          <q-item-section>
            {{ user.name }}
          </q-item-section>

        </template>

        <q-list>
          <!-- User profile for editing password/mfa settings -->
          <q-item clickable tag="a" href="#/profile" :active="profileActive" @click="onClickProfile">
            <q-item-section avatar>
              <q-icon name="supervisor_account" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Profile</q-item-label>
            </q-item-section>
          </q-item>

          <!-- Logout current session -->
          <q-item clickable tag="a" @click="onClickLogout">
            <q-item-section avatar>
              <q-icon name="desktop_access_disabled" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Logout</q-item-label>
            </q-item-section>
          </q-item>

        </q-list>

      </q-expansion-item>

    </q-drawer>

    <q-page-container>
      <transition
        enter-active-class="animated fadeIn"
        leave-active-class="animated fadeOut"
        appear
        :duration="200"
      >
        <router-view />

      </transition>
    </q-page-container>

  </q-layout>
</template>

<script lang="ts">

import SessionTab from '../components/SessionTab.vue'
import MFADialog from '../components/dialogs/MFADialog.vue'
import FileTransferDialog from '../components/dialogs/FileTransfer.vue'
import { getErrorMessage } from '../lib/util.js'
import { defineComponent }from 'vue'
import { useConfigStore } from '../stores/config'
import { useUserStore } from '../stores/user'
import { useDesktopSessions } from '../stores/desktop'

var menuTimeout: any = null

export default  defineComponent({
  name: 'MainLayout',
  components: { SessionTab },
  setup() {
    const configStore = useConfigStore()
    const userStore = useUserStore()
    const desktopSessions = useDesktopSessions()

    // **only return the whole store** instead of destructuring
    return { configStore , userStore,desktopSessions,unsubscribeSessions: () => {},unsubscribeUsers: () => {} }
  },
  

  async created () {
    this.subscribeToBuses()
    document.onfullscreenchange = this.handleFullScreenChange
    this.unsubscribeSessions = this.desktopSessions.$subscribe(this.handleSessionsChange)
    // this.unsubscribeUsers = this.userStore.$subscribe(this.handleAuthChange)
    await this.userStore.initStore()
    try {
      if (this.userStore.requiresMFA) {
        await this.$q.dialog({
          component: MFADialog,
          parent: this
        }).onOk(() => {
          this.handleLoggedIn()
        }).onCancel(() => {
        }).onDismiss(() => {
        })
        return
      }
      if (this.userStore.isLoggedIn) {
        this.handleLoggedIn()
      } else {
        this.onClickLogin()
        this.pushIfNotCurrent('login')
      }
    } catch (err) {
      this.$root?.$emit('notify-error', err)
      this.onClickLogin()
      this.pushIfNotCurrent('login')
    }
    this.onFocusSyncRemoteClipboardListener()
  },

  mounted () {
    window.addEventListener('mousemove', this.onMouseOver)
  },

  beforeUnmount () {
    this.unsubscribeFromBuses()
    this.unsubscribeSessions()
    this.unsubscribeUsers()
  },

  data () {
    return {
      openDrawer: this.$q.screen.width < 1023?false:true,
      revealHeader: true,

      desktopTemplatesActive: false,
      controlActive: false,
      settingsActive: false,
      apiExplorerActive: false,
      loginActive: false,
      profileActive: false,
      metricsActive: false,
      audioExpanded: false,

      controlSessions: []
    }
  },

  computed: {
    userInitial () {
      const user = this.userStore.user
      if (user.name !== undefined) {
        return user.name[0]
      }
      return ''
    },
    audioText () {
      if (this.audioEnabled) {
        return 'Mute'
      }
      return 'Unmute'
    },
    audioCaption () {
      if (this.audioEnabled) {
        return 'Audio is currently enabled'
      }
      return 'Audio is currently disabled'
    },
    audioIcon () {
      if (this.audioEnabled) {
        return 'volume_up'
      }
      return 'volume_off'
    },
    audioHeaderClass () {
      if (this.audioEnabled) {
        return 'text-blue'
      }
      return ''
    },
    recordingCaption () {
      if (this.recordingEnabled) {
        return 'Recording is currently enabled'
      }
      return 'Recording is currently disabled'
    },
    recordingIcon () {
      if (this.recordingEnabled) {
        return 'mic'
      }
      return 'mic_off'
    },
    audioEnabled () { return this.desktopSessions.audioEnabled },
    recordingEnabled () { return this.desktopSessions.recordingEnabled },
    user () { return this.userStore.user },
    isLoggedIn () { return this.userStore.isLoggedIn },
    grafanaEnabled () {  console.log(this.configStore);  return this.configStore.grafanaEnabled }
  },

  methods: {

    onFocusSyncRemoteClipboardListener () {
      window.addEventListener('focus', this.onPaste)
    },

    async onPaste () {
      try {
        const text = await navigator.clipboard.readText()
        this.$root?.$emit('paste-clipboard', text)
      } catch (err) {
        console.log(err)
        this.$root?.$emit('notify-error', new Error('This browser does not appear to support retrieving clipboard text'))
      }
    },

    async onFileTransfer () {
      const activeSession = this.desktopSessions.activeSession
      if (activeSession === undefined) {
        return
      }
      await this.$q.dialog({
        component: FileTransferDialog,
        parent: this,
        desktopNamespace: activeSession.namespace,
        desktopName: activeSession.name
      }).onOk(() => {
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onClickDesktopTemplates () {
      this.desktopTemplatesActive = true

      this.controlActive = false
      this.settingsActive = false
      this.apiExplorerActive = false
      this.loginActive = false
      this.profileActive = false
      this.metricsActive = false
    },

    onClickControl () {
      this.controlActive = true

      this.desktopTemplatesActive = false
      this.settingsActive = false
      this.apiExplorerActive = false
      this.loginActive = false
      this.profileActive = false
      this.metricsActive = false
    },

    onClickSettings () {
      this.settingsActive = true

      this.desktopTemplatesActive = false
      this.controlActive = false
      this.apiExplorerActive = false
      this.loginActive = false
      this.profileActive = false
      this.metricsActive = false
    },

    onClickAPIExplorer () {
      this.apiExplorerActive = true

      this.desktopTemplatesActive = false
      this.controlActive = false
      this.settingsActive = false
      this.loginActive = false
      this.profileActive = false
      this.metricsActive = false
    },

    onClickLogin () {
      this.loginActive = true

      this.profileActive = false
      this.apiExplorerActive = false
      this.desktopTemplatesActive = false
      this.controlActive = false
      this.settingsActive = false
      this.metricsActive = false
    },

    onClickProfile () {
      this.profileActive = true

      this.loginActive = false
      this.apiExplorerActive = false
      this.desktopTemplatesActive = false
      this.controlActive = false
      this.settingsActive = false
      this.metricsActive = false
    },

    onClickMetrics () {
      this.metricsActive = true

      this.loginActive = false
      this.apiExplorerActive = false
      this.desktopTemplatesActive = false
      this.controlActive = false
      this.settingsActive = false
      this.profileActive = false
    },

    onClickAudio () {
      this.desktopSessions.toggleAudio(!this.audioEnabled)
    },

    onClickRecord () {
      this.desktopSessions.toggleRecording(!this.recordingEnabled)
    },

    async handleLoggedIn () {
      await this.configStore.getServerConfig()
      this.onClickDesktopTemplates()
      this.pushIfNotCurrent('templates')
    },

    async onClickLogout () {
      this.desktopSessions.clearSessions()
      this.userStore.logout()
      this.onClickLogin()
      this.pushIfNotCurrent('login')
    },

    pushIfNotCurrent (route: any) {
      if (this.$router.currentRoute.name !== route) {
        this.$router.push(route)
      }
    },

    subscribeToBuses () {
      this.configStore.emitter.on('notify-error', this.notifyError)
      this.configStore.emitter.on('set-control', this.onClickControl)
      this.unsubscribeSessions = this.desktopSessions.$subscribe(this.handleSessionsChange)
    },

    unsubscribeFromBuses () {
      this.configStore.emitter.off('notify-error', this.notifyError)
      this.configStore.emitter.off('set-control', this.onClickControl)
      this.unsubscribeSessions()
    },

    handleSessionsChange (_mutation: any, state: any) {
      this.audioExpanded = state.audioEnabled
      this.controlSessions = state.sessions
    },

    async notifyError (err: any) {
      const errMsg = await getErrorMessage(err)
      this.$q.notify({
        color: 'red-4',
        textColor: 'black',
        icon: 'error',
        message: errMsg
      })
    },

    onMouseOver (event: any) {
      if (document.fullscreenElement) {
        if (event.pageX < 20) {
          // Show the menu if mouse is within 20 pixels
          // from the left or we are hovering over it
          clearTimeout(menuTimeout)
          menuTimeout = null
          this.openDrawer = true
        } else if (menuTimeout === null && event.pageX > 20) {
          // Hide the menu if the mouse is further than 20 pixels
          // from the left and it is not hovering over the menu
          // and we aren't already scheduled to hide it
          menuTimeout = setTimeout(() => { this.openDrawer = false }, 1000)
        }
      }
    },

    handleFullScreenChange () {
      if (document.fullscreenElement) {
        console.log('Entered fullscreen mode')
        this.openDrawer = false
        this.revealHeader = false
        this.$root?.$emit('set-fullscreen', true)
      } else {
        console.log('Leaving full-screen mode')
        this.openDrawer = true
        this.revealHeader = true
        this.$root?.$emit('set-fullscreen', false)
      }
    }
  }
})
</script>
