import type {IAbstract} from './IAbstract'

export interface IVisionBoardEdge extends IAbstract {
	id: number
	projectId: number
	boardId: number
	sourceNodeId: number
	targetNodeId: number
	sourceHandle: 'top' | 'left' | 'right' | 'bottom' | ''
	targetHandle: 'top' | 'left' | 'right' | 'bottom' | ''
	label: string
	color: string
	version: number
	created: Date
	updated: Date
}
