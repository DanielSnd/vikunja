import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IVisionBoardNode} from '@/modelTypes/IVisionBoardNode'
import VisionBoardNodeModel from '@/models/visionBoardNode'

export default class VisionBoardNodeService extends AbstractService<IVisionBoardNode> {
	constructor() {
		super({
			get: '/projects/{projectId}/vision-boards/{boardId}/nodes/{id}',
			getAll: '/projects/{projectId}/vision-boards/{boardId}/nodes',
			create: '/projects/{projectId}/vision-boards/{boardId}/nodes',
			update: '/projects/{projectId}/vision-boards/{boardId}/nodes/{id}',
			delete: '/projects/{projectId}/vision-boards/{boardId}/nodes/{id}',
		})
	}

	modelFactory(data: Partial<IAbstract>) {
		return new VisionBoardNodeModel(data)
	}
}
