<template>
  <q-card style="min-width: 350px">
    <q-card-section>
      <div class="text-h6">{{ name }}</div>
    </q-card-section>

    <UserEditor :userToEdit="name" editorFunction="update" />

  </q-card>
</template>

<script>
import UserEditor from 'components/UserEditor'

export default {
  name: 'EditUserDialog',
  components: { UserEditor },
  props: {
    name: {
      type: String,
      required: true
    }
  }
}
</script>
