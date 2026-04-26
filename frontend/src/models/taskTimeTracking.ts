import AbstractModel from './abstractModel'
import UserModel from './user'
import type {ITaskTimeTracking, ITaskTimeTrackingSummary, ITaskTimeTrackingTimer} from '@/modelTypes/ITaskTimeTracking'
import type {ITask} from '@/modelTypes/ITask'
import TaskModel from './task'

export class TaskTimeTrackingModel extends AbstractModel<ITaskTimeTracking> implements ITaskTimeTracking {
	id = 0
	taskId: ITask['id'] = 0
	user = new UserModel()
	timeSpent = 0
	trackedAt: Date = null
	isSnoozed = false

	constructor(data: Partial<ITaskTimeTracking> = {}) {
		super()
		this.assignData(data)

		this.user = new UserModel(this.user)
		this.trackedAt = new Date(this.trackedAt)
	}
}

export class TaskTimeTrackingTimerModel extends AbstractModel<ITaskTimeTrackingTimer> implements ITaskTimeTrackingTimer {
	id = 0
	taskId: ITask['id'] = 0
	task: ITask | null = null
	startedAt: Date = null
	stoppedAt: Date | null = null
	status: 'running' | 'snoozed' = 'running'

	constructor(data: Partial<ITaskTimeTrackingTimer> = {}) {
		super()
		this.assignData(data)

		this.task = this.task ? new TaskModel(this.task) : null
		this.startedAt = new Date(this.startedAt)
		this.stoppedAt = this.stoppedAt ? new Date(this.stoppedAt) : null
	}
}

export function normalizeTaskTimeTrackingSummary(summary: Partial<ITaskTimeTrackingSummary>[] = []): ITaskTimeTrackingSummary[] {
	return summary.map(item => ({
		user: new UserModel(item.user),
		timeSpent: Number(item.timeSpent ?? 0),
	}))
}
