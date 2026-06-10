<template>
  <div class="control-view">
    <h1 class="page-title">设备控制</h1>
    <div class="control-grid">
      <TripleScreen />
      <AmbientLight />
    </div>
  </div>
</template>

<script setup>
import TripleScreen from '../components/devices/TripleScreen.vue'
import AmbientLight from '../components/devices/AmbientLight.vue'
</script>

<style scoped>
.page-title { font-size: 22px; margin-bottom: 24px; }
.control-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}
</style>
