<template>
  <q-layout view="hHh lpR fFf">

    <q-header :value="revealHeader" reveal elevated class="bg-primary text-white" height-hint="98">
      <q-toolbar class="glossy">
        <q-btn dense flat round icon="menu" @click="openDrawer = !openDrawer" />

        <q-toolbar-title>
          <q-avatar>
            <img src="https://cdn.quasar.dev/logo/svg/quasar-logo.svg">
          </q-avatar>
          kVDI
        </q-toolbar-title>

        <q-space />

        <q-btn type="a" href="https://github.com/tinyzimmer/kvdi" target="_blank" stretch flat icon="img:statics/github.png" label="Github" />

      </q-toolbar>

      <q-tabs align="center" v-if="controlSessions.length != 0">
        <SessionTab v-for="tab in controlSessions" v-bind="tab" :key="tab.name" />
      </q-tabs>
    </q-header>

    <q-drawer show-if-above v-model="openDrawer" side="left" behavior="desktop" elevated>
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

            <q-item dense :active="audioEnabled" clickable @click="onClickAudio">

              <q-item-section avatar>
                <q-icon :name="audioIcon" />
              </q-item-section>

              <q-item-section>
                <q-item-label caption>{{ audioCaption }}</q-item-label>
              </q-item-section>

            </q-item>

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

<script>
import SessionTab from 'components/SessionTab.vue'
import MFADialog from 'components/dialogs/MFADialog.vue'
import FileTransferDialog from 'components/dialogs/FileTransfer.vue'
import { getErrorMessage } from '../util/common.js'

var menuTimeout = null

export default {
  name: 'MainLayout',

  components: { SessionTab },

  async created () {
    this.subscribeToBuses()
    document.onfullscreenchange = this.handleFullScreenChange
    this.unsubscribeSessions = this.$desktopSessions.subscribe(this.handleSessionsChange)
    // this.unsubscribeUsers = this.$userStore.subscribe(this.handleAuthChange)
    await this.$userStore.dispatch('initStore')
    try {
      if (this.$userStore.getters.requiresMFA) {
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
      if (this.$userStore.getters.isLoggedIn) {
        this.handleLoggedIn()
      } else {
        this.onClickLogin()
        this.pushIfNotCurrent('login')
      }
    } catch (err) {
      this.$root.$emit('notify-error', err)
      this.onClickLogin()
      this.pushIfNotCurrent('login')
    }
  },

  mounted () {
    window.addEventListener('mousemove', this.onMouseOver)
  },

  beforeDestroy () {
    this.unsubscribeFromBuses()
    this.unsubscribeSessions()
    this.unsubscribeUsers()
  },

  data () {
    return {
      openDrawer: true,
      revealHeader: true,

      desktopTemplatesActive: false,
      controlActive: false,
      settingsActive: false,
      apiExplorerActive: false,
      loginActive: false,
      profileActive: false,
      metricsActive: false,

      controlSessions: []
    }
  },

  computed: {
    userInitial () {
      const user = this.$userStore.getters.user
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
    audioEnabled () { return this.$desktopSessions.getters.audioEnabled },
    user () { return this.$userStore.getters.user },
    isLoggedIn () { return this.$userStore.getters.isLoggedIn },
    grafanaEnabled () { return this.$configStore.getters.grafanaEnabled }
  },

  methods: {

    async onPaste () {
      try {
        const text = await navigator.clipboard.readText()
        this.$root.$emit('paste-clipboard', text)
      } catch (err) {
        console.log('This browser does not appear to support retrieving clipboard text')
      }
    },

    async onFileTransfer () {
      const activeSession = this.$desktopSessions.getters.activeSession
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
      this.$desktopSessions.dispatch('toggleAudio', !this.audioEnabled)
    },

    async handleLoggedIn () {
      await this.$configStore.dispatch('getServerConfig')
      this.onClickDesktopTemplates()
      this.pushIfNotCurrent('templates')
    },

    async onClickLogout () {
      this.$desktopSessions.dispatch('clearSessions')
      this.$userStore.dispatch('logout')
      this.onClickLogin()
      this.pushIfNotCurrent('login')
    },

    pushIfNotCurrent (route) {
      if (this.$router.currentRoute.name !== route) {
        this.$router.push(route)
      }
    },

    subscribeToBuses () {
      this.$root.$on('notify-error', this.notifyError)
      this.$root.$on('set-control', this.onClickControl)
    },

    unsubscribeFromBuses () {
      this.$root.$off('notify-error', this.notifyError)
      this.$root.$off('set-control', this.onClickControl)
    },

    handleSessionsChange (mutation, state) {
      this.controlSessions = this.$desktopSessions.getters.sessions
    },

    async notifyError (err) {
      const errMsg = await getErrorMessage(err)
      this.$q.notify({
        color: 'red-4',
        textColor: 'black',
        icon: 'error',
        message: errMsg
      })
    },

    onMouseOver (event) {
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

    handleFullScreenChange (event) {
      if (document.fullscreenElement) {
        console.log('Entered fullscreen mode')
        this.openDrawer = false
        this.revealHeader = false
        this.$root.$emit('set-fullscreen', true)
      } else {
        console.log('Leaving full-screen mode')
        this.openDrawer = true
        this.revealHeader = true
        this.$root.$emit('set-fullscreen', false)
      }
    }
  }
}
</script>
