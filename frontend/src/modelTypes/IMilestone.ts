import type {IAbstract} from './IAbstract'
import type {IProject} from './IProject'
import type {IUser} from './IUser'

export interface IMilestone extends IAbstract {
	id: number
	projectId: IProject['id']
	name: string
	milestoneDate: Date | null
	hexColor: string
	users: IUser[]
	created: Date
	updated: Date
}
