<script setup lang="ts">
import {computed, ref, shallowReactive, watchEffect} from 'vue'
import {useRoute} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {useTitle} from '@vueuse/core'

import type {IProject} from '@/modelTypes/IProject'
import type {IMilestone} from '@/modelTypes/IMilestone'
import type {IUser} from '@/modelTypes/IUser'

import CreateEdit from '@/components/misc/CreateEdit.vue'
import FormField from '@/components/input/FormField.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import BaseButton from '@/components/base/BaseButton.vue'
import User from '@/components/misc/User.vue'

import ProjectService from '@/services/project'
import MilestoneService from '@/services/milestone'
import ProjectUserService from '@/services/projectUsers'
import ProjectModel from '@/models/project'
import MilestoneModel from '@/models/milestone'
import UserModel from '@/models/user'
import {useBaseStore} from '@/stores/base'
import {success} from '@/message'

defineOptions({name: 'ProjectSettingsMilestones'})

const {t} = useI18n({useScope: 'global'})
useTitle(t('project.milestones.title'))

const route = useRoute()
const projectId = computed(() => Number(route.params.projectId))

const project = ref<IProject>()
const milestones = ref<IMilestone[]>([])
const milestone = ref<IMilestone>(new MilestoneModel())
const foundUsers = ref<IUser[]>([])
const loading = ref(false)
const saving = ref(false)
const editingId = ref<IMilestone['id'] | null>(null)
const formVersion = ref(0)
const milestoneDateInput = ref('')
const milestoneDateDateInput = ref('')
const milestoneDateTimeInput = ref('')

const milestoneService = shallowReactive(new MilestoneService())
const projectUserService = shallowReactive(new ProjectUserService())

function formatMilestoneDateInput(date: Date | string | null): string {
	if (!date) {
		return ''
	}

	const parsedDate = typeof date === 'string'
		? new Date(date)
		: date

	if (Number.isNaN(parsedDate.getTime())) {
		return ''
	}

	const year = parsedDate.getFullYear()
	const month = String(parsedDate.getMonth() + 1).padStart(2, '0')
	const day = String(parsedDate.getDate()).padStart(2, '0')
	const hours = String(parsedDate.getHours()).padStart(2, '0')
	const minutes = String(parsedDate.getMinutes()).padStart(2, '0')

	return `${year}-${month}-${day}T${hours}:${minutes}`
}

function formatMilestoneDateParts(date: Date | string | null) {
	if (!date) {
		return {
			date: '',
			time: '',
		}
	}

	const parsedDate = typeof date === 'string'
		? new Date(date)
		: date

	if (Number.isNaN(parsedDate.getTime())) {
		return {
			date: '',
			time: '',
		}
	}

	const year = parsedDate.getFullYear()
	const month = String(parsedDate.getMonth() + 1).padStart(2, '0')
	const day = String(parsedDate.getDate()).padStart(2, '0')
	const hours = String(parsedDate.getHours()).padStart(2, '0')
	const minutes = String(parsedDate.getMinutes()).padStart(2, '0')

	return {
		date: `${year}-${month}-${day}`,
		time: `${hours}:${minutes}`,
	}
}

function updateMilestoneDateFromInput(value: string) {
	milestoneDateInput.value = value
}

async function loadProject(projectIdToLoad: number) {
	const projectService = new ProjectService()
	const newProject = await projectService.get(new ProjectModel({id: projectIdToLoad}))
	await useBaseStore().handleSetCurrentProject({project: newProject})
	project.value = newProject
}

async function loadMilestones() {
	if (!projectId.value) {
		return
	}

	loading.value = true
	try {
		milestones.value = await milestoneService.getAll({projectId: projectId.value})
	} finally {
		loading.value = false
	}
}

function resetForm() {
	editingId.value = null
	milestone.value = new MilestoneModel({
		projectId: projectId.value,
		users: [],
	})
	milestoneDateInput.value = ''
	milestoneDateDateInput.value = ''
	milestoneDateTimeInput.value = ''
	formVersion.value++
}

