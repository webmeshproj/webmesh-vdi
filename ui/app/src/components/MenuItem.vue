<template>
  <q-item
    v-if="children === undefined || children.length === 0"
    clickable
    tag="a"
    :href="`#/${link}`"
    :active="active"
    :dense="isChild"
    @click="onClick()"
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
    v-model="active"
    :label="title"
    :caption="caption"
    :icon="icon"
    :to="link"
    :active="active"
    group="menu"
    :content-inset-level="0.2"
    @click="onClick()"
  >
  <q-list>
    <MenuItem
      v-for="link in children"
      :key="link.title"
      v-bind="link"
      :isChild="true"
      :onClick="link.onClick"
    />
  </q-list>
  </q-expansion-item>

</template>

<script>
export default {
  name: 'MenuItem',
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

    active: {
      type: Boolean,
      default: false
    },

    children: {
      type: Array,
      required: false
    },

    isChild: {
      type: Boolean,
      default: false
    }
  }
}
</script>
