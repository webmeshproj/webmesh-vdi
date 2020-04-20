<template>
  <div class="fixed-center text-center">
    <div class="text-h6"><q-icon name="error_outline" color="primary" x-large />&nbsp;&nbsp;Please login to use kVDI</div>
    <br />
    <q-form
      @submit="onSubmit"
      @reset="onReset"
    >
      <q-input
        :loading="loading"
        input-style="width: 300px;"
        rounded standout
        v-model="username"
        label="Username"
        lazy-rules
        :rules="[ val => val && val.length > 0 || 'Username cannot be blank']"
      />
      <q-input
        rounded standout
        type="password"
        v-model="password"
        label="Password"
      />
      <br />
      <q-btn label="Login" type="submit" color="primary"/>
      <q-btn label="Reset" type="reset" color="primary" flat class="q-ml-sm" />
    </q-form>
  </div>
</template>

<script >
export default {
  name: 'Login',

  data () {
    return {
      username: null,
      password: null,
      loading: false
    }
  },

  methods: {
    async onSubmit () {
      try {
        await this.$userStore.dispatch('login', { username: this.username, password: this.password })
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Logged in as ${this.username}`
        })
        this.$root.$emit('set-logged-in', this.username)
        this.$root.$emit('set-active-title', 'Desktop Templates')
        this.$router.push('templates')
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    },

    onReset () {
      this.username = null
      this.password = null
    }
  },

  mounted () {
    this.$nextTick().then(() => {
      this.$root.$emit('set-active-title', 'Login')
    })
  }
}
</script>
