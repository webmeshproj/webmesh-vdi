<template>
  <div class="wrapper">
    <q-card-section avatar>
      <div class="container">
        <q-checkbox v-model="enabled" color="teal" @input="enableMFA"/>
        <q-item-label>Enable MFA</q-item-label>
      </div>
      <q-item-label caption>If enabled, you will require an additional OTP at login.</q-item-label>
    </q-card-section>
    <div v-if="enabled && !newUser && provisioningURI !== ''" class="container">
      <q-card-section>
        <qrcode-vue :value="provisioningURI" :size="300" level="H" />
      </q-card-section>
      <br />
      <q-item-label caption>Scan this QR code to setup MFA</q-item-label>
    </div>
  </div>
</template>

<script>
import QrcodeVue from 'qrcode.vue'

export default {
  name: 'MFAConfig',
  components: { QrcodeVue },
  props: {
    username: {
      type: String
    },
    newUser: {
      type: Boolean
    }
  },
  data () {
    return {
      enabled: false,
      provisioningURI: ''
    }
  },
  methods: {
    setMFAData (data) {
      if (data.enabled) {
        this.enabled = true
        this.provisioningURI = data.provisioningURI
      } else {
        this.enabled = false
        this.provisioningURI = ''
      }
      this.$root.$emit('reload-users')
    },
    enableMFA (val) {
      if (this.newUser) { return }
      this.$axios.put(`/api/users/${this.username}/mfa`, { enabled: val })
        .then((res) => {
          this.setMFAData(res.data)
        })
        .catch((err) => {
          this.$root.$emit('notify-error', err)
        })
    }
  },
  mounted () {
    this.$nextTick().then(() => {
      if (this.newUser) { return }
      this.$axios.get(`/api/users/${this.username}/mfa`)
        .then((res) => {
          this.setMFAData(res.data)
        })
        .catch((err) => {
          this.$root.$emit('notify-error', err)
        })
    })
  }
}
</script>

<style scoped>
.wrapper div {
  display: block;
  position: relative;
}
.container div {
  margin: 5px auto;
  display: inline-block;
}
</style>
