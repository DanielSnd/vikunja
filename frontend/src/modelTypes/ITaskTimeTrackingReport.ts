export interface ITaskTimeTrackingReportDay {
	date: string
	totalSeconds: number
}

export interface ITaskTimeTrackingReportItem {
	id: number
	title: string
	color: string
	totalSeconds: number
	dailySeconds: number[]
}

export interface ITaskTimeTrackingReport {
	groupBy: 'project' | 'label'
	dateFrom: string
	dateTo: string
	totalSeconds: number
	days: ITaskTimeTrackingReportDay[]
	items: ITaskTimeTrackingReportItem[]
}
