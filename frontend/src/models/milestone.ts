import type {IMilestone} from '@/modelTypes/IMilestone'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'

import AbstractModel from './abstractModel'
import UserModel from './user'

export default class MilestoneModel extends AbstractModel<IMilestone> implements IMilestone {
	id = 0
	projectId = 0
	name = ''
	milestoneDate: Date | null = null
	hexColor = ''
	users = []
	created: Date = null
	updated: Date = null

	constructor(data: Partial<IMilestone> = {}) {
		super()
		this.assignData(data)

		this.projectId = Number(this.projectId)
		this.milestoneDate = parseDateOrNull(this.milestoneDate)
		this.users = this.users.map(user => new UserModel(user))

		if (this.hexColor !== '' && this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}

		this.created = parseDateOrNull(this.created)
		this.updated = parseDateOrNull(this.updated)
	}
}
