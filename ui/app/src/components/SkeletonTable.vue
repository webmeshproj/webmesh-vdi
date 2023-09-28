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
      <tr v-for="n in 3" :key="n">
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

<script lang="ts">
export default {
  data () {
    return {
      skeletonClass: 'bg-gray',
      skeletonAnimation: 'wave'
    }
  }
}
</script>
