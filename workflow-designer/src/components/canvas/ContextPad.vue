<script setup lang="ts">
import { computed } from 'vue'
import { useVueFlow } from '@vue-flow/core'
import { useWorkflowStore } from '../../stores/workflow'
import { useHistory } from '../../composables/useHistory'
import { useActivityRegistry } from '../../composables/useActivityRegistry'
import type { PaletteItem } from '../../types/workflow'

const store = useWorkflowStore()
const { viewport } = useVueFlow()
const { pushHistory } = useHistory()
const { categories: registryCategories } = useActivityRegistry()

const node = computed(() => store.selectedNode)

// Position the pad to the right of the selected node, vertically centered
const padStyle = computed(() => {
  if (!node.value) return { display: 'none' }
  const n = node.value
  const zoom = viewport.value.zoom
  const x = n.position.x * zoom + viewport.value.x
  const y = n.position.y * zoom + viewport.value.y
  const isSmall = n.type === 'startEvent' || n.type === 'endEvent' || n.type === 'exclusiveGateway'
  const nodeWidth = isSmall ? 60 : 200
  const nodeHeight = isSmall ? 60 : 60
  return {
    position: 'absolute' as const,
    left: `${x + nodeWidth * zoom + 8}px`,
    top: `${y + (nodeHeight * zoom) / 2}px`,
    transform: 'translateY(-50%)',
    zIndex: 100,
  }
})

let appendCounter = 0

function appendNode(item: PaletteItem) {
  if (!node.value) return
  const sourceNode = node.value
  const newId = `step-${item.activityName || item.kind}-${++appendCounter}`
  const newX = sourceNode.position.x + 250
  const newY = sourceNode.position.y

  if (item.kind === 'activity' && item.activityName) {
    store.addNode({
      id: newId,
      type: 'activity',
      position: { x: newX, y: newY },
      data: {
        label: item.displayName,
        activityName: item.activityName,
        taskQueue: item.defaultTaskQueue || 'order-task-queue',
        timeoutSeconds: 30,
        retryPolicy: { maxAttempts: 3, initialIntervalSec: 1, backoffCoefficient: 2.0 },
        inputMapping: item.defaultInputMapping || {},
      },
    } as any)
  } else if (item.kind === 'startEvent') {
    store.addNode({
      id: newId,
      type: item.kind,
      position: { x: newX, y: newY },
      data: { label: item.displayName, startConfig: { triggerType: 'manual', formFields: [] } },
    } as any)
  } else if (item.kind === 'exclusiveGateway') {
    store.addNode({
      id: newId,
      type: item.kind,
      position: { x: newX, y: newY },
      data: { label: item.displayName },
    } as any)
  } else {
    store.addNode({
      id: newId,
      type: item.kind,
      position: { x: newX, y: newY },
      data: { label: item.displayName, endConfig: { endType: 'none', outputVariables: [] } },
    } as any)
  }

  const edgePayload: any = {
    id: `e-${sourceNode.id}-${newId}`,
    source: sourceNode.id,
    target: newId,
  }
  if (sourceNode.type === 'exclusiveGateway') {
    edgePayload.data = { condition: { expression: '', isDefault: false } }
  }
  store.addEdge(edgePayload)
  pushHistory()
}

function appendTask() {
  const activities = registryCategories.value.flatMap((c) => c.items)
  if (activities.length > 0) {
    appendNode(activities[0])
  } else {
    appendNode({
      kind: 'activity',
      displayName: 'New Task',
      activityName: 'generic-task',
      defaultTaskQueue: 'default-task-queue',
    } as PaletteItem)
  }
}

function appendGateway() {
  appendNode({
    kind: 'exclusiveGateway',
    displayName: 'Gateway',
  } as PaletteItem)
}

function appendIntermediateEvent() {
  appendNode({
    kind: 'activity',
    displayName: 'Timer Event',
    activityName: 'timer-event',
    defaultTaskQueue: 'default-task-queue',
  } as PaletteItem)
}

function appendEndEvent() {
  appendNode({
    kind: 'endEvent',
    displayName: 'End',
  } as PaletteItem)
}

function deleteSelected() {
  if (node.value) {
    store.removeNode(node.value.id)
    pushHistory()
  }
}
</script>

