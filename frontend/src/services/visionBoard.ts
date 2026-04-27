import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IVisionBoard} from '@/modelTypes/IVisionBoard'
import VisionBoardModel from '@/models/visionBoard'

export default class VisionBoardService extends AbstractService<IVisionBoard> {
	constructor() {
		super({
			get: '/projects/{projectId}/vision-boards/{id}',
			getAll: '/projects/{projectId}/vision-boards',
			create: '/projects/{projectId}/vision-boards',
			update: '/projects/{projectId}/vision-boards/{id}',
			delete: '/projects/{projectId}/vision-boards/{id}',
		})
	}

	modelFactory(data: Partial<IAbstract>) {
		return new VisionBoardModel(data)
	}
}
