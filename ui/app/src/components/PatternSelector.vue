<template>
  <q-select
    label="Resource patterns (Use '.*' All)"
    v-model="selection"
    use-input
    use-chips
    bottom-slots
    multiple
    clearable
    dense
    hide-dropdown-icon
    input-debounce="0"
    new-value-mode="add-unique"
    :v-close-popup="false"
    hint="Press Enter to add patterns"
  />
</template>

<script>
export default {
  name: 'PatternSelector',
  data () {
    return { selection: [] }
  }
}
</script>