<template>
  <div v-if="node" :style="padStyle" class="context-pad">
    <!-- Arrow / connection drag handle -->
    <button class="pad-icon" title="Connect" @click.stop>
      <svg width="22" height="22" viewBox="0 0 22 22">
        <path d="M4 11h11M12 7l4 4-4 4" fill="none" stroke="#8899aa" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <!-- Task: rounded rect with person silhouette -->
    <button class="pad-icon" title="Append Task" @click="appendTask">
      <svg width="22" height="22" viewBox="0 0 22 22">
        <rect x="2" y="4" width="18" height="14" rx="3" ry="3" fill="none" stroke="#4a9ade" stroke-width="1.5" />
        <circle cx="11" cy="9.5" r="2" fill="none" stroke="#4a9ade" stroke-width="1.2" />
        <path d="M7.5 16c0-2 1.6-3 3.5-3s3.5 1 3.5 3" fill="none" stroke="#4a9ade" stroke-width="1.2" stroke-linecap="round" />
      </svg>
    </button>

    <!-- Exclusive Gateway: diamond with X -->
    <button class="pad-icon" title="Append Gateway" @click="appendGateway">
      <svg width="22" height="22" viewBox="0 0 22 22">
        <polygon points="11,2 20,11 11,20 2,11" fill="none" stroke="#e8a230" stroke-width="1.5" />
        <line x1="8.5" y1="8.5" x2="13.5" y2="13.5" stroke="#e8a230" stroke-width="2" stroke-linecap="round" />
        <line x1="13.5" y1="8.5" x2="8.5" y2="13.5" stroke="#e8a230" stroke-width="2" stroke-linecap="round" />
      </svg>
    </button>

    <!-- Intermediate Catch Event: double circle with timer -->
    <button class="pad-icon" title="Append Intermediate Event" @click="appendIntermediateEvent">
      <svg width="22" height="22" viewBox="0 0 22 22">
        <circle cx="11" cy="11" r="9" fill="none" stroke="#52be80" stroke-width="1.4" />
        <circle cx="11" cy="11" r="7" fill="none" stroke="#52be80" stroke-width="1.2" />
        <!-- Timer clock -->
        <line x1="11" y1="7" x2="11" y2="11" stroke="#52be80" stroke-width="1.3" stroke-linecap="round" />
        <line x1="11" y1="11" x2="14" y2="13" stroke="#52be80" stroke-width="1.3" stroke-linecap="round" />
      </svg>
    </button>

    <!-- End Event: thick circle with filled dot -->
    <button class="pad-icon" title="Append End Event" @click="appendEndEvent">
      <svg width="22" height="22" viewBox="0 0 22 22">
        <circle cx="11" cy="11" r="8.5" fill="none" stroke="#d44a3c" stroke-width="2.5" />
        <circle cx="11" cy="11" r="4" fill="#d44a3c" />
      </svg>
    </button>

    <!-- Divider -->
    <div class="pad-divider"></div>

    <!-- Delete -->
    <button class="pad-icon pad-icon--delete" title="Delete" @click="deleteSelected">
      <svg width="22" height="22" viewBox="0 0 22 22">
        <path d="M7 4.5h8M6 6.5h10M7.5 6.5v10a1.5 1.5 0 001.5 1.5h4a1.5 1.5 0 001.5-1.5v-10" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" />
        <line x1="9.5" y1="9" x2="9.5" y2="15" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
        <line x1="12.5" y1="9" x2="12.5" y2="15" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
.context-pad {
  pointer-events: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 4px;
  background: rgba(0, 0, 0, 0.02);
  border-radius: 6px;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.08);
}

.pad-icon {
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: none;
  color: #8899aa;
  cursor: pointer;
  border-radius: 50%;
  transition: background 0.12s ease, transform 0.1s ease;
  padding: 0;
}

.pad-icon:hover {
  background: rgba(120, 160, 200, 0.15);
}

.pad-icon:active {
  transform: scale(0.88);
}

.pad-divider {
  width: 18px;
  height: 1px;
  background: rgba(120, 140, 160, 0.25);
  margin: 2px 0;
}

.pad-icon--delete {
  color: #8899aa;
}

.pad-icon--delete:hover {
  background: rgba(212, 74, 60, 0.15);
  color: #d44a3c;
}

.pad-icon--delete:active {
  transform: scale(0.88);
}
</style>
