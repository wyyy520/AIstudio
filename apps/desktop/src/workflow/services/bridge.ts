import { ref, watch, onUnmounted } from 'vue'
import type { Node, Edge, Connection } from '@vue-flow/core'
import type { WorkflowNodeSchema, WorkflowEdgeSchema } from '../schema'
import { useWorkflowStore } from '../store'
import { deserializeNodesFromVueFlow, deserializeEdgesFromVueFlow, serializeNodesToVueFlow, serializeEdgesToVueFlow } from '../serializer'
import { isPortTypeCompatible } from '../edge/registry'

export function useWorkflowBridge() {
  const store = useWorkflowStore()

  const vueFlowNodes = ref<Node[]>([])
  const vueFlowEdges = ref<Edge[]>([])

  let updating = false

  function syncStoreToVueFlow() {
    updating = true
    const sn = serializeNodesToVueFlow(store.workflow.nodes)
    const se = serializeEdgesToVueFlow(store.workflow.edges)
    vueFlowNodes.value = sn as Node[]
    vueFlowEdges.value = se as Edge[]
    updating = false
  }

  function syncVueFlowToStore() {
    if (updating) return
    updating = true
    const nodes: WorkflowNodeSchema[] = []
    // @ts-expect-error - deep type inference
    const deNodes = deserializeNodesFromVueFlow(vueFlowNodes.value)
    for (const n of deNodes) { nodes.push(n as WorkflowNodeSchema) }
    const edges: WorkflowEdgeSchema[] = []
    const deEdges = deserializeEdgesFromVueFlow(vueFlowEdges.value)
    for (const e of deEdges) { edges.push(e as WorkflowEdgeSchema) }

    nodes.forEach(n => {
      const existing = store.workflow.nodes.find(en => en.id === n.id)
      if (existing) {
        existing.position = n.position
        existing.updated_at = new Date().toISOString()
      }
    })

    const removedNodes = store.workflow.nodes.filter(en => !nodes.find(n => n.id === en.id))
    removedNodes.forEach(rn => store.removeNode(rn.id))

    nodes.forEach(n => {
      if (!store.workflow.nodes.find(en => en.id === n.id)) {
        store.workflow.nodes.push(n)
      }
    })

    store.workflow.edges = edges
    store.workflow.updated_at = new Date().toISOString()
    store.isDirty = true
    updating = false
  }

  const stopWatch = watch(
    () => ({
      nodes: store.workflow.nodes.map(n => ({ id: n.id, x: n.position.x, y: n.position.y, config: n.config })),
      edges: store.workflow.edges.map(e => ({ id: e.id, source: e.source_node, target: e.target_node })),
    }),
    () => {
      syncStoreToVueFlow()
    },
    { deep: true },
  )

  syncStoreToVueFlow()

  onUnmounted(() => {
    stopWatch()
  })

  function onNodeDragStop(node: Node) {
    const x = node.position.x ?? 0
    const y = node.position.y ?? 0
    store.updateNodePosition(node.id, x, y)
  }

  function onConnect(connection: Connection) {
    if (!connection.source || !connection.target) return

    const sourceNode = store.workflow.nodes.find(n => n.id === connection.source)
    const targetNode = store.workflow.nodes.find(n => n.id === connection.target)
    if (!sourceNode || !targetNode) return

    const sourcePort = sourceNode.outputs.find(p => p.id === (connection.sourceHandle || ''))
    const targetPort = targetNode.inputs.find(p => p.id === (connection.targetHandle || ''))

    if (sourcePort && targetPort && !isPortTypeCompatible(sourcePort.type, targetPort.type)) {
      console.warn(`[bridge] incompatible port types: ${sourcePort.type} -> ${targetPort.type}`)
      return
    }

    const existing = store.workflow.edges.find(
      e => e.source_node === connection.source && e.target_node === connection.target
    )
    if (existing) return

    store.addEdge(
      connection.source,
      connection.sourceHandle || '',
      connection.target,
      connection.targetHandle || '',
    )
  }

  function onEdgeDoubleClick(edge: Edge) {
    store.removeEdge(edge.id)
  }

  function onNodeClick(node: Node) {
    store.selectNode(node.id)
  }

  function onEdgeClick(edge: Edge) {
    store.selectEdge(edge.id)
  }

  function onPaneClick() {
    store.selectNode(null)
    store.selectEdge(null)
  }

  return {
    vueFlowNodes,
    vueFlowEdges,
    onNodeDragStop,
    onConnect,
    onEdgeDoubleClick,
    onNodeClick,
    onEdgeClick,
    onPaneClick,
    syncStoreToVueFlow,
  }
}
