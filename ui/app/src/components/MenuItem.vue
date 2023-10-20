<!--
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
-->

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

<script lang="ts">
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
    this.configStore.emitter.on('set-active-title', this.setActive)
  },

  beforeUnmount () {
    this.configStore.emitter.off('set-active-title', this.setActive)
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
