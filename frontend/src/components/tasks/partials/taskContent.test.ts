import {describe, expect, it} from 'vitest'

import {buildTaskEditorContent, splitTaskEditorContent} from './taskContent'

describe('taskContent', () => {
	it('preserves standalone images in the description when splitting content', () => {
		const result = splitTaskEditorContent('<h1><strong>Task title</strong></h1><img src="/files/test.png" alt="test">')

		expect(result).toEqual({
			title: 'Task title',
			description: '<img src="/files/test.png" alt="test">',
		})
	})

	it('keeps image-only descriptions when rebuilding and splitting content', () => {
		const content = buildTaskEditorContent({
			title: 'Task title',
			description: '<img src="/files/test.png" alt="test">',
		})

		expect(splitTaskEditorContent(content)).toEqual({
			title: 'Task title',
			description: '<img src="/files/test.png" alt="test">',
		})
	})
})
