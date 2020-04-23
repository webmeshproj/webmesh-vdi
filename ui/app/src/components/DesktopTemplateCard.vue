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
                <strong>&nbsp;&nbsp;Sound:</strong>&nbsp;{{ spec.config.enableSound || 'false' }}
              </span>
            </q-item-section>
          </q-item>
          <q-item v-ripple>
            <q-item-section avatar>
              <span>
                <q-icon color="primary" name="error" />
                <strong>&nbsp;&nbsp;Root:</strong>&nbsp;{{ spec.config.allowRoot || 'false' }}
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
      <div class="wrapper">
        <div class="namespace-select">
          <q-select
            filled
            v-model="namespaceSelection"
            use-input
            borderless
            dense
            @filter="filterFn"
            label="Namespace: (default)"
            :options="namespaces"
            style="width: 300px"
          >
            <template v-slot:no-option>
              <q-item>
                <q-item-section class="text-grey">
                  No results
                </q-item-section>
              </q-item>
            </template>
          </q-select>
        </div>
        <div class="launch-button">
          <q-btn
            color="primary"
            @click="createNewSession()"
            style="width: 150px"
          >
            Launch Desktop
          </q-btn>
        </div>
      </div>
    </q-card-actions>
  </q-card>
</template>

<script>

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
      namespaceSelection: null,
      namespaces: []
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

    async filterFn (val, update) {
      try {
        const res = await this.$axios.get('/api/namespaces')
        if (res.data.length === 1) {
          this.namespaceSelection = res.data[0]
        }
        if (val === '') {
          update(() => {
            this.namespaces = res.data
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.namespaces = res.data.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async createNewSession () {
      try {
        await this.$desktopSessions.dispatch('newSession', { template: this.metadata.name, namespace: this.namespaceSelection })
        this.$root.$emit('set-active-title', 'Control')
        this.$router.push('control')
      } catch (err) {
        this.$root.$emit('notify-error', err)
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

.namespace-select
  position: absolute
  left: auto
  height: 45px

.launch-button
  position: absolute
  left: 325px

.wrapper
  position: relative
  height: 45px
</style>
