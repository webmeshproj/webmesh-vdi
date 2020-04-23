<template>
  <q-card style="min-width: 350px">

    <q-card-section v-if="editorFunction == 'create'">
      <q-input dense debounce="500" label="Username" v-model="username" :rules="[validateUser]"/>
    </q-card-section>

    <q-card-section class="q-pt-none">
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
    </q-card-section>
    <q-card-section class="q-pt-none">
      <q-select
        v-model="roleSelection"
        use-input
        use-chips
        multiple
        dense
        clearable
        :loading="loading"
        @filter="filterFn"
        label="Roles"
        :options="roles"
        :rules="[val => val.length > 0 || 'You must select at least one role']"
      >
        <template v-slot:no-option>
          <q-item>
            <q-item-section class="text-grey">
              No results
            </q-item-section>
          </q-item>
        </template>
      </q-select>
    </q-card-section>

    <q-card-actions align="right" class="text-primary">
      <q-btn flat label="Cancel" v-close-popup />
      <q-btn flat :label="submitLabel" v-close-popup @click="submitFunc" />
    </q-card-actions>
  </q-card>
</template>

<script>

const CharacterSet = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789![]{}()%&*$#^<>~@|'
const PasswordSize = 16

export default {
  name: 'UserEditor',
  props: {
    editorFunction: {
      type: String,
      required: false,
      default: 'create'
    },
    userToEdit: {
      type: String,
      required: false
    }
  },
  data () {
    return {
      username: null,
      password: null,
      passwordVisible: false,
      editPassword: false,
      roleSelection: [],
      roles: [],
      loading: true
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
    },
    submitLabel () {
      if (!this.isUpdating) {
        return 'Create User'
      }
      return 'Update User'
    },
    submitFunc () {
      if (!this.isUpdating) {
        return this.addUser
      }
      return this.updateUser
    },
    isUpdating () {
      return this.editorFunction !== 'create'
    },
    passwordIsDisabled () {
      if (!this.isUpdating) {
        return false
      }
      if (this.isUpdating && this.editPassword) {
        return false
      }
      return true
    }
  },
  methods: {

    onEditPassword () {
      this.password = ''
      this.editPassword = !this.editPassword
    },

    async validateUser (val) {
      if (!val) {
        return 'Username is required'
      }
      try {
        await this.$axios.get(`/api/users/${val}`)
        return 'User already exists'
      } catch (err) {}
    },

    validatePassword (val) {
      if (!this.isUpdating && !val) {
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
    },

    async addUser () {
      const payload = {
        username: this.username,
        password: this.password,
        roles: this.roleSelection
      }
      try {
        await this.$axios.post('/api/users', payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `New user '${this.username}' created`
        })
        this.$root.$emit('reload-users')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async updateUser () {
      const payload = {
        roles: this.roleSelection
      }
      if (this.editPassword) {
        payload.password = this.password
      }
      try {
        await this.$axios.put(`/api/users/${this.userToEdit}`, payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `User '${this.userToEdit}' updated succesfully`
        })
        this.$root.$emit('reload-users')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    async filterFn (val, update) {
      try {
        const res = await this.$axios.get('/api/roles')
        const roles = []
        res.data.forEach((role) => {
          roles.push(role.name)
        })
        if (val === '') {
          update(() => {
            this.roles = roles
          })
        }
        update(() => {
          const needle = val.toLowerCase()
          this.roles = roles.filter(v => v.toLowerCase().indexOf(needle) > -1)
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }

  },
  mounted () {
    this.$nextTick().then(() => {
      if (this.isUpdating) {
        this.$axios.get(`/api/users/${this.userToEdit}`)
          .then((res) => {
            const roles = []
            res.data.roles.forEach((role) => {
              roles.push(role.name)
            })
            this.password = '******************'
            this.roleSelection = roles
            this.loading = false
          })
          .catch((err) => {
            this.$root.$emit('notify-error', err)
          })
      } else {
        this.loading = false
      }
    })
  }
}
</script>

<style>

</style>
