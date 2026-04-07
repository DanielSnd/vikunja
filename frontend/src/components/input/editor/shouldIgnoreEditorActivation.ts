export function shouldIgnoreEditorActivation(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) {
		return false
	}

	return target.closest('[data-type="taskItem"], [data-checked]') !== null
}
