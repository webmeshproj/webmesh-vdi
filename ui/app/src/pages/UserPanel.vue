<template>
  <div class="q-pa-md" stretch>

    <div style="float: right">
      <q-btn flat color="primary" icon-right="add" label="New User" @click="onNewUser" />
    </div>

    <div style="clear: right">
      <SkeletonTable v-if="loading"/>

        <q-table
          class="user-table"
          title="Users"
          :data="data"
          :columns="columns"
          row-key="name"
          v-if="!loading"
        >
          <template v-slot:body="props">

            <q-tr :props="props">

              <q-td auto-width>
                <q-btn size="xs" color="primary" round dense @click="props.expand = !props.expand" :icon="props.expand ? 'remove' : 'add'" />
              </q-td>

              <q-td key="name" :props="props">
                <strong>{{ props.row.name }}</strong>
              </q-td>

              <q-td key="roles" :props="props">
                <div v-for="r in props.row.roles" :v-bind="r" :key="r.name">
                  <q-badge color="teal">
                    {{ r.name }}
                  </q-badge>
                  &nbsp;
                </div>
              </q-td>

              <q-td key="mfaEnabled" :props="props">
                <q-avatar v-if="props.row.mfaEnabled" size="27px" font-size="20px" color="green" text-color="white" icon="verified_user" />
                <q-avatar v-if="!props.row.mfaEnabled" size="27px" font-size="20px" color="warning" text-color="white" icon="info" />
              </q-td>

              <q-td key="updateUser">
                <q-btn round dense flat icon="create"  size="sm" color="grey" @click="onEditUser(props.row.name)">
                  <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Edit User</q-tooltip>
                </q-btn>
                <q-btn round dense flat icon="remove_circle"  size="sm" color="red" @click="onConfirmDeleteUser(props.row.name)">
                  <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete User</q-tooltip>
                </q-btn>
              </q-td>

            </q-tr>

            <q-tr v-show="props.expand" :props="props">
               <q-td colspan="100%">
                 <q-markup-table v-for="role in props.row.roles" :key="role.name" v-bind="role">
                   <thead>
                     <tr>
                       <th class="text-left text-black">Role</th>
                       <th class="text-center text-black">Rules</th>
                     </tr>
                   </thead>
                   <tbody>
                     <tr v-for="(rule, idx) in role.rules" :key="idx">
                       <td>{{ role.name }}</td>
                       <td><RuleDisplay v-bind="rule" /></td>
                     </tr>
                   </tbody>
                 </q-markup-table>
               </q-td>
             </q-tr>

          </template>
        </q-table>

      </div>

  </div>
</template>

<script>
import SkeletonTable from 'components/SkeletonTable.vue'
import NewUserDialog from 'components/NewUserDialog.vue'
import EditUserDialog from 'components/EditUserDialog.vue'
import RuleDisplay from 'components/RuleDisplay.vue'
import ConfirmDelete from 'components/ConfirmDelete.vue'

const userColumns = [
  {},
  {
    name: 'name',
    required: true,
    label: 'Username',
    align: 'left',
    field: row => row.name,
    format: val => `${val}`,
    sortable: true,
    classes: 'bg-grey-2 ellipsis',
    headerClasses: 'bg-primary text-white'
  },
  {
    name: 'roles',
    align: 'center',
    label: 'Roles'
  },
  {
    name: 'mfaEnabled',
    align: 'center',
    label: 'MFA Enabled'
  },
  {
    name: 'updateUser',
    align: 'center'
  }
]

export default {
  name: 'UserPanel',
  components: { SkeletonTable, RuleDisplay },

  data () {
    return {
      loading: true,
      data: [],
      columns: userColumns,
      editUserDialog: false,
      editUser: ''
    }
  },

  created () {
    this.$root.$on('reload-users', this.fetchData)
  },

  beforeDestroy () {
    this.$root.$off('reload-users', this.fetchData)
  },

  methods: {
    onNewUser () {
      this.$q.dialog({
        component: NewUserDialog,
        parent: this
      }).onOk(() => {
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onEditUser (username) {
      this.$q.dialog({
        component: EditUserDialog,
        parent: this,
        name: username
      }).onOk(() => {
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    onConfirmDeleteUser (userName) {
      // TODO: There is no server-side check for this yet - and there should be
      if (userName === 'admin') {
        this.$q.notify({
          color: 'red-4',
          textColor: 'white',
          icon: 'warning',
          message: 'You cannot delete the admin user'
        })
        return
      }
      this.$q.dialog({
        component: ConfirmDelete,
        parent: this,
        resourceName: userName
      }).onOk(() => {
        this.doDeleteUser(userName)
      }).onCancel(() => {
      }).onDismiss(() => {
      })
    },

    async doDeleteUser (userName) {
      try {
        await this.$axios.delete(`/api/users/${userName}`)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Deleted user '${userName}'`
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.fetchData()
    },

    async fetchData () {
      try {
        const res = await this.$axios.get('/api/users')
        this.data = res.data
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }
  },

  async mounted () {
    await this.$nextTick()
    this.loading = true
    await new Promise((resolve, reject) => setTimeout(resolve, 500))
    await this.fetchData()
    this.loading = false
  }
}
</script>

<style lang="sass" scoped>
.user-table
  /* height or max-height is important */
  max-height: 500px

  // /* specifying max-width so the example can
  //   highlight the sticky column on any browser window */
  // max-width: 600px

  td:first-child
    /* bg color is important for td; just specify one */
    // background-color: #c1f4cd !important

  tr th
    position: sticky
    /* higher than z-index for td below */
    z-index: 2
    /* bg color is important; just specify one */
    background: #fff

  /* this will be the loading indicator */
  thead tr:last-child th
    /* height of all previous header rows */
    top: 48px
    /* highest z-index */
    z-index: 3
  thead tr:first-child th
    top: 0
    z-index: 1
  tr:first-child th:first-child
    /* highest z-index */
    z-index: 3

  td:first-child
    z-index: 1

  td:first-child, th:first-child
    position: sticky
    left: 0
</style>
