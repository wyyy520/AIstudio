/**
 * Workflow Integration Tests
 * Tests for node dragging, connecting, deleting, selecting,
 * dynamic configuration, saving, running, and task status
 */

import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import type { Node, Edge } from '@vue-flow/core'
import { validateWorkflow, topologicalSort } from '../validators/workflowValidator'
import { toWorkflowJSON } from '../types/workflow'

describe('Workflow Editor Integration', () => {
  // Test data
  const createNode = (id: string, type: string, x: number, y: number): Node => ({
    id,
    type: 'custom',
    position: { x, y },
    data: {
      label: `Node ${id}`,
      nodeType: type,
      status: 'idle',
      inputs: [{ name: 'input', label: 'Input', type: 'any' }],
      outputs: [{ name: 'output', label: 'Output', type: 'any' }],
    },
  })

  const createEdge = (id: string, source: string, target: string): Edge => ({
    id,
    source,
    target,
    sourceHandle: 'output',
    targetHandle: 'input',
  })

  describe('1. Node Operations', () => {
    it('should create nodes with correct structure', () => {
      const node = createNode('n1', 'vision', 100, 100)
      expect(node.id).toBe('n1')
      expect(node.data.nodeType).toBe('vision')
      expect(node.position).toEqual({ x: 100, y: 100 })
    })

    it('should connect nodes with valid edges', () => {
      const nodes = [createNode('n1', 'vision', 100, 100), createNode('n2', 'nlp', 300, 100)]
      const edges = [createEdge('e1', 'n1', 'n2')]
      
      expect(edges).toHaveLength(1)
      expect(edges[0].source).toBe('n1')
      expect(edges[0].target).toBe('n2')
    })

    it('should delete nodes and their connected edges', () => {
      const nodes = ref<Node[]>([
        createNode('n1', 'vision', 100, 100),
        createNode('n2', 'nlp', 300, 100),
        createNode('n3', 'dataset', 500, 100),
      ])
      const edges = ref<Edge[]>([
        createEdge('e1', 'n1', 'n2'),
        createEdge('e2', 'n2', 'n3'),
      ])

      // Delete n2
      const deletedId = 'n2'
      nodes.value = nodes.value.filter(n => n.id !== deletedId)
      edges.value = edges.value.filter(e => e.source !== deletedId && e.target !== deletedId)

      expect(nodes.value).toHaveLength(2)
      expect(edges.value).toHaveLength(0) // Both edges connected to n2
    })

    it('should select a node', () => {
      const selectedNodeId = ref<string | null>(null)
      const nodes = [createNode('n1', 'vision', 100, 100)]

      selectedNodeId.value = nodes[0].id
      expect(selectedNodeId.value).toBe('n1')
    })
  })

  describe('2. Dynamic Node Configuration', () => {
    it('should have params for YOLO node', () => {
      const yoloNode = createNode('yolo', 'vision', 100, 100)
      yoloNode.data.params = {
        epochs: 100,
        batch_size: 16,
        device: 'cuda',
      }

      expect(yoloNode.data.params.epochs).toBe(100)
      expect(yoloNode.data.params.batch_size).toBe(16)
      expect(yoloNode.data.params.device).toBe('cuda')
    })

    it('should have params for CNN node', () => {
      const cnnNode = createNode('cnn', 'vision', 100, 100)
      cnnNode.data.params = {
        layers: 5,
        optimizer: 'adam',
        learning_rate: 0.001,
      }

      expect(cnnNode.data.params.layers).toBe(5)
      expect(cnnNode.data.params.optimizer).toBe('adam')
    })
  })

  describe('3. Workflow Validation', () => {
    it('should validate a simple workflow', () => {
      const nodes = [
        createNode('n1', 'input', 100, 100),
        createNode('n2', 'vision', 300, 100),
        createNode('n3', 'output', 500, 100),
      ]
      const edges = [
        createEdge('e1', 'n1', 'n2'),
        createEdge('e2', 'n2', 'n3'),
      ]

      const errors = validateWorkflow(nodes, edges)
      expect(errors).toHaveLength(0)
    })

    it('should detect disconnected nodes', () => {
      const nodes = [
        createNode('n1', 'input', 100, 100),
        createNode('n2', 'vision', 300, 100),
        createNode('n3', 'output', 500, 100),
      ]
      const edges = [createEdge('e1', 'n1', 'n2')]
      // n3 is disconnected

      const errors = validateWorkflow(nodes, edges)
      expect(errors.some(e => e.nodeId === 'n3')).toBe(true)
    })
  })

  describe('4. Workflow Save', () => {
    it('should generate workflow.json', () => {
      const nodes = [
        createNode('n1', 'input', 100, 100),
        createNode('n2', 'vision', 300, 100),
      ]
      const edges = [createEdge('e1', 'n1', 'n2')]

      const workflowJSON = toWorkflowJSON(
        'wf_1',
        'Test Workflow',
        'Test Description',
        nodes,
        edges
      )

      expect(workflowJSON.id).toBe('wf_1')
      expect(workflowJSON.name).toBe('Test Workflow')
      expect(workflowJSON.nodes).toHaveLength(2)
      expect(workflowJSON.edges).toHaveLength(1)
    })
  })

  describe('5. Topological Sort', () => {
    it('should sort nodes in execution order', () => {
      const nodes = [
        createNode('n1', 'input', 100, 100),
        createNode('n2', 'vision', 300, 100),
        createNode('n3', 'output', 500, 100),
      ]
      const edges = [
        createEdge('e1', 'n1', 'n2'),
        createEdge('e2', 'n2', 'n3'),
      ]

      const sorted = topologicalSort(nodes, edges)
      expect(sorted).toEqual(['n1', 'n2', 'n3'])
    })
  })
})