export const PRIORITIES = {
	'UNSET': 0,
	'LOW': 1,
	'MEDIUM': 2,
	'HIGH': 3,
	'URGENT': 4,
	'DO_NOW': 5,
} as const

export type Priority = typeof PRIORITIES[keyof typeof PRIORITIES]

export const STATUSES = {
	'UNSET': 0,
	'IN_PROGRESS': 1,
	'BLOCKED': 2,
	'REVIEW': 3,
	'DONE': 4,
} as const

export type Status = typeof STATUSES[keyof typeof STATUSES]