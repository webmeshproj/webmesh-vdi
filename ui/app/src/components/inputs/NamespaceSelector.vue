<!--
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
-->

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
          this.options = [res.data[0]]
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
