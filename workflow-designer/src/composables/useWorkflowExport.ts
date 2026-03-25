import { useWorkflowStore } from '../stores/workflow'
import type { WorkflowDefinition } from '../types/workflow'

export function useWorkflowExport() {
  const store = useWorkflowStore()

  function exportDefinition(): WorkflowDefinition {
    return {
      id: store.workflowId || '',
      name: store.workflowName,
      description: store.workflowDescription,
      version: store.workflowVersion,
      steps: store.exportSteps(),
      events: store.exportEvents(),
      gateways: store.exportGateways(),
      edges: store.exportEdges(),
    }
  }

  return { exportDefinition }
}
