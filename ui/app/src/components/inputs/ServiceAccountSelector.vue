<template>
  <q-select
    v-model="selection"
    :label="label"
    use-input clearable dense
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
  name: 'ServiceAccountSelector',
  props: {
    idx: {
      type: Number
    },
    parentRefs: {
      type: Object
    },
    label: {
      type: String,
      default: 'ServiceAccounts'
    }
  },
  data () {
    return {
      loading: false,
      selection: '',
      options: []
    }
  },
  methods: {
    async filterFn (val, update) {
      this.loading = true
      try {
        let namespace
        console.log(this.parentRefs)
        const nsRef = this.parentRefs[`ns-${this.idx}`]
        if (!nsRef.selection || typeof (nsRef.selection) === 'object' || nsRef.selection === '' || nsRef.selection === []) {
          namespace = this.$configStore.getters.serverConfig.appNamespace || 'default'
        } else {
          namespace = nsRef.selection
        }
        const res = await this.$axios.get(`/api/serviceaccounts/${namespace}`)
        if (res.data.length === 1) {
          this.options = [res.data[0]]
        }
        if (val === '') {
          update(() => {
            this.options = res.data
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
