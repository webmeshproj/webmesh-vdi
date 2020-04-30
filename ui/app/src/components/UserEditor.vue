<template>
  <q-card style="min-width: 350px">

    <q-card-section v-if="editorFunction == 'create'">
      <q-input dense debounce="500" label="Username" v-model="username" :rules="[validateUser]"/>
    </q-card-section>

    <q-card-section class="q-pt-none">
      <PasswordInput ref="password" :startDisabled="passwordIsDisabled" />
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
import PasswordInput from 'components/PasswordInput.vue'

export default {
  name: 'UserEditor',
  components: { PasswordInput },
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
      roleSelection: [],
      roles: [],
      loading: true
    }
  },
  computed: {
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
      return true
    }
  },
  methods: {

    async validateUser (val) {
      if (!val) {
        return 'Username is required'
      }
      try {
        await this.$axios.get(`/api/users/${val}`)
        return 'User already exists'
      } catch (err) {}
    },

    async addUser () {
      const payload = {
        username: this.username,
        password: this.$refs.password.password,
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
        payload.password = this.$refs.password.password
      }
      try {
        await this.$axios.put(`/api/users/${this.userToEdit}`, payload)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `User '${this.userToEdit}' updated successfully`
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
          roles.push(role.metadata.name)
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
            this.roleSelection = roles
            this.loading = false
            this.$refs.password.password = '*******************'
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
