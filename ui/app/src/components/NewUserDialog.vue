<template>
  <q-card style="min-width: 350px">
    <q-card-section>
      <div class="text-h6">New User</div>
    </q-card-section>

    <UserEditor editorFunction="create" />

  </q-card>
</template>

<script>
import UserEditor from 'components/UserEditor'

export default {
  name: 'NewUserDialog',
  components: { UserEditor }
}
</script>
