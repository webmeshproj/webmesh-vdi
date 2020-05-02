<template>
  <q-page class="flex">
    <div v-if="noTemplatesBanner" class="q-pa-md q-gutter-sm">
      <q-banner inline-actions rounded class="bg-orange text-white">
        Could not find any available desktop templates to launch.
        <template v-slot:action>
          <q-btn flat label="Refresh" @click="refreshTemplates" :loading="refreshLoading" />
          <q-btn flat label="Dismiss" @click="dismissBanner = true" />
        </template>
      </q-banner>
    </div>
    <div class="q-pa-md row items-start q-gutter-md">

      <div v-if="loading">

        <q-card style="max-width: 300px">
          <q-item>
            <q-item-section avatar>
              <q-skeleton type="QAvatar" />
            </q-item-section>

            <q-item-section>
              <q-item-label>
                <q-skeleton type="text" />
              </q-item-label>
              <q-item-label caption>
                <q-skeleton type="text" />
              </q-item-label>
            </q-item-section>
          </q-item>

          <q-skeleton height="200px" square />

          <q-card-actions align="right" class="q-gutter-md">
            <q-skeleton type="QBtn" />
            <q-skeleton type="QBtn" />
          </q-card-actions>
        </q-card>

      </div>

      <DesktopTemplateCard
        v-for="item in items"
        :key="item.metadata.name"
        v-bind="item"
      />

    </div>
  </q-page>
</template>

<script>
import DesktopTemplateCard from 'components/DesktopTemplateCard.vue'

export default {
  name: 'DesktopTemplates',

  components: {
    DesktopTemplateCard
  },

  data () {
    return {
      loading: false,
      refreshLoading: false,
      dismissBanner: false,
      items: []
    }
  },

  computed: {
    noTemplatesBanner () {
      return !this.loading && !this.dismissBanner && this.items.length === 0
    }
  },

  methods: {
    async refreshTemplates () {
      this.refreshLoading = true
      try {
        const response = await this.$axios.get('/api/templates')
        this.items = response.data
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.refreshLoading = false
    },

    async fetchTemplates () {
      this.loading = true
      try {
        const response = await this.$axios.get('/api/templates')
        this.items = response.data
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.loading = false
    }
  },

  mounted () {
    this.$nextTick().then(() => {
      this.fetchTemplates()
    })
  }

}
</script>
