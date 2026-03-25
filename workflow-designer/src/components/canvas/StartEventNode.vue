<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'

defineProps<{
  data: { label: string }
  selected: boolean
}>()
</script>

<template>
  <div class="start-event-node" :class="{ selected }">
    <svg width="60" height="60" viewBox="0 0 60 60">
      <circle cx="30" cy="30" r="28" fill="#1e2a3a" stroke="#27ae60" stroke-width="2" />
      <polygon points="24,18 24,42 44,30" fill="#27ae60" />
    </svg>
    <div class="node-label">{{ data.label }}</div>
    <Handle type="source" :position="Position.Right" :style="{ background: '#27ae60' }" />
  </div>
</template>

<style scoped>
.start-event-node {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.start-event-node.selected svg circle {
  filter: drop-shadow(0 0 8px rgba(39, 174, 96, 0.6));
  stroke-width: 3;
}

.node-label {
  margin-top: 4px;
  font-size: 11px;
  color: #c0c0d0;
  text-align: center;
  white-space: nowrap;
}
</style>