async function editMilestone(existingMilestone: IMilestone) {
	const loadedMilestone = await milestoneService.get(new MilestoneModel({
		id: existingMilestone.id,
		projectId: projectId.value,
	}))

	editingId.value = loadedMilestone.id
	const loadedMilestoneDate = loadedMilestone.milestoneDate
		? new Date(loadedMilestone.milestoneDate)
		: null
	milestone.value = new MilestoneModel({
		...loadedMilestone,
		milestoneDate: loadedMilestoneDate,
		users: loadedMilestone.users.map(user => new UserModel(user)),
	})
	milestoneDateInput.value = formatMilestoneDateInput(loadedMilestone.milestoneDate)
	const milestoneDateParts = formatMilestoneDateParts(loadedMilestone.milestoneDate)
	milestoneDateDateInput.value = milestoneDateParts.date
	milestoneDateTimeInput.value = milestoneDateParts.time
	formVersion.value++
}

async function saveMilestone() {
	if (saving.value || !projectId.value) {
		return
	}

	saving.value = true
	try {
		milestone.value.projectId = projectId.value
		milestone.value.milestoneDate = milestoneDateInput.value === ''
			? null
			: new Date(milestoneDateInput.value)

		const savedMilestone = editingId.value === null
			? await milestoneService.create(milestone.value)
			: await milestoneService.update(milestone.value)

		const existingIndex = milestones.value.findIndex(({id}) => id === savedMilestone.id)
		if (existingIndex === -1) {
			milestones.value.push(savedMilestone)
		} else {
			milestones.value.splice(existingIndex, 1, savedMilestone)
		}

		milestones.value.sort((a, b) => {
			const aDate = a.milestoneDate ? new Date(a.milestoneDate).getTime() : Number.MAX_SAFE_INTEGER
			const bDate = b.milestoneDate ? new Date(b.milestoneDate).getTime() : Number.MAX_SAFE_INTEGER

			if (aDate === bDate) {
				return a.name.localeCompare(b.name)
			}

			return aDate - bDate
		})

		success({
			message: editingId.value === null
				? t('project.milestones.createSuccess')
				: t('project.milestones.updateSuccess'),
		})
		resetForm()
	} finally {
		saving.value = false
	}
}

async function deleteMilestone(milestoneToDelete: IMilestone) {
	if (!window.confirm(t('project.milestones.deleteText', {milestone: milestoneToDelete.name}))) {
		return
	}

	await milestoneService.delete({
		id: milestoneToDelete.id,
		projectId: projectId.value,
	})
	milestones.value = milestones.value.filter(({id}) => id !== milestoneToDelete.id)
	if (editingId.value === milestoneToDelete.id) {
		resetForm()
	}
	success({message: t('project.milestones.deleteSuccess')})
}

async function findUsers(query: string) {
	foundUsers.value = await projectUserService.getAll({projectId: projectId.value}, {s: query}) as IUser[]
}

function getMilestoneStyles(currentMilestone: IMilestone) {
	if (!currentMilestone.hexColor) {
		return {}
	}

	return {
		borderInlineStart: `4px solid ${currentMilestone.hexColor}`,
	}
}

watchEffect(async () => {
	if (!projectId.value) {
		return
	}

	await loadProject(projectId.value)
	await loadMilestones()
	resetForm()
})
</script>

