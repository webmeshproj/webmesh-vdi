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
      <q-btn
        color="primary"
        @click="createNewSession()"
        style="width: 150px"
      >
        Launch Desktop
      </q-btn>
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
    return {}
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

    async createNewSession () {
      try {
        await this.$desktopSessions.dispatch('newSession', this.metadata.name)
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
</style>
