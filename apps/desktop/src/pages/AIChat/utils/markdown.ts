import { marked } from 'marked'
import hljs from 'highlight.js'

const renderer = new marked.Renderer()

renderer.code = ({ text, lang }) => {
  let highlighted: string
  if (lang && hljs.getLanguage(lang)) {
    try {
      highlighted = hljs.highlight(text, { language: lang }).value
    } catch {
      highlighted = escapeHtml(text)
    }
  } else {
    highlighted = escapeHtml(text)
  }
  const langClass = lang ? ` class="language-${escapeHtml(lang)}"` : ''
  return `<pre><code${langClass}>${highlighted}</code></pre>`
}

marked.setOptions({
  gfm: true,
  breaks: true,
})

export function renderMarkdown(text: string): string {
  if (!text) return ''
  return marked.parse(text, { renderer }) as string
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}

export function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard) {
    return navigator.clipboard.writeText(text)
  }
  const textarea = document.createElement('textarea')
  textarea.value = text
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  textarea.select()
  document.execCommand('copy')
  document.body.removeChild(textarea)
  return Promise.resolve()
}
