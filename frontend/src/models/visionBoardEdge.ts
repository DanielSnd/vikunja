import AbstractModel from '@/models/abstractModel'
import type {IVisionBoardEdge} from '@/modelTypes/IVisionBoardEdge'

export default class VisionBoardEdgeModel extends AbstractModel<IVisionBoardEdge> implements IVisionBoardEdge {
	id = 0
	projectId = 0
	boardId = 0
	sourceNodeId = 0
	targetNodeId = 0
	sourceHandle: IVisionBoardEdge['sourceHandle'] = ''
	targetHandle: IVisionBoardEdge['targetHandle'] = ''
	label = ''
	color = ''
	version = 1
	created = new Date()
	updated = new Date()

	constructor(data: Partial<IVisionBoardEdge> = {}) {
		super()
		this.assignData(data)
	}
}
