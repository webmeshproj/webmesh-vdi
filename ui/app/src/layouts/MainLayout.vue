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
              setActive(link.title)
            }
          }"
        />
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

  </q-layout>
</template>

<script>
import MenuItem from 'components/MenuItem'
import SessionTab from 'components/SessionTab'

export default {
  name: 'MainLayout',

  components: {
    MenuItem,
    SessionTab
  },

  created () {
    this.subscribeToBuses()
    document.onfullscreenchange = this.handleFullScreenChange
    this.unsubscribeUsers = this.$userStore.subscribe(this.handleAuthChange)
    this.unsubscribeSessions = this.$desktopSessions.subscribe(this.handleSessionsChange)
    this.$userStore.dispatch('initStore')
      .then(() => {
        if (this.$userStore.getters.isLoggedIn) {
          this.pushIfNotCurrent('templates')
          this.$root.$emit('set-active-title', 'Desktop Templates')
        } else {
          this.pushIfNotCurrent('login')
        }
      })
      .catch((err) => {
        this.notifyError(err)
        this.pushIfNotCurrent('login')
      })
  },

  beforeDestroy () {
    this.unsubscribeFromBuses()
    this.unsubscribeUsers()
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
          link: 'customize',
          active: false
        },
        {
          title: 'Login',
          icon: 'meeting_room',
          link: 'login',
          active: false,
          hidden: false,
          name: 'login'
        },
        {
          title: 'Anonymous',
          icon: 'supervisor_account',
          active: false,
          hidden: true,
          name: 'user',
          onClick: () => {},
          children: [
            {
              title: 'Log out',
              icon: 'desktop_access_disabled',
              link: 'login',
              onClick: () => {
                this.$userStore.dispatch('logout')
                this.$desktopSessions.dispatch('clearSessions')
                this.$root.$emit('set-logged-out')
                this.pushIfNotCurrent('login')
              }
            }
          ]
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
      this.$root.$on('set-active-title', this.setActive)
      this.$root.$on('set-logged-in', this.setLoggedIn)
      this.$root.$on('set-logged-out', this.setLoggedOut)
      this.$root.$on('notify-error', this.notifyError)
    },

    unsubscribeFromBuses () {
      this.$root.$off('set-active-title', this.setActive)
      this.$root.$off('set-logged-in', this.setLoggedIn)
      this.$root.$off('set-logged-out', this.setLoggedOut)
      this.$root.$off('notify-error', this.notifyError)
    },

    handleAuthChange (mutation, state) {
      if (mutation.type === 'auth_success') {
        this.setLoggedIn(this.$userStore.getters.username)
      }
    },

    handleSessionsChange (mutation, state) {
      this.controlSessions = this.$desktopSessions.getters.sessions
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
    },

    setLoggedIn (username) {
      const newMenuItems = []
      this.menuItems.forEach((item) => {
        if (item.name === 'user') {
          item.hidden = false
          item.title = username
        } else if (item.name === 'login') {
          item.hidden = true
        }
        newMenuItems.push(item)
      })
      this.menuItems = newMenuItems
    },

    setLoggedOut () {
      const newMenuItems = []
      this.menuItems.forEach((item) => {
        if (item.name === 'user') {
          item.hidden = true
        } else if (item.name === 'login') {
          item.active = true
          item.hidden = false
        } else {
          item.active = false
        }
        newMenuItems.push(item)
      })
      this.menuItems = newMenuItems
    },

    setActive (title) {
      const newMenuItems = []
      this.menuItems.forEach((item) => {
        if (title === item.title) {
          console.log(`Setting ${item.title} to active menu item`)
          item.active = true
        } else {
          item.active = false
        }
        newMenuItems.push(item)
      })
      this.menuItems = newMenuItems
    }
  }
}
</script>
