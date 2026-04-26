import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IUser} from './IUser'

export interface ITaskTimeTracking extends IAbstract {
	id: number
	taskId: ITask['id']
	user: IUser
	timeSpent: number
	trackedAt: Date
	isSnoozed: boolean
}

export interface ITaskTimeTrackingSummary {
	user: IUser
	timeSpent: number
}

export interface ITaskTimeTrackingTimer extends IAbstract {
	id: number
	taskId: ITask['id']
	task: ITask | null
	startedAt: Date
	stoppedAt: Date | null
	status: 'running' | 'snoozed'
}
