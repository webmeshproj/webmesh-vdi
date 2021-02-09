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
      <q-tab-panel name="users" class="tab-panel-bg">
        <UserPanel />
      </q-tab-panel>

      <q-tab-panel name="roles" class="tab-panel-bg">
        <RolePanel />
      </q-tab-panel>

      <q-tab-panel name="config" class="tab-panel-bg">
        <VDIConfigPanel />
      </q-tab-panel>
    </q-tab-panels>

  </q-page>
</template>

<script>
import UserPanel from 'pages/UserPanel'
import RolePanel from 'pages/RolePanel'
import VDIConfigPanel from 'pages/VDIConfigPanel'

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
