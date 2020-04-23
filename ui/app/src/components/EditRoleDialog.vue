<template>
  <q-card style="min-width: 350px">
    <q-card-section>
      <div class="text-h6">{{ name }}</div>
    </q-card-section>

    <RoleEditor :roleToEdit="name" editorFunction="update" />

  </q-card>
</template>

<script>
import RoleEditor from 'components/RoleEditor'

export default {
  name: 'EditRoleDialog',
  components: { RoleEditor },
  props: {
    name: {
      type: String,
      required: true
    }
  }
}
</script>
