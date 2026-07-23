<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="dialog-overlay" @click.self="cancel">
        <div class="dialog-container">
          <div class="dialog-header">
            <h2 class="dialog-title">New Project</h2>
            <button class="dialog-close" @click="cancel">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M18 6L6 18M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="dialog-content">
            <div class="form-group">
              <label class="form-label">Project Name</label>
              <input
                v-model="form.name"
                type="text"
                class="form-input"
                placeholder="输入项目名称"
              />
            </div>

            <div class="form-group">
              <label class="form-label">Template</label>
              <div class="template-grid">
                <div
                  v-for="template in templates"
                  :key="template.id"
                  class="template-card"
                  :class="{ selected: form.template === template.id }"
                  @click="selectTemplate(template)"
                >
                  <div class="template-name">{{ template.name }}</div>
                  <div class="template-desc">{{ template.description }}</div>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">Framework</label>
              <div class="framework-options">
                <label
                  v-for="fw in frameworks"
                  :key="fw.value"
                  class="framework-option"
                  :class="{ selected: form.framework === fw.value }"
                >
                  <input
                    v-model="form.framework"
                    type="radio"
                    :value="fw.value"
                    class="framework-radio"
                  />
                  <span class="framework-label">{{ fw.label }}</span>
                </label>
              </div>
            </div>
          </div>

          <div class="dialog-footer">
            <button class="dialog-btn dialog-btn-cancel" @click="cancel">取消</button>
            <button
              class="dialog-btn dialog-btn-create"
              :disabled="!canCreate"
              @click="create"
            >
              Create Project
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import type { ProjectTemplate } from '@/types/project'

interface Props {
  visible: boolean
  templates?: ProjectTemplate[]
}

const props = withDefaults(defineProps<Props>(), {
  templates: () => [
    { id: 'python', name: 'Python AI', description: 'AI training, inference, data processing', type: 'custom', framework: 'pytorch', plugins: [] },
    { id: 'custom', name: 'Custom', description: 'Empty project, build from scratch', type: 'custom', framework: 'pytorch', plugins: [] },
  ],
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  create: [data: { name: string; description?: string; target?: string }]
}>()

const form = reactive({
  name: '',
  template: 'tpl-empty',
  framework: 'pytorch',
})

const frameworks = [
  { value: 'pytorch', label: 'PyTorch' },
  { value: 'tensorflow', label: 'TensorFlow' },
  { value: 'onnx', label: 'ONNX' },
  { value: 'tensorrt', label: 'TensorRT' },
  { value: 'auto', label: 'Auto' },
]

const selectedTemplate = computed(() => {
  return props.templates.find(t => t.id === form.template)
})

const canCreate = computed(() => {
  return form.name.trim().length > 0 && form.template
})

function selectTemplate(template: ProjectTemplate) {
  form.template = template.id
  if (template.framework !== 'auto') {
    form.framework = template.framework
  }
}

function cancel() {
  emit('update:visible', false)
}

function create() {
  if (!canCreate.value) return

  const template = selectedTemplate.value
  emit('create', {
    name: form.name,
    description: `Template: ${template?.name || 'Custom'}`,
    target: form.framework,
  })

  emit('update:visible', false)

  form.name = ''
  form.template = 'tpl-empty'
  form.framework = 'pytorch'
}
</script>

<style scoped>
.dialog-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  z-index: 1000;
}

.dialog-container {
  width: 520px;
  max-height: 80vh;
  background: var(--bg-secondary);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-4) var(--spacing-6);
  border-bottom: 1px solid var(--border-subtle);
}

.dialog-title {
  font-size: var(--text-h3);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
}

.dialog-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dialog-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.dialog-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-6);
}

.form-group {
  margin-bottom: var(--spacing-5);
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-label {
  display: block;
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-secondary);
  margin-bottom: var(--spacing-2);
}

.form-input {
  width: 100%;
  height: 40px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: var(--text-body);
}

.form-input:focus {
  border-color: var(--primary);
  outline: none;
}

.form-input::placeholder {
  color: var(--text-tertiary);
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--spacing-2);
}

.template-card {
  padding: var(--spacing-3);
  background: var(--bg-tertiary);
  border: 2px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.template-card:hover {
  border-color: var(--border-strong);
}

.template-card.selected {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.template-name {
  font-size: var(--text-body-sm);
  font-weight: var(--font-semibold);
  color: var(--text-primary);
  margin-bottom: var(--spacing-1);
}

.template-desc {
  font-size: var(--text-caption);
  color: var(--text-tertiary);
  line-height: var(--leading-caption);
}

.framework-options {
  display: flex;
  gap: var(--spacing-2);
  flex-wrap: wrap;
}

.framework-option {
  display: flex;
  align-items: center;
  gap: var(--spacing-2);
  height: 36px;
  padding: 0 var(--spacing-3);
  background: var(--bg-tertiary);
  border: 2px solid transparent;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.framework-option:hover {
  border-color: var(--border-strong);
}

.framework-option.selected {
  border-color: var(--primary);
  background: var(--primary-bg);
}

.framework-radio {
  display: none;
}

.framework-label {
  font-size: var(--text-body-sm);
  color: var(--text-primary);
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-3);
  padding: var(--spacing-4) var(--spacing-6);
  border-top: 1px solid var(--border-subtle);
}

.dialog-btn {
  height: 40px;
  padding: 0 var(--spacing-5);
  border-radius: var(--radius-md);
  font-size: var(--text-body-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.dialog-btn-cancel {
  background: transparent;
  border: 1px solid var(--border-default);
  color: var(--text-secondary);
}

.dialog-btn-cancel:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.dialog-btn-create {
  background: var(--primary);
  border: none;
  color: white;
}

.dialog-btn-create:hover:not(:disabled) {
  background: var(--primary-hover);
}

.dialog-btn-create:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .dialog-container,
.modal-leave-active .dialog-container {
  transition: transform 0.2s ease;
}

.modal-enter-from .dialog-container,
.modal-leave-to .dialog-container {
  transform: scale(0.95) translateY(-10px);
}
</style>
