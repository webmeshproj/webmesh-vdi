<template>
  <q-card class="template-card bg-grey-1">
    <q-card-section>
      <div class="row items-center no-wrap">
        <div class="text-h6"><q-icon name="dns" />{{ metadata.name }}</div>
      </div>
    </q-card-section>
    <q-separator />

    <q-card-section>
      <div class="q-pa-md">
        <q-list dense>
          <q-item v-ripple>
            <q-item-section avatar>
              <span>
                <q-icon color="primary" name="dvr" />
                <strong>&nbsp;&nbsp;Image:</strong> {{ spec.image }}
              </span>
            </q-item-section>
          </q-item>
          <q-item v-ripple>
            <q-item-section avatar>
              <span>
                <q-icon color="primary" name="volume_up" />
                <strong>&nbsp;&nbsp;Sound:</strong>&nbsp;{{ spec.config.enableSound }}
              </span>
            </q-item-section>
          </q-item>
          <q-item v-ripple>
            <q-item-section avatar>
              <span>
                <q-icon color="primary" name="error" />
                <strong>&nbsp;&nbsp;Root:</strong>&nbsp;{{ spec.config.allowRoot }}
              </span>
            </q-item-section>
          </q-item>
          <q-item v-ripple>
            <q-item-section avatar>
              <span>
                <q-icon color="primary" name="list_alt" />
                <strong>&nbsp;&nbsp;Tags: </strong>
                  <li class="inline-tags" v-for="tag in tags" :key="tag" dense>
                      <q-chip dense icon="bookmark">{{ tag }}</q-chip>
                  </li>
              </span>
            </q-item-section>
          </q-item>
        </q-list>
      </div>
    </q-card-section>

    <q-separator />

    <q-card-actions>
      <q-btn
        :loading="loading"
        :percentage="percentage"
        :disable="booted"
        color="primary"
        @click="launchDesktop()"
        style="width: 150px"
      >
        Launch Desktop
        <template v-slot:loading>
          <q-spinner-gears class="on-left" />
          Booting...
        </template>
      </q-btn>
      <q-btn v-if="booted" color="teal" style="width: 150px" @click="connectToDesktop()">Connect</q-btn>
    </q-card-actions>
  </q-card>
</template>

<script>
import { setEndpoint } from 'pages/VNCViewer.vue'

export default {
  name: 'DesktopTemplateCard',

  props: {

    metadata: {
      type: Object,
      required: true
    },

    spec: {
      type: Object,
      required: true
    }

  },

  data () {
    return {
      loading: false,
      percentage: 0,
      booted: this.isBooted()
    }
  },

  computed: {
    tags () {
      const tags = []
      Object.keys(this.spec.tags).forEach((key) => {
        tags.push(`${key}: ${this.spec.tags[key]}`)
      })
      return tags
    }
  },

  methods: {
    isBooted () {
      return this.$templateIsBooted(this.metadata.name)
    },

    setBooted () {
      this.$setTemplateBooted(this.metadata.name)
      this.booted = true
    },

    setLoading () {
      console.log(`Launching desktop template ${this.metadata.name}`)
      this.percentage = 0
      this.loading = true
    },

    stopLoading () {
      this.loading = false
    },

    connectToDesktop () {
      console.log(`Connecting to destkop from template ${this.metadata.name}`)
      this.$root.$emit('set-active-title', 'Control')
      this.$router.push('vnc')
    },

    async createNewSession () {
      const session = await this.$axios.post('/api/sessions', { template: this.metadata.name })
      setEndpoint(session.data.endpoint)
      return session.data
    },

    async waitForSessionReady (session) {
      let running = false
      let resolvable = false
      while (!running && !resolvable) {
        const res = await this.$axios.get(`/api/sessions/${session.namespace}/${session.name}`)
        running = res.data.running
        resolvable = res.data.resolvable
        await new Promise(resolve => setTimeout(resolve, 2000))
        if (this.percentage >= 95) {
          this.percentage = 95
        } else {
          this.percentage += 5
        }
      }
    },

    async launchDesktop () {
      this.setLoading()
      try {
        const session = await this.createNewSession()
        this.percentage = 20
        await this.waitForSessionReady(session)
        this.percentage = 100
        this.stopLoading()
        this.setBooted()
      } catch (err) {
        console.error(err)
        this.stopLoading()
      }
    }
  }
}
</script>

<style lang="sass" scoped>
.template-card
  width: 100%
  max-width: 500px

.inline-tags
  display: inline
</style>
