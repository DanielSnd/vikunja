export function formatDuration(seconds: number): string {
	const totalSeconds = Math.max(0, Math.round(seconds))
	const hours = Math.floor(totalSeconds / 3600)
	const minutes = Math.floor((totalSeconds % 3600) / 60)

	if (hours > 0) {
		return `${hours}h ${minutes}m`
	}

	return `${minutes}m`
}

export function formatDurationCompact(seconds: number): string {
	const totalSeconds = Math.max(0, Math.round(seconds))
	const hours = Math.floor(totalSeconds / 3600)
	const minutes = Math.floor((totalSeconds % 3600) / 60)
	const remainingSeconds = totalSeconds % 60

	if (hours > 0) {
		return `${hours}:${String(minutes).padStart(2, '0')}:${String(remainingSeconds).padStart(2, '0')}`
	}

	if (minutes > 0) {
		return `${minutes}:${String(remainingSeconds).padStart(2, '0')}`
	}

	return `${remainingSeconds}s`
}
