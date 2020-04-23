<template>
  <div v-if="loggedIn">
    <q-expansion-item
      group="menu"
      :content-inset-level="0.2"
    >
      <template v-slot:header>
        <q-item-section avatar>
          <q-avatar color="teal" text-color="white">{{ user.name[0] }}</q-avatar>
        </q-item-section>
        <q-item-section>
          {{ user.name }}
        </q-item-section>
      </template>
      <q-list>

        <q-item
          clickable
          tag="a"
          href="#/profile"
          :active="profileLinkActive"
          @click="onProfileClick()"
          dense
        >
        <q-item-section avatar>
          <q-icon name="supervisor_account" />
        </q-item-section>
          <q-item-section>
            <q-item-label>Profile</q-item-label>
          </q-item-section>
        </q-item>

        <q-item
          clickable
          tag="a"
          dense
          @click="onLogout()"
        >
          <q-item-section avatar>
            <q-icon name="desktop_access_disabled" />
          </q-item-section>
          <q-item-section>
            <q-item-label>Logout</q-item-label>
          </q-item-section>
        </q-item>

      </q-list>
    </q-expansion-item>
  </div>

  <div v-else>
    <q-item
      clickable
      tag="a"
      href="#/login"
      :active="loginLinkActive"
    >
      <q-item-section avatar>
        <q-icon name="meeting_room" />
      </q-item-section>
      <q-item-section>
        <q-item-label>Login</q-item-label>
      </q-item-section>
    </q-item>
  </div>
</template>

<script>
export default {
  name: 'UserControls',

  data () {
    return {
      loggedIn: false,
      user: null,
      profileLinkActive: false,
      loginLinkActive: true
    }
  },

  created () {
    this.unsubscribeUsers = this.$userStore.subscribe(this.handleAuthChange)
    this.$root.$on('set-active-title', this.setActive)
    this.$userStore.dispatch('initStore')
      .then(() => {
        if (this.$userStore.getters.isLoggedIn) {
          this.$configStore.dispatch('getServerConfig')
            .catch((err) => {
              this.$root.$emit('notify-error', err)
            })
          this.pushIfNotCurrent('templates')
          this.$root.$emit('set-active-title', 'Desktop Templates')
        } else {
          this.pushIfNotCurrent('login')
        }
      })
      .catch((err) => {
        this.$root.$emit('notify-error', err)
        this.pushIfNotCurrent('login')
      })
  },

  methods: {
    beforeDestroy () {
      this.$root.$off('set-active-title', this.setActive)
      this.unsubscribeUsers()
    },

    onLogout () {
      this.$userStore.dispatch('logout')
      this.$desktopSessions.dispatch('clearSessions')
      this.$root.$emit('set-active-title', 'Login')
    },

    onProfileClick () {
      this.$root.$emit('set-active-title', 'Profile')
    },

    handleAuthChange (mutation, state) {
      if (mutation.type === 'auth_success') {
        this.setLoggedIn()
      } else if (mutation.type === 'logout') {
        this.setLoggedOut()
      }
    },

    setLoggedIn (username) {
      this.user = this.$userStore.getters.user
      this.loggedIn = true
      this.loginLinkActive = false
    },

    setLoggedOut () {
      this.user = null
      this.loggedIn = false
      this.loginLinkActive = true
      this.pushIfNotCurrent('login')
    },

    setActive (title) {
      this.loginLinkActive = false
      this.profileLinkActive = false
      if (title === 'Login') {
        this.loginLinkActive = true
      } else if (title === 'Profile') {
        this.profileLinkActive = true
      }
    },

    pushIfNotCurrent (route) {
      if (this.$router.currentRoute.name !== route) {
        this.$router.push(route)
      }
    }
  }
}
</script>
