import AbstractModel from '@/models/abstractModel'
import VisionBoardEdgeModel from '@/models/visionBoardEdge'
import type {IVisionBoard} from '@/modelTypes/IVisionBoard'
import VisionBoardNodeModel from '@/models/visionBoardNode'

export default class VisionBoardModel extends AbstractModel<IVisionBoard> implements IVisionBoard {
	id = 0
	projectId = 0
	taskId = 0
	title = ''
	taskTitle = ''
	nodes = []
	edges = []
	viewportX = 0
	viewportY = 0
	viewportZoom = 1
	createdById = 0
	created = new Date()
	updated = new Date()

	constructor(data: Partial<IVisionBoard> = {}) {
		super()
		this.assignData(data)
		this.nodes = (data.nodes || []).map(node => new VisionBoardNodeModel(node))
		this.edges = (data.edges || []).map(edge => new VisionBoardEdgeModel(edge))
	}
}
