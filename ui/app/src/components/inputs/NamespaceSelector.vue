<template>
  <q-select
    v-model="selection"
    :label="label"
    use-input clearable dense
    :use-chips="multiSelect"
    :multiple="multiSelect"
    :loading="loading"
    transition-show="scale"
    transition-hide="scale"
    :virtual-scroll-slice-size="5"
    @filter="filterFn"
    :options="options"
  >
    <template v-slot:no-option>
      <q-item>
        <q-item-section class="text-grey">
          No results
        </q-item-section>
      </q-item>
    </template>
  </q-select>
</template>

<script>
export default {
  name: 'NamespaceSelector',
  props: {
    showAllOption: {
      type: Boolean
    },
    multiSelect: {
      type: Boolean
    },
    label: {
      type: String,
      default: 'Namespaces'
    }
  },
  data () {
    return {
      loading: false,
      selection: [],
      options: []
    }
  },
  methods: {
    async filterFn (val, update) {
      this.loading = true
      try {
        const res = await this.$axios.get('/api/namespaces')
        if (res.data.length === 1) {
          this.namespaces = res.data[0]
        }
        if (val === '') {
          update(() => {
            if (this.showAllOption) {
              this.options = res.data.unshift('*')
            } else {
              this.options = res.data
            }
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.options = res.data.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.loading = false
    }
  }
}
</script>
