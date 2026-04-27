import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IVisionBoardEdge} from './IVisionBoardEdge'
import type {IVisionBoardNode} from './IVisionBoardNode'

export interface IVisionBoard extends IAbstract {
	id: number
	projectId: number
	taskId: ITask['id']
	title: string
	taskTitle: string
	nodes?: IVisionBoardNode[]
	edges?: IVisionBoardEdge[]
	viewportX: number
	viewportY: number
	viewportZoom: number
	createdById: number
	created: Date
	updated: Date
}
