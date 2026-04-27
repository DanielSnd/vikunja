import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IVisionBoardEdge} from '@/modelTypes/IVisionBoardEdge'
import VisionBoardEdgeModel from '@/models/visionBoardEdge'

export default class VisionBoardEdgeService extends AbstractService<IVisionBoardEdge> {
	constructor() {
		super({
			get: '/projects/{projectId}/vision-boards/{boardId}/edges/{id}',
			getAll: '/projects/{projectId}/vision-boards/{boardId}/edges',
			create: '/projects/{projectId}/vision-boards/{boardId}/edges',
			update: '/projects/{projectId}/vision-boards/{boardId}/edges/{id}',
			delete: '/projects/{projectId}/vision-boards/{boardId}/edges/{id}',
		})
	}

	modelFactory(data: Partial<IAbstract>) {
		return new VisionBoardEdgeModel(data)
	}
}
