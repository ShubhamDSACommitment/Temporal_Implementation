<script setup lang="ts">
import { markRaw, watch } from 'vue'
import { VueFlow, useVueFlow, MarkerType } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'
import ActivityNode from './ActivityNode.vue'
import StartEventNode from './StartEventNode.vue'
import EndEventNode from './EndEventNode.vue'
import GatewayNode from './GatewayNode.vue'
import ContextPad from './ContextPad.vue'
import { useWorkflowStore } from '../../stores/workflow'
import { useDragAndDrop } from '../../composables/useDragAndDrop'
import { useHistory } from '../../composables/useHistory'

import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const store = useWorkflowStore()
const { onDragOver, onDrop } = useDragAndDrop()
const { pushHistory } = useHistory()

const nodeTypes = {
  activity: markRaw(ActivityNode),
  startEvent: markRaw(StartEventNode),
  endEvent: markRaw(EndEventNode),
  exclusiveGateway: markRaw(GatewayNode),
}

const defaultEdgeOptions = {
  type: 'smoothstep',
  animated: true,
  style: { stroke: '#4a6080', strokeWidth: 2 },
  markerEnd: {
    type: MarkerType.ArrowClosed,
    color: '#4a6080',
  },
}

const { onConnect, onNodeClick, onPaneClick, onEdgeClick } = useVueFlow()

onConnect((params: any) => {
  const sourceNode = store.nodes.find((n) => n.id === params.source)
  const edgeData: any = {}
  if (sourceNode?.type === 'exclusiveGateway') {
    edgeData.data = { condition: { expression: '', isDefault: false } }
  }
  store.addEdge({
    id: `e-${params.source}-${params.target}`,
    source: params.source,
    target: params.target,
    ...edgeData,
  })
  pushHistory()
})

onNodeClick(({ node }: any) => {
  store.selectNode(node.id)
})

onEdgeClick(({ edge }: any) => {
  store.selectEdge(edge.id)
})

onPaneClick(() => {
  store.selectNode(null)
  store.selectEdge(null)
})

watch(() => store.selectedEdgeId, (newId) => {
  store.edges.forEach((edge: any) => {
    edge.class = edge.id === newId ? 'edge-selected' : ''
  })
})
</script>

<template>
  <div class="canvas-container" @drop="onDrop" @dragover="onDragOver">
    <VueFlow
      v-model:nodes="(store.nodes as any)"
      v-model:edges="(store.edges as any)"
      :node-types="(nodeTypes as any)"
      :default-edge-options="(defaultEdgeOptions as any)"
      fit-view-on-init
    >
      <Background :gap="20" :size="1" pattern-color="#2a3040" />
      <Controls />
      <MiniMap
        :node-color="(n: any) => {
          if (n.type === 'startEvent') return '#27ae60'
          if (n.type === 'endEvent') return '#e74c3c'
          if (n.type === 'exclusiveGateway') return '#f39c12'
          return '#3498db'
        }"
        :mask-color="'rgba(20, 24, 37, 0.8)'"
      />
    </VueFlow>
    <ContextPad />
  </div>
</template>

<style scoped>
.canvas-container {
  flex: 1;
  height: 100%;
  position: relative;
}

.canvas-container :deep(.vue-flow) {
  background: #141825;
}

.canvas-container :deep(.vue-flow__edge-textbg) {
  fill: #1e2a3a;
}

.canvas-container :deep(.vue-flow__edge-text) {
  fill: #e0e0e0;
  font-size: 11px;
}

.canvas-container :deep(.edge-selected .vue-flow__edge-path) {
  stroke: #5dade2;
  stroke-width: 3;
  filter: drop-shadow(0 0 4px rgba(93, 173, 226, 0.5));
}

.canvas-container :deep(.vue-flow__minimap) {
  background: #1e2a3a;
  border: 1px solid #2a4060;
  border-radius: 4px;
}
</style>
