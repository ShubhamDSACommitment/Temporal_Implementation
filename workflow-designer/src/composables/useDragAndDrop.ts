import { useVueFlow } from '@vue-flow/core'
import { useWorkflowStore, type ActivityNodeData } from '../stores/workflow'
import { useHistory } from './useHistory'
import type { NodeKind } from '../types/workflow'

let nodeIdCounter = 0

export function useDragAndDrop() {
  const { screenToFlowCoordinate } = useVueFlow()
  const store = useWorkflowStore()
  const { pushHistory } = useHistory()

  function onDragOver(event: DragEvent) {
    event.preventDefault()
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = 'move'
    }
  }

  function onDrop(event: DragEvent) {
    if (!event.dataTransfer) return

    const kind = (event.dataTransfer.getData('application/node-kind') || 'activity') as NodeKind
    const displayName = event.dataTransfer.getData('application/display-name')

    if (!displayName) return

    const position = screenToFlowCoordinate({
      x: event.clientX,
      y: event.clientY,
    })

    if (kind === 'startEvent') {
      const id = `${kind}-${++nodeIdCounter}`
      store.addNode({
        id,
        type: kind,
        kind,
        position,
        data: { label: displayName, startConfig: { triggerType: 'manual', formFields: [] } },
      })
      pushHistory()
      return
    }

    if (kind === 'endEvent') {
      const id = `${kind}-${++nodeIdCounter}`
      store.addNode({
        id,
        type: kind,
        kind,
        position,
        data: { label: displayName, endConfig: { endType: 'none', outputVariables: [] } },
      })
      pushHistory()
      return
    }

    if (kind === 'exclusiveGateway') {
      const id = `${kind}-${++nodeIdCounter}`
      store.addNode({
        id,
        type: kind,
        kind,
        position,
        data: { label: displayName },
      })
      pushHistory()
      return
    }

    // Activity node
    const activityName = event.dataTransfer.getData('application/activity-name')
    const taskQueue = event.dataTransfer.getData('application/task-queue')
    const inputMappingRaw = event.dataTransfer.getData('application/input-mapping')

    if (!activityName) return

    const id = `step-${activityName}-${++nodeIdCounter}`

    let inputMapping: Record<string, string> = {}
    if (inputMappingRaw) {
      try { inputMapping = JSON.parse(inputMappingRaw) } catch { /* empty */ }
    }

    const data: ActivityNodeData = {
      label: displayName,
      activityName,
      taskQueue,
      timeoutSeconds: 30,
      retryPolicy: {
        maxAttempts: 3,
        initialIntervalSec: 1,
        backoffCoefficient: 2.0,
      },
      inputMapping,
    }

    store.addNode({
      id,
      type: 'activity',
      kind: 'activity',
      position,
      data,
    })
    pushHistory()
  }

  return { onDragOver, onDrop }
}
