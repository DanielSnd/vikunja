import {test, expect} from 'vitest'

import {createDateFromString} from './createDateFromString'

test('YYYY-MM-DD HH:MM', () => {
	const dateString = '2021-02-06 12:00'
	const date = createDateFromString(dateString)
	expect(date).toBeInstanceOf(Date)
	expect(date.getDate()).toBe(6)
	expect(date.getMonth()).toBe(1)
	expect(date.getFullYear()).toBe(2021)
	expect(date.getHours()).toBe(12)
	expect(date.getMinutes()).toBe(0)
	expect(date.getSeconds()).toBe(0)
})

test('RFC3339 timestamp', () => {
	const dateString = '2026-05-20T14:30:00Z'
	const date = createDateFromString(dateString)

	expect(date).toBeInstanceOf(Date)
	expect(Number.isNaN(date.getTime())).toBe(false)
	expect(date.getTime()).toBe(new Date(dateString).getTime())
})

test('timestamp number', () => {
	const timestamp = 1747749600000
	const date = createDateFromString(timestamp)

	expect(date).toBeInstanceOf(Date)
	expect(Number.isNaN(date.getTime())).toBe(false)
	expect(date.getTime()).toBe(timestamp)
})
