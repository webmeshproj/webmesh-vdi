<template>
  <q-page class="flex">
    <div class="q-pa-md row items-start q-gutter-md">
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
      items: []
    }
  },

  methods: {
    async fetchTemplates () {
      try {
        const response = await this.$axios.get('/api/templates')
        this.items = response.data
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }
  },

  mounted () {
    this.$nextTick().then(() => {
      this.fetchTemplates()
    })
  }

}
</script>
