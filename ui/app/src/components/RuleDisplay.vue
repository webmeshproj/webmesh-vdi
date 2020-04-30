<template>
  <div>
    <q-btn v-if="editable" round dense flat icon="remove"  size="sm" color="red" @click="onDeleteRule">
      <q-tooltip anchor="bottom middle" self="top middle" :offset="[10, 10]">Delete Rule</q-tooltip>
    </q-btn>
    <q-btn rounded flat  dense @click="onEditRule">
      <q-breadcrumbs :class="`text-${color}-10`" :active-color="`${color}-10`">
        <template v-slot:separator>
          <q-icon
            size="1.5em"
            name="arrow_forward"
            color="black"
          />
        </template>
        <q-breadcrumbs-el to="" :label="`ACTIONS: ${display('verbs')}`" icon="settings_remote" />
        <q-breadcrumbs-el :label="`RESOURCES: ${display('resources')}`" icon="widgets" />
        <q-breadcrumbs-el :label="`MATCHING: ${display('resourcePatterns')}`" icon="label" />
        <q-breadcrumbs-el :label="`IN NAMESPACES: ${display('namespaces')}`" icon="web" />
      </q-breadcrumbs>
      <q-tooltip v-if="editable" anchor="center right" self="center middle">Click to edit this rule</q-tooltip>
    </q-btn>
    <br />
  </div>
</template>

<script>
import RuleEditor from 'components/RuleEditor.vue'

export default {
  name: 'RuleDisplay',
  props: {
    ruleIdx: { type: Number },
    roleIdx: { type: Number },
    roleName: { type: String },
    verbs: {
      type: Array
    },
    resources: {
      type: Array
    },
    resourcePatterns: {
      type: Array
    },
    namespaces: {
      type: Array
    },
    editable: {
      type: Boolean,
      required: false,
      default: false
    }
  },
  methods: {
    onDeleteRule () {
      this.$root.$emit(this.roleName, {
        roleName: this.roleName,
        roleIdx: this.roleIdx,
        ruleIdx: this.ruleIdx,
        deleteRule: true
      })
    },
    onEditRule () {
      if (!this.editable) {
        console.log('not in editing mode')
        return
      }
      this.$q.dialog({
        component: RuleEditor,
        parent: this,
        verbs: this.verbs,
        resources: this.resources,
        resourcePatterns: this.resourcePatterns,
        namespaces: this.namespaces
      }).onOk((payload) => {
        this.$root.$emit(this.roleName, {
          roleName: this.roleName,
          roleIdx: this.roleIdx,
          ruleIdx: this.ruleIdx,
          payload: payload
        })
      }).onCancel(() => {
        console.log('Cancelled rule edit')
      }).onDismiss(() => {
        // console.log('Called on OK or Cancel')
      })
    },
    display (item) {
      if (this[item] === undefined || this[item].length === 0) {
        return 'NONE'
      }
      let itemAll = false
      this[item].forEach((x) => {
        if (x === '*' || x === '.*') {
          itemAll = true
        }
      })
      if (itemAll) {
        return 'ANY'
      }
      return this[item].join(',')
    }
  },
  computed: {

    color () {
      if (this.editable) {
        return 'blue'
      }
      if (this.fullAccess) {
        return 'green'
      }
      if (this.noAccess) {
        return 'red'
      }
      if (this.limitedAccess) {
        return 'orange'
      }
      return 'cyan'
    },

    fullAccess () {
      return this.display('verbs') === 'ANY' &&
        this.display('resources') === 'ANY' &&
        this.display('resourcePatterns') === 'ANY' &&
        this.display('namespaces') === 'ANY'
    },

    limitedAccess () {
      return this.display('verbs') === 'NONE' ||
        this.display('resources') === 'NONE' ||
        this.display('resourcePatterns') === 'NONE' ||
        this.display('namespaces') === 'NONE'
    },

    noAccess () {
      return this.display('verbs') === 'NONE' &&
        this.display('resources') === 'NONE' &&
        this.display('resourcePatterns') === 'NONE' &&
        this.display('namespaces') === 'NONE'
    }

  }

}
</script>
