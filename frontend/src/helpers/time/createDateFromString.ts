/**
 * Returns a new date from any format in a way that all browsers, especially safari, can understand.
 *
 * @see https://kolaente.dev/vikunja/frontend/issues/207
 *
 * @param dateString
 * @returns {Date}
 */
export function createDateFromString(dateString: string | number | Date) {
	if (dateString instanceof Date) {
		return dateString
	}

	// Keep RFC3339/ISO timestamps intact and only normalize legacy
	// `YYYY-MM-DD HH:mm[:ss]` strings for Safari.
	if (
		typeof dateString === 'string' &&
		dateString.includes('-') &&
		!dateString.includes('T')
	) {
		dateString = dateString.replace(/-/g, '/')
	}

	return new Date(dateString)
}
