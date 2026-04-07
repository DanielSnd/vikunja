import {describe, expect, it} from 'vitest'

import {shouldIgnoreEditorActivation} from './shouldIgnoreEditorActivation'

describe('shouldIgnoreEditorActivation', () => {
	it('ignores clicks inside task list items', () => {
		const wrapper = document.createElement('div')
		wrapper.innerHTML = `
			<li data-type="taskItem" data-checked="false">
				<label><input type="checkbox"><span></span></label>
				<div><p>Checklist item</p></div>
			</li>
		`

		const checkbox = wrapper.querySelector('input')
		const label = wrapper.querySelector('p')

		expect(shouldIgnoreEditorActivation(checkbox)).toBe(true)
		expect(shouldIgnoreEditorActivation(label)).toBe(true)
	})

	it('does not ignore clicks outside task list items', () => {
		const paragraph = document.createElement('p')

		expect(shouldIgnoreEditorActivation(paragraph)).toBe(false)
		expect(shouldIgnoreEditorActivation(null)).toBe(false)
	})
})
