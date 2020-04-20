<template>
  <q-btn-dropdown
    :unelevated="!active"
    :outline="active"
    :flat="!active"
    dense
    auto-close stretch split
    @click="this.onConnect"
  >
    <template v-slot:label>
      <div>
        <div class="row justify-around items-center no-wrap">
          <q-icon name="laptop" />
        </div>
        <div class="row items-center no-wrap">
          {{ name }}
        </div>
      </div>
    </template>

    <q-list>
      <q-item clickable @click="this.onDisconnect">
        <q-item-section>Disconnect</q-item-section>
      </q-item>
    </q-list>
  </q-btn-dropdown>
</template>

<script>
export default {
  name: 'SessionTab',

  props: {
    name: {
      type: String,
      required: true
    },

    namespace: {
      type: String,
      required: true
    },

    endpoint: {
      type: String,
      required: true
    },

    active: {
      type: Boolean,
      required: false,
      default: false
    }
  },

  methods: {
    onConnect () {
      this.$desktopSessions.dispatch('setActiveSession', this)
      this.$root.$emit('set-active-title', 'Control')
      if (this.$router.currentRoute !== '/control') {
        this.$router.push('control')
      }
    },
    onDisconnect () {
      this.$desktopSessions.dispatch('deleteSession', this)
    }
  }
}
</script>
