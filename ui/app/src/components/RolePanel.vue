<template>
  <div class="q-pa-md" stretch>

    <div style="float: right">
      <q-btn flat color="primary" icon-right="add" label="New Role" @click="onNewRole" />
    </div>

    <q-dialog v-model="newRoleDialog">
      <NewRoleDialog />
    </q-dialog>

    <div style="clear: right">
      <SkeletonTable v-if="loading"/>

      <q-table
        class="roles-table"
        title="Roles"
        :data="data"
        :columns="columns"
        row-key="name"
      >
        <template v-slot:body="props">
          <q-tr :props="props">

            <q-td key="name" :props="props">
              <strong>{{ props.row.name }}</strong>
            </q-td>

            <q-td key="grants" :props="props">
              <div v-for="grant in props.row.grants" :v-bind="grant" :key="grant" style="float: left;">
                <q-badge color="green">
                  {{ grant }}
                </q-badge>
                &nbsp;
              </div>
            </q-td>

            <q-td key="namespaces" :props="props">
              <div v-for="ns in props.row.namespaces" :v-bind="ns" :key="ns" style="float: left;">
                <q-badge color="teal">
                  {{ ns }}
                </q-badge>
                &nbsp;
              </div>
            </q-td>

            <q-td key="templatePatterns" :props="props">
              <div v-for="pattern in props.row.templatePatterns" :v-bind="pattern" :key="pattern" style="float: left;">
                <q-badge color="purple">
                  {{ pattern }}
                </q-badge>
                &nbsp;
              </div>
            </q-td>

            <q-td key="updateRole" :props="props">
              <q-btn round dense flat icon="create"  size="sm" color="grey" @click="onEditRole(props.row.name)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Edit Role</q-tooltip>
              </q-btn>
              &nbsp;
              <q-btn round dense flat icon="remove_circle"  size="sm" color="red" @click="onDeleteRole(props.row.name)">
                <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete Role</q-tooltip>
              </q-btn>
            </q-td>

          </q-tr>
        </template>
      </q-table>
    </div>
    <q-dialog v-model="editRoleDialog">
      <EditRoleDialog :name="editRole"/>
    </q-dialog>
  </div>
</template>

<script>
import SkeletonTable from 'components/SkeletonTable'
import NewRoleDialog from 'components/NewRoleDialog'
import EditRoleDialog from 'components/EditRoleDialog'

const roleColumns = [
  {
    name: 'name',
    required: true,
    label: 'Role',
    align: 'left',
    field: row => row.name,
    format: val => `${val}`,
    sortable: true,
    classes: 'bg-grey-2 ellipsis',
    headerClasses: 'bg-primary text-white'
  },
  {
    name: 'grants',
    align: 'center',
    label: 'Grants'
  },
  {
    name: 'namespaces',
    align: 'center',
    label: 'Namespaces'
  },
  {
    name: 'templatePatterns',
    align: 'center',
    label: 'Template Patterns'
  },
  {
    name: 'updateRole',
    align: 'center'
  }
]

export default {
  name: 'RolesPanel',
  components: { SkeletonTable, NewRoleDialog, EditRoleDialog },

  data () {
    return {
      loading: true,
      data: [],
      columns: roleColumns,
      newRoleDialog: false,
      editRoleDialog: false,
      editRole: ''
    }
  },

  created () {
    this.$root.$on('reload-roles', this.fetchData)
  },

  beforeDestroy () {
    this.$root.$off('reload-roles', this.fetchData)
  },

  methods: {

    onNewRole () {
      this.newRoleDialog = true
    },

    onEditRole (role) {
      this.editRole = role
      this.editRoleDialog = true
    },

    async onDeleteRole (roleName) {
      try {
        await this.$axios.delete(`/api/roles/${roleName}`)
        this.$q.notify({
          color: 'green-4',
          textColor: 'white',
          icon: 'cloud_done',
          message: `Deleted role '${roleName}'`
        })
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
      this.fetchData()
    },

    async fetchData () {
      try {
        const res = await this.$axios.get('/api/roles')
        this.data = res.data
      } catch (err) {
        this.$root.$emit('notify-error', err)
      }
    }

  },

  mounted () {
    this.$nextTick().then(() => {
      this.fetchData().then(() => {
        this.loading = false
      })
    })
  }
}
</script>

<style lang="sass">
.roles-table
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
