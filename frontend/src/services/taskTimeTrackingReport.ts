import AbstractService from './abstractService'
import type {ITaskTimeTrackingReport} from '@/modelTypes/ITaskTimeTrackingReport'
import {objectToCamelCase} from '@/helpers/case'

export interface TaskTimeTrackingReportParams {
	dateFrom: string
	dateTo: string
	groupBy: 'project' | 'label'
	projectIds?: number[]
	labelIds?: number[]
}

export default class TaskTimeTrackingReportService extends AbstractService<ITaskTimeTrackingReport> {
	constructor() {
		super()
	}

	modelGetFactory(data) {
		const report = objectToCamelCase(data) as ITaskTimeTrackingReport
		report.days = report.days ?? []
		report.items = report.items ?? []
		report.items = report.items.map(item => ({
			...item,
			dailySeconds: item.dailySeconds ?? [],
		}))
		return report
	}

	async getReport(params: TaskTimeTrackingReportParams): Promise<ITaskTimeTrackingReport> {
		return this.getM('/reports/time-tracking', {} as ITaskTimeTrackingReport, {
			dateFrom: params.dateFrom,
			dateTo: params.dateTo,
			groupBy: params.groupBy,
			projectIds: params.projectIds?.join(',') ?? '',
			labelIds: params.labelIds?.join(',') ?? '',
		})
	}
}
