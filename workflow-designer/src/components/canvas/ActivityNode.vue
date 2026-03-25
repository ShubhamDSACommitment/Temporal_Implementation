<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'

const props = defineProps<{
  data: {
    label: string
    activityName: string
    taskQueue: string
  }
  selected: boolean
}>()

const colorBar = computed(() => {
  if (props.data.taskQueue === 'payment-task-queue') return '#9b59b6'
  return '#3498db'
})
</script>

<template>
  <div class="activity-node" :class="{ selected }">
    <div class="color-bar" :style="{ background: colorBar }"></div>
    <Handle type="target" :position="Position.Left" />
    <div class="node-content">
      <div class="node-header">
        <svg class="node-icon" width="14" height="14" viewBox="0 0 14 14">
          <circle cx="7" cy="7" r="6" fill="none" stroke="currentColor" stroke-width="1.2" />
          <circle cx="7" cy="7" r="2" fill="currentColor" />
          <line x1="7" y1="1" x2="7" y2="3.5" stroke="currentColor" stroke-width="1" />
          <line x1="7" y1="10.5" x2="7" y2="13" stroke="currentColor" stroke-width="1" />
          <line x1="1" y1="7" x2="3.5" y2="7" stroke="currentColor" stroke-width="1" />
          <line x1="10.5" y1="7" x2="13" y2="7" stroke="currentColor" stroke-width="1" />
        </svg>
        <span class="node-label">{{ data.label }}</span>
      </div>
      <div class="node-details">
        <div class="node-activity">{{ data.activityName }}</div>
        <div class="node-queue">{{ data.taskQueue }}</div>
      </div>
    </div>
    <Handle type="source" :position="Position.Right" />
  </div>
</template>

<style scoped>
.activity-node {
  background: #1e2a3a;
  border: 2px solid #2a4060;
  border-radius: 10px;
  min-width: 180px;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  display: flex;
  overflow: hidden;
}

.activity-node:hover {
  border-color: #3a6090;
}

.activity-node.selected {
  border-color: #5dade2;
  box-shadow: 0 4px 16px rgba(93, 173, 226, 0.3);
}

.color-bar {
  width: 5px;
  flex-shrink: 0;
}

.node-content {
  padding: 10px 14px;
  flex: 1;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.node-icon {
  color: #8090a0;
  flex-shrink: 0;
}

.node-label {
  font-weight: 600;
  font-size: 13px;
  color: #ffffff;
}

.node-details {
  font-size: 11px;
  color: #8090a0;
}

.node-activity {
  font-family: monospace;
  color: #60a0d0;
}

.node-queue {
  margin-top: 2px;
  font-style: italic;
}
</style>
