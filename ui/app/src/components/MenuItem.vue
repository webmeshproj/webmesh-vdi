<template>
  <q-item
    v-if="children === undefined || children.length === 0"
    clickable
    tag="a"
    :href="`#/${link}`"
    :active="active"
    :dense="isChild"
    @click="() => { onClick(parent) }"
  >
    <q-item-section v-if="icon" avatar>
      <q-icon :name="icon" />
    </q-item-section>

    <q-item-section>
      <q-item-label>{{ title }}</q-item-label>
      <q-item-label caption>{{ caption }}</q-item-label>
    </q-item-section>

  </q-item>

  <q-expansion-item
    v-else
    :label="title"
    :caption="caption"
    :icon="icon"
    :to="link"
    :active="active"
    group="menu"
    :content-inset-level="0.2"
    @click="() => { onClick(parent) }"
  >
  <q-list>
    <MenuItem
      v-for="child in children"
      :key="child.title"
      v-bind="child"
      :isChild="true"
      :onClick="() => { child.onClick(parent, child) }"
    />
  </q-list>
  </q-expansion-item>
</template>

<script>
export default {
  name: 'MenuItem',
  data () {
    return {
      parent: this,
      active: false
    }
  },
  props: {

    onClick: {
      type: Function,
      required: true
    },

    title: {
      type: String,
      required: true
    },

    caption: {
      type: String,
      default: ''
    },

    link: {
      type: String,
      default: ''
    },

    icon: {
      type: String,
      default: ''
    },

    children: {
      type: Array,
      required: false
    },

    isChild: {
      type: Boolean,
      default: false
    }
  },

  created () {
    this.$root.$on('set-active-title', this.setActive)
  },

  beforeDestroy () {
    this.$root.$off('set-active-title', this.setActive)
  },

  methods: {
    setActive (title) {
      if (this.title === title) {
        this.active = true
      } else {
        this.active = false
      }
    }
  }
}
</script>
