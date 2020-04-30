<template>
  <q-input dense label="Password" v-model="password" :type="pwPromptType" :rules="[validatePassword]" :bottom-slots="passwordIsDisabled" :disabled="passwordIsDisabled">
    <template v-slot:append>
      <q-btn round dense flat :icon="revealIcon"  size="sm" color="grey" @click="revealPassword" :disabled="passwordIsDisabled">
        <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">{{ pwPromptToolTip }}</q-tooltip>
      </q-btn>
      <q-btn round dense flat icon="loop" size="sm" color="teal" @click="generatePassword" :disabled="passwordIsDisabled">
        <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Generate password</q-tooltip>
      </q-btn>
    </template>
    <template v-slot:hint>
      <div style="float: right;" v-if="passwordIsDisabled">
        <q-btn dense right flat size="sm" color="blue" label="Reset password" @click="onEditPassword" />
        <br />
      </div>
    </template>
  </q-input>
</template>

<script>
const CharacterSet = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789![]{}()%&*$#^<>~@|'
const PasswordSize = 16

export default {
  name: 'PasswordInput',
  props: {
    startDisabled: {
      type: Boolean,
      default: false
    }
  },
  data () {
    return {
      password: '',
      passwordIsDisabled: this.startDisabled,
      passwordVisible: false
    }
  },
  computed: {
    revealIcon () {
      if (!this.passwordVisible) {
        return 'visibility'
      }
      return 'visibility_off'
    },
    pwPromptType () {
      if (!this.passwordVisible) {
        return 'password'
      }
      return ''
    },
    pwPromptToolTip () {
      if (!this.passwordVisible) {
        return 'Reveal password'
      }
      return 'Hide password'
    }
  },
  methods: {
    onEditPassword () {
      this.passwordIsDisabled = !this.passwordIsDisabled
      this.$root.$emit('edit-password')
    },

    validatePassword (val) {
      if (!val) {
        return 'Password is required'
      }
    },

    generatePassword () {
      this.generate()
      this.passwordVisible = true
    },

    revealPassword () {
      this.passwordVisible = !this.passwordVisible
    },

    generate () {
      let password = ''
      for (let i = 0; i < PasswordSize; i++) {
        password += CharacterSet.charAt(Math.floor(Math.random() * CharacterSet.length))
      }
      this.password = password
    }
  }
}
</script>
