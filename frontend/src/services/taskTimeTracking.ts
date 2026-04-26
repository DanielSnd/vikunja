import AbstractService from './abstractService'
import {TaskTimeTrackingModel} from '@/models/taskTimeTracking'
import type {ITaskTimeTracking} from '@/modelTypes/ITaskTimeTracking'
import {objectToSnakeCase} from '@/helpers/case'

export default class TaskTimeTrackingService extends AbstractService<ITaskTimeTracking> {
	constructor() {
		super({
			create: '/tasks/{taskId}/time-tracking',
			getAll: '/tasks/{taskId}/time-tracking',
			get: '/tasks/{taskId}/time-tracking/{id}',
			update: '/tasks/{taskId}/time-tracking/{id}',
			delete: '/tasks/{taskId}/time-tracking/{id}',
		})
	}

	modelFactory(data) {
		return new TaskTimeTrackingModel(data)
	}

	autoTransformBeforePost(): boolean {
		return false
	}

	beforeCreate(model: ITaskTimeTracking) {
		return this.transformModel(model)
	}

	beforeUpdate(model: ITaskTimeTracking) {
		return this.transformModel(model)
	}

	private transformModel(model: ITaskTimeTracking) {
		return objectToSnakeCase({
			...model,
			trackedAt: model.trackedAt ? new Date(model.trackedAt).toISOString() : null,
		}) as ITaskTimeTracking
	}
}
