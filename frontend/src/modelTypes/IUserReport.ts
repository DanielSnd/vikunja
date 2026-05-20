import type {IAbstract} from '@/modelTypes/IAbstract'

export interface IUserReportApplication extends IAbstract {
	id: number
	name: string
	projectId: number
	maxUploadSize: number
	bucketId: number
	criticalProjectId: number
	highProjectId: number
	lowProjectId: number
	criticalBucketId: number
	highBucketId: number
	lowBucketId: number
	criticalPriority: number
	highPriority: number
	lowPriority: number
	created: string
	updated: string
}

export interface IUserReportToken extends IAbstract {
	id: number
	applicationId: number
	label: string
	token?: string
	isEnabled: boolean
	created: string
}
