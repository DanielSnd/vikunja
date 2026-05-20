import AbstractService from '@/services/abstractService'
import type {IUserReportApplication, IUserReportToken} from '@/modelTypes/IUserReport'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'

export interface IUserReportApplicationCreateResponse {
	application: IUserReportApplication
	accessKey: string
	defaultReportToken: IUserReportToken
}

export default class UserReportService extends AbstractService<IUserReportApplication> {
	constructor() {
		super({
			create: '/user-reports/applications',
			getAll: '/user-reports/applications',
			update: '/user-reports/applications/{id}',
			delete: '/user-reports/applications/{id}',
		})
	}

	processApplication(model: IUserReportApplication): IUserReportApplication {
		const parsedModel = objectToCamelCase(model) as IUserReportApplication

		return {
			...parsedModel,
			created: new Date(parsedModel.created).toISOString(),
			updated: new Date(parsedModel.updated).toISOString(),
		}
	}

	modelFactory(data: Partial<IUserReportApplication>) {
		return this.processApplication(data as IUserReportApplication)
	}

	async createApplication(data: {name: string, projectId: number}): Promise<IUserReportApplicationCreateResponse> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.put('/user-reports/applications', objectToSnakeCase(data))
			return {
				...response.data,
				application: this.processApplication(response.data.application),
				defaultReportToken: this.processToken(response.data.defaultReportToken),
			}
		} finally {
			cancel()
		}
	}

	async updateApplication(application: IUserReportApplication): Promise<IUserReportApplication> {
		const cancel = this.setLoading()

		try {
			const payload = objectToSnakeCase({
				name: application.name,
				projectId: normalizeInteger(application.projectId),
				maxUploadSize: normalizeInteger(application.maxUploadSize),
				bucketId: normalizeInteger(application.bucketId),
				criticalProjectId: normalizeInteger(application.criticalProjectId),
				highProjectId: normalizeInteger(application.highProjectId),
				lowProjectId: normalizeInteger(application.lowProjectId),
				criticalBucketId: normalizeInteger(application.criticalBucketId),
				highBucketId: normalizeInteger(application.highBucketId),
				lowBucketId: normalizeInteger(application.lowBucketId),
				criticalPriority: normalizeInteger(application.criticalPriority),
				highPriority: normalizeInteger(application.highPriority),
				lowPriority: normalizeInteger(application.lowPriority),
			})
			const response = await this.http.post(
				`/user-reports/applications/${application.id}`,
				payload,
			)
			return this.processApplication(response.data)
		} finally {
			cancel()
		}
	}

	async deleteApplication(applicationId: number): Promise<void> {
		const cancel = this.setLoading()

		try {
			await this.http.delete(`/user-reports/applications/${applicationId}`)
		} finally {
			cancel()
		}
	}

	async regenerateAccessKey(applicationId: number): Promise<string> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.post(`/user-reports/applications/${applicationId}/access-key/regenerate`)
			return response.data.accessKey
		} finally {
			cancel()
		}
	}

	async getTokens(applicationId: number): Promise<IUserReportToken[]> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get(`/user-reports/applications/${applicationId}/tokens`)
			return response.data.map(token => this.processToken(token))
		} finally {
			cancel()
		}
	}

	async createToken(applicationId: number, label: string): Promise<IUserReportToken> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.put(
				`/user-reports/applications/${applicationId}/tokens`,
				objectToSnakeCase({label}),
			)
			return this.processToken(response.data)
		} finally {
			cancel()
		}
	}

	async updateToken(applicationId: number, token: IUserReportToken): Promise<IUserReportToken> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.post(
				`/user-reports/applications/${applicationId}/tokens/${token.id}`,
				objectToSnakeCase({
					isEnabled: token.isEnabled,
				}),
			)
			return this.processToken(response.data)
		} finally {
			cancel()
		}
	}

	processToken(token: IUserReportToken): IUserReportToken {
		const parsedToken = objectToCamelCase(token) as IUserReportToken

		return {
			...parsedToken,
			created: new Date(parsedToken.created).toISOString(),
		}
	}
}

function normalizeInteger(value: unknown): number {
	const parsed = Number(value)
	return Number.isFinite(parsed) ? Math.trunc(parsed) : 0
}
