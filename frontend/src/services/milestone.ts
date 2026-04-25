import AbstractService from './abstractService'

import MilestoneModel from '@/models/milestone'
import type {IMilestone} from '@/modelTypes/IMilestone'
import {colorFromHex} from '@/helpers/color/colorFromHex'

function parseDate(date: Date | null) {
	if (date) {
		return new Date(date).toISOString()
	}

	return null
}

export default class MilestoneService extends AbstractService<IMilestone> {
	constructor() {
		super({
			create: '/projects/{projectId}/milestones',
			getAll: '/projects/{projectId}/milestones',
			get: '/projects/{projectId}/milestones/{id}',
			update: '/projects/{projectId}/milestones/{id}',
			delete: '/projects/{projectId}/milestones/{id}',
		})
	}

	modelFactory(data) {
		return new MilestoneModel(data)
	}

	modelGetFactory(data) {
		return this.modelFactory(data)
	}

	modelGetAllFactory(data) {
		return this.modelFactory(data)
	}

	beforeCreate(model: IMilestone) {
		return this.processModel(model)
	}

	beforeUpdate(model: IMilestone) {
		return this.processModel(model)
	}

	processModel(model: IMilestone): IMilestone {
		return {
			...model,
			projectId: Number(model.projectId),
			milestoneDate: parseDate(model.milestoneDate),
			hexColor: colorFromHex(model.hexColor),
		} as IMilestone
	}
}
