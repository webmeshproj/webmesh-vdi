<template>
  <q-layout view="hHh lpR fFf">

    <q-header :value="revealHeader" reveal elevated class="bg-primary text-white" height-hint="98">
      <q-toolbar class="glossy">
        <q-btn dense flat round icon="menu" @click="openDrawer = !openDrawer" />

        <q-toolbar-title>
          <q-avatar>
            <!-- <img src="https://cdn.quasar.dev/logo/svg/quasar-logo.svg"> -->
          </q-avatar>
          kVDI
        </q-toolbar-title>
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
        <MenuItem
          v-for="link in menuItems"
          :key="link.title"
          v-bind="link"
          :onClick="() => {
            if (link.onClick !== undefined) {
              link.onClick()
            } else {
              emitActive(link.title)
            }
          }"
        />
      </q-list>
      <q-separator />
      <UserControls />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

  </q-layout>
</template>

<script>
import MenuItem from 'components/MenuItem'
import UserControls from 'components/UserControls'
import SessionTab from 'components/SessionTab'

export default {
  name: 'MainLayout',

  components: {
    MenuItem,
    SessionTab,
    UserControls
  },

  created () {
    this.subscribeToBuses()
    document.onfullscreenchange = this.handleFullScreenChange
    this.unsubscribeSessions = this.$desktopSessions.subscribe(this.handleSessionsChange)
  },

  beforeDestroy () {
    this.unsubscribeFromBuses()
    this.unsubscribeSessions()
  },

  data () {
    return {
      openDrawer: true,
      revealHeader: true,
      controlSessions: [],
      menuItems: [
        {
          title: 'Desktop Templates',
          caption: 'Containerized workspace environments',
          icon: 'devices',
          link: 'templates',
          active: true
        },
        {
          title: 'Control',
          icon: 'cast',
          link: 'control',
          active: false,
          children: [
            {
              title: 'Fullscreen',
              icon: 'fullscreen',
              link: 'control',
              onClick: () => {
                this.$root.$emit('set-active-title', 'Control')
                this.$q.fullscreen.request()
              }
            }
          ]
        },
        {
          title: 'Settings',
          icon: 'settings',
          link: 'settings',
          active: false
        }
      ]
    }
  },

  methods: {

    pushIfNotCurrent (route) {
      if (this.$router.currentRoute.name !== route) {
        this.$router.push(route)
      }
    },

    subscribeToBuses () {
      this.$root.$on('notify-error', this.notifyError)
    },

    unsubscribeFromBuses () {
      this.$root.$off('notify-error', this.notifyError)
    },

    handleSessionsChange (mutation, state) {
      this.controlSessions = this.$desktopSessions.getters.sessions
    },

    emitActive (title) {
      this.$root.$emit('set-active-title', title)
    },

    notifyError (err) {
      let error
      if (err.response !== undefined && err.response.data !== undefined) {
        error = err.response.data.error
      } else {
        error = err
      }
      this.$q.notify({
        color: 'red-4',
        textColor: 'black',
        icon: 'error',
        message: error
      })
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
