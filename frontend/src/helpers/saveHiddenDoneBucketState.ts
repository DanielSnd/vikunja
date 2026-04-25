import type {IBucket} from '@/modelTypes/IBucket'
import type {IProject} from '@/modelTypes/IProject'
import type {IProjectView} from '@/modelTypes/IProjectView'

const key = 'hiddenDoneBuckets'

export type HiddenDoneBuckets = {[id: IBucket['id']]: boolean}

function getAllState() {
	const saved = localStorage.getItem(key)
	return saved === null
		? {}
		: JSON.parse(saved)
}

function getStateKey(projectId: IProject['id'], viewId: IProjectView['id']) {
	return `${projectId}-${viewId}`
}

export const saveHiddenDoneBucketState = (
	projectId: IProject['id'],
	viewId: IProjectView['id'],
	hiddenDoneBuckets: HiddenDoneBuckets,
) => {
	const state = getAllState()
	const stateKey = getStateKey(projectId, viewId)
	state[stateKey] = hiddenDoneBuckets

	for (const bucketId in state[stateKey]) {
		if (!state[stateKey][bucketId]) {
			delete state[stateKey][bucketId]
		}
	}

	localStorage.setItem(key, JSON.stringify(state))
}

export function getHiddenDoneBucketState(projectId: IProject['id'], viewId: IProjectView['id']) {
	const state = getAllState()
	const stateKey = getStateKey(projectId, viewId)

	return typeof state[stateKey] !== 'undefined'
		? state[stateKey]
		: {}
}
