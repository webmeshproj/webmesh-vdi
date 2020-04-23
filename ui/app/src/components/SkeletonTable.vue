<template>
  <q-markup-table>
    <thead>
      <tr>
        <th class="text-left" style="width: 150px">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
        <th class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
        <th class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
        <th class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
        <th class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
        <th class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" />
        </th>
      </tr>
    </thead>

    <tbody>
      <tr v-for="n in 5" :key="n">
        <td class="text-left">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="85px" />
        </td>
        <td class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="50px" />
        </td>
        <td class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="35px" />
        </td>
        <td class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="65px" />
        </td>
        <td class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="25px" />
        </td>
        <td class="text-right">
          <q-skeleton :class="skeletonClass" :animation="skeletonAnimation" type="text" width="85px" />
        </td>
      </tr>
    </tbody>
  </q-markup-table>
</template>

<script>
export default {
  data () {
    return {
      skeletonClass: 'bg-gray',
      skeletonAnimation: 'wave'
    }
  }
}
</script>

<style>

</style>
