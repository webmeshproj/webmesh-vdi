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
  <q-dialog ref="dialog" @hide="onDialogHide">
    <q-card>
      <q-card-section>
        <div class="text-h6">Enter your two-factor code</div>
        <q-space />
        <div class="q-gutter-md row items-start">
          <q-input mask="#" ref="1" maxlength="1" standout="bg-teal text-white" v-model="d1" @keyup="(ev) => { handleInput(1, ev) }" dense input-style="width:10px" autofocus />
          <q-input mask="#" ref="2" maxlength="1" standout="bg-teal text-white" v-model="d2" @keyup="(ev) => { handleInput(2, ev) }" dense input-style="width:10px"/>
          <q-input mask="#" ref="3" maxlength="1" standout="bg-teal text-white" v-model="d3" @keyup="(ev) => { handleInput(3, ev) }" dense input-style="width:10px"/>
          <q-input mask="#" ref="4" maxlength="1" standout="bg-teal text-white" v-model="d4" @keyup="(ev) => { handleInput(4, ev) }" dense input-style="width:10px"/>
          <q-input mask="#" ref="5" maxlength="1" standout="bg-teal text-white" v-model="d5" @keyup="(ev) => { handleInput(5, ev) }" dense input-style="width:10px"/>
          <q-input mask="#" ref="6" maxlength="1" standout="bg-teal text-white" v-model="d6" @keyup="(ev) => { handleInput(6, ev) }" dense input-style="width:10px"/>
          <q-spinner-grid v-if="loading" color="teal" size="2em" />
        </div>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script lang="ts">

export default {
  name: 'MFADialog',

  data () {
    return {
      d1: '',
      d2: '',
      d3: '',
      d4: '',
      d5: '',
      d6: '',
      loading: false
    }
  },

  methods: {

    show () {
      this.$refs.dialog.show()
    },

    hide () {
      this.$refs.dialog.hide()
    },

    onDialogHide () {
      this.$emit('hide')
    },

    onOKClick () {
      this.$emit('ok')
      this.hide()
    },

    onCancelClick () {
      this.hide()
    },

    async handleInput (idx, ev) {
      if (ev.key === 'Backspace') {
        const prev = idx - 1
        if (this.$refs[prev] !== undefined) {
          this.$refs[prev].focus()
          this[`d${prev}`] = ''
        }
        return
      }
      if (this[`d${idx}`] === '') { return }
      const next = idx + 1
      if (next !== 7) {
        this.$refs[next].focus()
        return
      }
      this.loading = true
      await new Promise((resolve, reject) => setTimeout(resolve, 1000))
      const otp = `${this.d1}${this.d2}${this.d3}${this.d4}${this.d5}${this.d6}`
      try {
        await this.userStore.authorize(otp)
        this.onOKClick()
      } catch (err) {
        this.loading = false
        this.configStore.emitter.emit('notify-error', err)
      }
    }
  }

}
</script>
