<template>
  <div class="q-pa-md" stretch>
    <div style="float: right">
      <q-btn flat color="primary" icon-right="add" label="New User" />
    </div>
    <div style="clear: right">
      <SkeletonTable v-if="loading"/>
        <q-table
          class="user-table"
          title="Users"
          :data="data"
          :columns="columns"
          row-key="name"
        >
          <template v-slot:body="props">
            <q-tr :props="props">
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
              <q-td key="grants" :props="props">
                <div v-for="r in props.row.roles" :v-bind="r" :key="r.name" style="float: left;">
                  <div v-for="grant in r.grants" :v-bind="grant" :key="grant" style="float: left;">
                    <q-badge color="green">
                      {{ grant }}
                    </q-badge>
                    &nbsp;
                  </div>
                </div>
              </q-td>
            </q-tr>
          </template>
        </q-table>
      </div>
  </div>
</template>

<script>
import SkeletonTable from 'components/SkeletonTable'

const userColumns = [
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
    name: 'grants',
    align: 'center',
    label: 'Grants'
  }
]

export default {
  name: 'UserPanel',
  components: { SkeletonTable },

  data () {
    return {
      loading: true,
      data: [],
      columns: userColumns
    }
  },

  methods: {
    async fetchData () {
      try {
        const res = await this.$axios.get('/api/users')
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
.user-table
  /* height or max-height is important */
  height: 310px

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