<template>
	<CreateEdit
		:title="$t('project.milestones.title')"
		:has-primary-action="false"
		:wide="true"
	>
		<div class="columns is-variable is-5">
			<div class="column is-5">
				<h3 class="title is-5">
					{{ editingId === null ? $t('project.milestones.create') : $t('project.milestones.edit') }}
				</h3>

				<FormField
					id="milestone-name"
					v-model="milestone.name"
					:label="$t('project.milestones.name')"
					:placeholder="$t('project.milestones.namePlaceholder')"
				/>

				<FormField :label="$t('project.milestones.date')">
					<div
						:key="`milestone-date-${formVersion}`"
						class="milestone-date-inputs"
					>
						<input
							v-model="milestoneDateDateInput"
							class="input"
							type="date"
						>
						<input
							v-model="milestoneDateTimeInput"
							class="input"
							type="time"
						>
					</div>
					<div class="buttons mbs-2">
						<BaseButton
							v-if="milestoneDateInput"
							@click="
								updateMilestoneDateFromInput('');
								milestoneDateDateInput = '';
								milestoneDateTimeInput = '';
							"
						>
							{{ $t('misc.remove') }}
						</BaseButton>
					</div>
				</FormField>

				<FormField :label="$t('project.milestones.color')">
					<ColorPicker v-model="milestone.hexColor" />
				</FormField>

				<FormField :label="$t('project.milestones.users')">
					<Multiselect
						:key="`milestone-users-${formVersion}`"
						v-model="milestone.users"
						:loading="projectUserService.loading"
						:placeholder="$t('project.milestones.usersPlaceholder')"
						:multiple="true"
						:search-results="foundUsers"
						label="name"
						:autocomplete-enabled="false"
						@search="findUsers"
					>
						<template #tag="{item: user, remove}">
							<span class="tag mis-2 mbs-2">
								{{ user.name || user.username }}
								<BaseButton
									class="delete is-small"
									@click="() => remove(user)"
								/>
							</span>
						</template>
						<template #searchResult="{option: user}">
							<User
								:avatar-size="24"
								:show-username="true"
								:user="user"
							/>
						</template>
					</Multiselect>
				</FormField>

				<div class="buttons">
					<BaseButton
						class="is-primary"
						:class="{'is-loading': saving}"
						@click="saveMilestone"
					>
						{{ editingId === null ? $t('project.milestones.create') : $t('misc.save') }}
					</BaseButton>
					<BaseButton
						v-if="editingId !== null"
						@click="resetForm"
					>
						{{ $t('misc.cancel') }}
					</BaseButton>
				</div>
			</div>

			<div class="column">
				<h3 class="title is-5">
					{{ $t('project.milestones.existing') }}
				</h3>

				<p
					v-if="!loading && milestones.length === 0"
					class="has-text-grey"
				>
					{{ $t('project.milestones.empty') }}
				</p>

				<div
					v-for="existingMilestone in milestones"
					:key="existingMilestone.id"
					class="milestone-card"
					:style="getMilestoneStyles(existingMilestone)"
				>
					<div class="milestone-card-header">
						<div>
							<p class="milestone-name">
								{{ existingMilestone.name }}
							</p>
							<p
								v-if="existingMilestone.milestoneDate"
								class="has-text-grey"
							>
								{{ $d(existingMilestone.milestoneDate, 'short') }}
							</p>
						</div>

						<div class="buttons">
							<BaseButton @click="editMilestone(existingMilestone)">
								{{ $t('menu.edit') }}
							</BaseButton>
							<BaseButton
								class="is-danger is-outlined"
								@click="deleteMilestone(existingMilestone)"
							>
								{{ $t('misc.delete') }}
							</BaseButton>
						</div>
					</div>

					<div
						v-if="existingMilestone.users.length > 0"
						class="milestone-users"
					>
						<span
							v-for="assignedUser in existingMilestone.users"
							:key="assignedUser.id"
							class="tag"
						>
							{{ assignedUser.name || assignedUser.username }}
						</span>
					</div>
				</div>
			</div>
		</div>
	</CreateEdit>
</template>

<style lang="scss" scoped>
.milestone-card {
	padding: 1rem;
	border-radius: var(--radius);
	background: var(--white);
	box-shadow: var(--shadow-sm);
}

.milestone-date-inputs {
	display: grid;
	grid-template-columns: minmax(0, 1fr) auto;
	gap: .75rem;
}

.milestone-card + .milestone-card {
	margin-block-start: 1rem;
}

.milestone-card-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
}

.milestone-name {
	font-weight: 700;
}

.milestone-users {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
	margin-block-start: .75rem;
}
</style>
