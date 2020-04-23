<template>
  <q-card style="min-width: 350px">
    <q-card-section>
      <div class="text-h6">New Role</div>
    </q-card-section>

    <RoleEditor editorFunction="create" />

  </q-card>
</template>

<script>
import RoleEditor from 'components/RoleEditor'

export default {
  name: 'NewRoleDialog',
  components: { RoleEditor }
}
</script>
