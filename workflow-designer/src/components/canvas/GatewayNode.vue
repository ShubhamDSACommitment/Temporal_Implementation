<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  id: string
  data: { label: string }
  selected: boolean
}>()

const glowStyle = computed(() =>
  props.selected ? 'filter: drop-shadow(0 0 8px #f39c12)' : ''
)
</script>

<template>
  <div class="gateway-node" :style="glowStyle">
    <Handle type="target" :position="Position.Left" />
    <svg width="60" height="60" viewBox="0 0 60 60">
      <rect
        x="8.79" y="8.79"
        width="42.43" height="42.43"
        rx="4"
        transform="rotate(45 30 30)"
        fill="#1e2a3a"
        stroke="#f39c12"
        stroke-width="2"
      />
      <!-- X marker for exclusive gateway -->
      <line x1="22" y1="22" x2="38" y2="38" stroke="#f39c12" stroke-width="2" stroke-linecap="round" />
      <line x1="38" y1="22" x2="22" y2="38" stroke="#f39c12" stroke-width="2" stroke-linecap="round" />
    </svg>
    <Handle type="source" :position="Position.Right" />
    <div class="gateway-label">{{ data.label }}</div>
  </div>
</template>

<style scoped>
.gateway-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.gateway-label {
  margin-top: 4px;
  font-size: 11px;
  color: #c0c0d0;
  text-align: center;
  max-width: 100px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
