import AbstractModel from '@/models/abstractModel'
import type {IVisionBoardNode} from '@/modelTypes/IVisionBoardNode'

export default class VisionBoardNodeModel extends AbstractModel<IVisionBoardNode> implements IVisionBoardNode {
	id = 0
	projectId = 0
	boardId = 0
	parentNodeId = 0
	taskId = 0
	attachmentId = 0
	kind: IVisionBoardNode['kind'] = 'text'
	title = ''
	content = ''
	url = ''
	x = 0
	y = 0
	width = 240
	height = 160
	color = ''
	zIndex = 0
	version = 1
	createdById = 0
	updatedById = 0
	created = new Date()
	updated = new Date()

	constructor(data: Partial<IVisionBoardNode> = {}) {
		super()
		this.assignData(data)
	}
}
