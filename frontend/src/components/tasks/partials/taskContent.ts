import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'
import type {ITask} from '@/modelTypes/ITask'

function escapeHtml(value: string) {
	return value
		.replaceAll('&', '&amp;')
		.replaceAll('<', '&lt;')
		.replaceAll('>', '&gt;')
		.replaceAll('"', '&quot;')
		.replaceAll('\'', '&#39;')
}

function hasMeaningfulContent(node: ChildNode) {
	if (node.nodeType === Node.TEXT_NODE) {
		return node.textContent?.trim() !== ''
	}

	if (!(node instanceof HTMLElement)) {
		return false
	}

	return (
		(node.textContent?.trim() ?? '') !== ''
		|| node.tagName === 'IMG'
		|| node.querySelector('img') !== null
	)
}

export function buildTaskEditorContent(task: Pick<ITask, 'title' | 'description'>) {
	const title = task.title.trim()
	const description = isEditorContentEmpty(task.description) ? '' : task.description

	if (title === '' && description === '') {
		return ''
	}

	return `<h1><strong>${escapeHtml(title)}</strong></h1>${description}`
}

export function splitTaskEditorContent(value: string) {
	if (isEditorContentEmpty(value)) {
		return {
			title: '',
			description: '',
		}
	}

	const parser = new DOMParser()
	const doc = parser.parseFromString(`<div>${value}</div>`, 'text/html')
	const root = doc.body.firstElementChild

	if (!root) {
		return {
			title: '',
			description: '',
		}
	}

	const nodes = Array.from(root.childNodes).filter(hasMeaningfulContent)

	if (nodes.length === 0) {
		return {
			title: '',
			description: '',
		}
	}

	const [titleNode, ...descriptionNodes] = nodes
	const title = titleNode.textContent?.split('\n')[0]?.trim() ?? ''

	const descriptionWrapper = doc.createElement('div')
	descriptionNodes.forEach(node => {
		descriptionWrapper.appendChild(node.cloneNode(true))
	})

	const description = descriptionWrapper.innerHTML.trim()

	return {
		title,
		description: isEditorContentEmpty(description) ? '' : description,
	}
}
