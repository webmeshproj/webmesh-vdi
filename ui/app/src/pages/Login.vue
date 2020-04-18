<template>
  <div class="fixed-center text-center">
    <q-form
      @submit="onSubmit"
      @reset="onReset"
      class="q-gutter-md"
    >
      <q-input
        filled
        rounded
        outlined
        v-model="username"
        label="Username"
        lazy-rules
        :rules="[ val => val && val.length > 0 || 'Username cannot be blank']"
      />

      <q-input
        filled
        rounded
        outlined
        type="password"
        v-model="password"
        label="Password"

        lazy-rules
        :rules="[
          val => val !== null && val !== '' || 'Password cannot be blank',
        ]"
      />
      <div>
        <q-btn label="Login" type="submit" color="primary"/>
        <q-btn label="Reset" type="reset" color="primary" flat class="q-ml-sm" />
      </div>
    </q-form>

  </div>
</template>

<script >
export default {
  name: 'Login',

  data () {
    return {
      username: null,
      password: null
    }
  },

  methods: {
    onSubmit () {
      this.$q.notify({
        color: 'green-4',
        textColor: 'white',
        icon: 'cloud_done',
        message: 'Submitted'
      })
    },

    onReset () {
      this.username = null
      this.password = null
    }
  }
}
</script>
