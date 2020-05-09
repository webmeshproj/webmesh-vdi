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

<script>

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
        await this.$userStore.dispatch('authorize', otp)
        this.onOKClick()
      } catch (err) {
        this.loading = false
        this.$root.$emit('notify-error', err)
      }
    }
  }

}
</script>
