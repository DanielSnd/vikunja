import type {IAbstract} from './IAbstract'

export type VisionBoardNodeKind = 'text' | 'image' | 'video' | 'card' | 'container'

export interface IVisionBoardNode extends IAbstract {
	id: number
	projectId: number
	boardId: number
	parentNodeId: number
	taskId: number
	attachmentId: number
	kind: VisionBoardNodeKind
	title: string
	content: string
	url: string
	x: number
	y: number
	width: number
	height: number
	color: string
	zIndex: number
	version: number
	createdById: number
	updatedById: number
	created: Date
	updated: Date
}
