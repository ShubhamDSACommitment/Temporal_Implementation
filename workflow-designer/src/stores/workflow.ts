import { defineStore } from 'pinia'
import { ref, computed, type Ref } from 'vue'
import type { Edge } from '@vue-flow/core'
import type { StepDefinition, NodeKind, StartEventConfig, EndEventConfig, EventDefinition, GatewayDefinition, EdgeDefinition, EdgeCondition } from '../types/workflow'

export interface ActivityNodeData {
  label: string
  activityName: string
  taskQueue: string
  timeoutSeconds: number
  retryPolicy: {
    maxAttempts: number
    initialIntervalSec: number
    backoffCoefficient: number
  }
  inputMapping: Record<string, string>
}

export interface EventNodeData {
  label: string
  startConfig?: StartEventConfig
  endConfig?: EndEventConfig
}

export interface GatewayNodeData {
  label: string
}

export interface DesignerNode {
  id: string
  type: string
  kind?: NodeKind
  position: { x: number; y: number }
  data: ActivityNodeData | EventNodeData | GatewayNodeData
}

// Backward compat alias
export type ActivityNode = DesignerNode

export const useWorkflowStore = defineStore('workflow', () => {
  const nodes: Ref<DesignerNode[]> = ref([])
  const edges: Ref<Edge[]> = ref([])
  const selectedNodeId = ref<string | null>(null)
  const selectedEdgeId = ref<string | null>(null)
  const workflowName = ref('Untitled Workflow')
  const workflowDescription = ref('')
  const workflowId = ref('')
  const workflowVersion = ref(1)

  const selectedNode = computed((): DesignerNode | null => {
    if (!selectedNodeId.value) return null
    return nodes.value.find((n) => n.id === selectedNodeId.value) ?? null
  })

  const selectedEdge = computed(() => {
    if (!selectedEdgeId.value) return null
    return edges.value.find((e) => e.id === selectedEdgeId.value) ?? null
  })

  function addNode(node: DesignerNode) {
    nodes.value.push(node)
  }

  function removeNode(id: string) {
    nodes.value = nodes.value.filter((n) => n.id !== id)
    edges.value = edges.value.filter((e) => e.source !== id && e.target !== id)
    if (selectedNodeId.value === id) selectedNodeId.value = null
  }

  function addEdge(edge: Edge) {
    const exists = edges.value.some((e) => e.source === edge.source && e.target === edge.target)
    if (!exists) {
      edges.value.push(edge)
    }
  }

  function removeEdge(id: string) {
    edges.value = edges.value.filter((e) => e.id !== id)
  }

  function selectNode(id: string | null) {
    selectedNodeId.value = id
    selectedEdgeId.value = null
  }

  function selectEdge(id: string | null) {
    selectedEdgeId.value = id
    selectedNodeId.value = null
  }

  function updateNodeData(id: string, data: Partial<ActivityNodeData> | Partial<EventNodeData>) {
    const node = nodes.value.find((n) => n.id === id)
    if (node) {
      node.data = { ...node.data, ...data }
    }
  }

  function clearCanvas() {
    nodes.value = []
    edges.value = []
    selectedNodeId.value = null
    selectedEdgeId.value = null
    workflowName.value = 'Untitled Workflow'
    workflowDescription.value = ''
    workflowId.value = ''
    workflowVersion.value = 1
  }

  function updateEdgeData(edgeId: string, data: { condition?: EdgeCondition; label?: string }) {
    const edge = edges.value.find((e) => e.id === edgeId)
    if (edge) {
      if (data.condition !== undefined) {
        (edge as any).data = { ...(edge as any).data, condition: data.condition }
      }
      if (data.label !== undefined) {
        edge.label = data.label
      }
    }
  }

  function setDefaultEdge(gatewayId: string, edgeId: string) {
    const outgoing = edges.value.filter((e) => e.source === gatewayId)
    for (const e of outgoing) {
      const edgeData = (e as any).data || {}
      if (e.id === edgeId) {
        (e as any).data = { ...edgeData, condition: { expression: '', isDefault: true } }
        e.label = '(default)'
      } else {
        if (edgeData.condition?.isDefault) {
          (e as any).data = { ...edgeData, condition: { ...edgeData.condition, isDefault: false } }
          e.label = edgeData.condition?.expression || ''
        }
      }
    }
  }

  function exportGateways(): GatewayDefinition[] {
    return nodes.value
      .filter((node) => node.type === 'exclusiveGateway')
      .map((node) => ({
        id: node.id,
        type: 'exclusiveGateway' as const,
        label: (node.data as GatewayNodeData).label,
        position: { x: node.position.x, y: node.position.y },
      }))
  }

  function exportEdges(): EdgeDefinition[] {
    return edges.value.map((edge) => {
      const edgeDef: EdgeDefinition = {
        id: edge.id,
        source: edge.source,
        target: edge.target,
      }
      const edgeData = (edge as any).data
      if (edgeData?.condition) {
        edgeDef.condition = edgeData.condition
      }
      if (edge.label) {
        edgeDef.label = edge.label as string
      }
      return edgeDef
    })
  }

  function exportSteps(): StepDefinition[] {
    return nodes.value
      .filter((node) => node.type === 'activity')
      .map((node) => {
        const data = node.data as ActivityNodeData
        // Remap deps: skip event nodes in depends_on
        const deps = edges.value
          .filter((e) => e.target === node.id)
          .map((e) => e.source)
          .filter((srcId) => {
            const srcNode = nodes.value.find((n) => n.id === srcId)
            return srcNode && srcNode.type === 'activity'
          })

        return {
          id: node.id,
          name: data.label,
          activity_name: data.activityName,
          task_queue: data.taskQueue,
          timeout_seconds: data.timeoutSeconds,
          retry_policy: data.retryPolicy.maxAttempts > 0
            ? {
                max_attempts: data.retryPolicy.maxAttempts,
                initial_interval_sec: data.retryPolicy.initialIntervalSec,
                backoff_coefficient: data.retryPolicy.backoffCoefficient,
              }
            : undefined,
          input_mapping: data.inputMapping,
          depends_on: deps,
        }
      })
  }

  function exportEvents(): EventDefinition[] {
    return nodes.value
      .filter((node) => node.type === 'startEvent' || node.type === 'endEvent')
      .map((node) => {
        const data = node.data as EventNodeData
        const eventDef: EventDefinition = {
          id: node.id,
          type: node.type as 'startEvent' | 'endEvent',
          label: data.label,
          position: { x: node.position.x, y: node.position.y },
        }
        if (data.startConfig) eventDef.start_config = data.startConfig
        if (data.endConfig) eventDef.end_config = data.endConfig
        return eventDef
      })
  }

  function loadFromSteps(steps: StepDefinition[], events?: EventDefinition[], gateways?: GatewayDefinition[], edgeDefs?: EdgeDefinition[]) {
    const activityNodes: DesignerNode[] = steps.map((step, i) => ({
      id: step.id,
      type: 'activity',
      kind: 'activity' as NodeKind,
      position: { x: 100 + (i % 3) * 300, y: 100 + Math.floor(i / 3) * 200 },
      data: {
        label: step.name,
        activityName: step.activity_name,
        taskQueue: step.task_queue,
        timeoutSeconds: step.timeout_seconds,
        retryPolicy: {
          maxAttempts: step.retry_policy?.max_attempts || 3,
          initialIntervalSec: step.retry_policy?.initial_interval_sec || 1,
          backoffCoefficient: step.retry_policy?.backoff_coefficient || 2.0,
        },
        inputMapping: step.input_mapping || {},
      },
    }))

    const eventNodes: DesignerNode[] = (events || []).map((evt) => ({
      id: evt.id,
      type: evt.type,
      kind: evt.type as NodeKind,
      position: { x: evt.position.x, y: evt.position.y },
      data: {
        label: evt.label,
        startConfig: evt.start_config,
        endConfig: evt.end_config,
      },
    }))

    const gatewayNodes: DesignerNode[] = (gateways || []).map((gw) => ({
      id: gw.id,
      type: 'exclusiveGateway',
      kind: 'exclusiveGateway' as NodeKind,
      position: { x: gw.position.x, y: gw.position.y },
      data: {
        label: gw.label,
      },
    }))

    nodes.value = [...activityNodes, ...eventNodes, ...gatewayNodes]

    edges.value = []
    if (edgeDefs && edgeDefs.length > 0) {
      for (const ed of edgeDefs) {
        edges.value.push({
          id: ed.id,
          source: ed.source,
          target: ed.target,
          animated: true,
          label: ed.label || '',
          data: ed.condition ? { condition: ed.condition } : undefined,
        } as any)
      }
    } else {
      for (const step of steps) {
        for (const dep of step.depends_on || []) {
          edges.value.push({
            id: `e-${dep}-${step.id}`,
            source: dep,
            target: step.id,
            animated: true,
          })
        }
      }
    }
  }

  return {
    nodes,
    edges,
    selectedNodeId,
    selectedEdgeId,
    selectedNode,
    selectedEdge,
    workflowName,
    workflowDescription,
    workflowId,
    workflowVersion,
    addNode,
    removeNode,
    addEdge,
    removeEdge,
    selectNode,
    selectEdge,
    updateNodeData,
    clearCanvas,
    updateEdgeData,
    setDefaultEdge,
    exportSteps,
    exportEvents,
    exportGateways,
    exportEdges,
    loadFromSteps,
  }
})
