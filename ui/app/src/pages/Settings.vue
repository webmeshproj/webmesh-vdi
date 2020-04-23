<template>
  <q-page>

    <q-tabs
      v-model="tab"
      class="bg-grey-3"
      align="justify"
      narrow-indicator
    >
      <q-tab name="users" label="Users" />
      <q-tab name="roles" label="Roles" />
      <q-tab name="config" label="Configuration" />
    </q-tabs>

    <q-separator />

    <q-tab-panels v-model="tab" animated style="text-align: center">
      <q-tab-panel name="users" >
        <UserPanel />
      </q-tab-panel>

      <q-tab-panel name="roles">
        <RolePanel />
      </q-tab-panel>

      <q-tab-panel name="config">
        <VDIConfigPanel />
      </q-tab-panel>
    </q-tab-panels>

  </q-page>
</template>

<script>
import UserPanel from 'components/UserPanel'
import RolePanel from 'components/RolePanel'
import VDIConfigPanel from 'components/VDIConfigPanel'

export default {
  name: 'Settings',
  components: {
    UserPanel,
    RolePanel,
    VDIConfigPanel
  },
  data () {
    return {
      tab: 'users'
    }
  }
}
</script>
