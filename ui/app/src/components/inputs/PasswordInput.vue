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

<script lang="ts">
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
      this.configStore.emitter.emit('edit-password')
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
