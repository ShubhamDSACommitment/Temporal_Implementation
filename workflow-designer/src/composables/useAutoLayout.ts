import dagre from 'dagre'
import { useWorkflowStore } from '../stores/workflow'
import { useHistory } from './useHistory'

export function useAutoLayout() {
  const store = useWorkflowStore()
  const { pushHistory } = useHistory()

  function applyLayout() {
    if (store.nodes.length === 0) return

    const g = new dagre.graphlib.Graph()
    g.setDefaultEdgeLabel(() => ({}))
    g.setGraph({ rankdir: 'LR', nodesep: 80, ranksep: 200 })

    for (const node of store.nodes) {
      const isEvent = node.type === 'startEvent' || node.type === 'endEvent' || node.type === 'exclusiveGateway'
      const width = isEvent ? 60 : 200
      const height = isEvent ? 60 : 80
      g.setNode(node.id, { width, height })
    }

    for (const edge of store.edges) {
      g.setEdge(edge.source, edge.target)
    }

    dagre.layout(g)

    for (const node of store.nodes) {
      const pos = g.node(node.id)
      if (pos) {
        const isEvent = node.type === 'startEvent' || node.type === 'endEvent' || node.type === 'exclusiveGateway'
        const width = isEvent ? 60 : 200
        const height = isEvent ? 60 : 80
        node.position = {
          x: pos.x - width / 2,
          y: pos.y - height / 2,
        }
      }
    }

    pushHistory()
  }

  return { applyLayout }
}
