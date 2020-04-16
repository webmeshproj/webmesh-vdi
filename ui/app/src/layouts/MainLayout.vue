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
      </q-toolbar>

      <q-tabs align="left" v-if="controlSessions.length != 0">
        <q-route-tab v-for="tab in controlSessions" v-bind="tab" :key="tab.name" :to="tab.route" :label="tab.name" />
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
          :onClick="() => { setActive(link.title) }"
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

export default {
  name: 'MainLayout',

  components: {
    MenuItem
  },

  created () {
    this.$root.$on('set-active-title', this.setActive)
    document.onfullscreenchange = (event) => {
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
  },

  beforeDestroy () {
    this.$root.$off('set-active-title', this.setActive)
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
          link: 'vnc',
          active: false,
          children: [
            {
              title: 'Fullscreen',
              icon: 'fullscreen',
              link: 'vnc',
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
          title: 'Anonymous',
          caption: 'anon',
          icon: 'supervisor_account',
          link: 'account',
          active: false
        }
      ]
    }
  },

  methods: {
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
