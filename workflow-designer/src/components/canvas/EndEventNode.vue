<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: { label: string }
  selected: boolean
}>()
</script>

<template>
  <div class="end-event-node" :class="{ selected }">
    <svg width="60" height="60" viewBox="0 0 60 60">
      <circle cx="30" cy="30" r="28" fill="#1e2a3a" stroke="#e74c3c" stroke-width="4" />
      <rect x="22" y="22" width="16" height="16" fill="#e74c3c" />
    </svg>
    <div class="node-label">{{ data.label }}</div>
    <Handle type="target" :position="Position.Left" :style="{ background: '#e74c3c' }" />
  </div>
</template>

<style scoped>
.end-event-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.end-event-node.selected svg circle {
  filter: drop-shadow(0 0 8px rgba(231, 76, 60, 0.6));
  stroke-width: 5;
}

.node-label {
  margin-top: 4px;
  font-size: 11px;
  color: #c0c0d0;
  text-align: center;
  white-space: nowrap;
}
</style>
