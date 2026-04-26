import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {TaskTimeTrackingTimerModel, TaskTimeTrackingModel} from '@/models/taskTimeTracking'
import type {ITaskTimeTracking, ITaskTimeTrackingTimer} from '@/modelTypes/ITaskTimeTracking'

export default class TaskTimerService {
	async getCurrent(): Promise<ITaskTimeTrackingTimer | null> {
		const response = await AuthenticatedHTTPFactory().get('/tasks/time-tracking/timer')
		return response.data ? new TaskTimeTrackingTimerModel(response.data) : null
	}

	async start(taskId: number): Promise<ITaskTimeTrackingTimer> {
		const response = await AuthenticatedHTTPFactory().put(`/tasks/${taskId}/time-tracking/timer`, {})
		return new TaskTimeTrackingTimerModel(response.data)
	}

	async stop(taskId: number): Promise<ITaskTimeTracking> {
		const response = await AuthenticatedHTTPFactory().post(`/tasks/${taskId}/time-tracking/timer/stop`, {})
		return new TaskTimeTrackingModel(response.data)
	}

	async adjust(deltaSeconds: number): Promise<ITaskTimeTrackingTimer> {
		const response = await AuthenticatedHTTPFactory().post('/tasks/time-tracking/timer/adjust', {
			delta_seconds: deltaSeconds,
		})
		return new TaskTimeTrackingTimerModel(response.data)
	}
}
