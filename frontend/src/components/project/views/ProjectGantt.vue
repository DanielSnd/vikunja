<template>
	<ProjectWrapper
		class="project-gantt"
		:is-loading-project="isLoadingProject"
		:project-id="filters.projectId"
		:view-id
	>
		<template #default>
			<Card :has-content="false">
				<div class="gantt-options">
					<FormField :label="$t('project.gantt.range')">
						<Foo
							id="range"
							ref="flatPickerEl"
							v-model="flatPickerDateRange"
							:config="flatPickerConfig"
							class="input"
							:placeholder="$t('project.gantt.range')"
						/>
					</FormField>
					<div
						v-if="!hasDefaultFilters"
						class="field"
					>
						<label
							class="label"
							for="range"
						>Reset</label>
						<div class="control">
							<XButton @click="setDefaultFilters">
								Reset
							</XButton>
						</div>
					</div>
					<FancyCheckbox
						v-model="filters.showTasksWithoutDates"
						is-block
					>
						{{ $t('task.show.noDates') }}
					</FancyCheckbox>
				</div>
			</Card>

			<div class="gantt-chart-container">
				<Card
					:has-content="false"
					:padding="false"
					class="has-overflow"
				>
					<GanttChart
						:filters="filters"
						:milestones="milestones"
						:tasks="tasks"
						:is-loading="isLoading"
						:default-task-start-date="defaultTaskStartDate"
						:default-task-end-date="defaultTaskEndDate"
						:selected-milestone-id="selectedMilestoneId"
						@update:task="updateTask"
						@selectMilestone="toggleMilestone"
					/>
					<TaskForm
						v-if="canWrite"
						@createTask="addGanttTask"
					/>
				</Card>
			</div>

			<GanttMilestoneDetails
				v-if="selectedMilestone"
				:milestone="selectedMilestone"
				:tasks="selectedMilestoneTasks"
				:is-loading="isLoadingMilestoneTasks"
				@close="selectedMilestoneId = null"
			/>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, toRefs, watch} from 'vue'
import type Flatpickr from 'flatpickr'
import {useI18n} from 'vue-i18n'
import type {RouteLocationNormalized} from 'vue-router'

import {useBaseStore} from '@/stores/base'
import {useFlatpickrLanguage} from '@/helpers/useFlatpickrLanguage'

import Foo from '@/components/misc/flatpickr/Flatpickr.vue'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FancyCheckbox from '@/components/input/FancyCheckbox.vue'
import TaskForm from '@/components/tasks/TaskForm.vue'
import FormField from '@/components/input/FormField.vue'

import GanttChart from '@/components/gantt/GanttChart.vue'
import GanttMilestoneDetails from '@/components/gantt/GanttMilestoneDetails.vue'
import {useGanttFilters} from '../../../views/project/helpers/useGanttFilters'
import {PERMISSIONS} from '@/constants/permissions'

import type {DateISO} from '@/types/DateISO'
import type {ITask} from '@/modelTypes/ITask'
import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IMilestone} from '@/modelTypes/IMilestone'
import MilestoneService from '@/services/milestone'
import TaskCollectionService, {getDefaultTaskFilterParams, type TaskFilterParams} from '@/services/taskCollection'
import TaskService from '@/services/task'

type Options = Flatpickr.Options.Options

const props = defineProps<{
	isLoadingProject: boolean,
	route: RouteLocationNormalized
	viewId: IProjectView['id']
}>()


const baseStore = useBaseStore()
const canWrite = computed(() => baseStore.currentProject?.maxPermission > PERMISSIONS.READ)

const {route, viewId} = toRefs(props)
const {
	filters,
	hasDefaultFilters,
	setDefaultFilters,
	tasks,
	isLoading,
	addTask,
	updateTask,
} = useGanttFilters(route, viewId)

const milestoneService = shallowReactive(new MilestoneService())
const taskCollectionService = shallowReactive(new TaskCollectionService())
const milestoneTaskService = shallowReactive(new TaskService())
const milestones = ref<IMilestone[]>([])
const selectedMilestoneId = ref<IMilestone['id'] | null>(null)
const selectedMilestoneTasks = ref<ITask[]>([])
const isLoadingMilestoneTasks = computed(() => milestoneTaskService.loading)
const selectedMilestone = computed(() => milestones.value.find(milestone => milestone.id === selectedMilestoneId.value) ?? null)

watch(
	() => filters.value.projectId,
	async projectId => {
		if (!projectId || projectId < 1) {
			milestones.value = []
			return
		}

		milestones.value = await milestoneService.getAll(
			{projectId},
			{includeParents: true},
		)
		if (selectedMilestoneId.value && !milestones.value.some(milestone => milestone.id === selectedMilestoneId.value)) {
			selectedMilestoneId.value = null
		}
	},
	{immediate: true},
)

async function fetchMilestoneTasks(milestoneId: number, page = 1): Promise<ITask[]> {
	const params: TaskFilterParams = {
		...getDefaultTaskFilterParams(),
		sort_by: ['done', 'id'],
		order_by: ['asc', 'desc'],
		expand: 'subtasks',
		filter: `milestone_id = ${milestoneId}`,
	}

	const tasks = await milestoneTaskService.getAll({} as ITask, params, page) as ITask[]
	if (page < milestoneTaskService.totalPages) {
		return tasks.concat(await fetchMilestoneTasks(milestoneId, page + 1))
	}

	return tasks
}

watch(
	() => [filters.value.projectId, selectedMilestoneId.value] as const,
	async ([projectId, milestoneId]) => {
		if (!projectId || !milestoneId) {
			selectedMilestoneTasks.value = []
			return
		}

		selectedMilestoneTasks.value = await fetchMilestoneTasks(milestoneId)
	},
	{immediate: true},
)

const DEFAULT_DATE_RANGE_DAYS = 7

const today = new Date()
const defaultTaskStartDate: DateISO = new Date(today.setHours(0, 0, 0, 0)).toISOString()
const defaultTaskEndDate: DateISO = new Date(new Date(
	today.getFullYear(),
	today.getMonth(),
	today.getDate() + DEFAULT_DATE_RANGE_DAYS,
).setHours(23, 59, 0, 0)).toISOString()

async function addGanttTask(title: ITask['title']) {
	return await addTask({
		title,
		projectId: filters.value.projectId,
		startDate: defaultTaskStartDate,
		endDate: defaultTaskEndDate,
	})
}

const flatPickerEl = ref<typeof Foo | null>(null)
const flatPickerDateRange = computed<Date[]>({
	get: () => ([
		new Date(filters.value.dateFrom),
		new Date(filters.value.dateTo),
	]),
	set(newVal) {
		const [dateFrom, dateTo] = newVal.map((date) => date?.toISOString())

		// only set after whole range has been selected
		if (!dateTo) return

		Object.assign(filters.value, {dateFrom, dateTo})
	},
})

const {t} = useI18n({useScope: 'global'})
const flatPickerConfig = computed(() => ({
	altFormat: t('date.altFormatShort'),
	altInput: true,
	defaultDate: [filters.value.dateFrom, filters.value.dateTo],
	enableTime: false,
	mode: 'range',
	locale: useFlatpickrLanguage().value,
} as Options))

function toggleMilestone(milestoneId: IMilestone['id']) {
	selectedMilestoneId.value = selectedMilestoneId.value === milestoneId ? null : milestoneId
}
</script>

<style lang="scss" scoped>
.gantt-chart-container {
	padding-block-end: 1rem;
	position: relative;
	z-index: 0;
}

.gantt-options {
	display: flex;
	justify-content: space-between;
	align-items: center;
	margin-block-end: 1rem;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

:global(.link-share-view:not(.has-background)) .gantt-options {
	border: none;
	box-shadow: none;

	.card-content {
		padding: .5rem;
	}
}

.field {
	margin-block-end: 0;
	inline-size: 33%;

	&:not(:last-child) {
		padding-inline-end: .5rem;
	}

	@media screen and (max-width: $tablet) {
		inline-size: 100%;
		max-inline-size: 100%;
		margin-block-start: .5rem;
		padding-inline-end: 0 !important;
	}

	&, .input {
		font-size: .8rem;
	}

	.select,
	.select select {
		block-size: auto;
		inline-size: 100%;
		font-size: .8rem;
	}

	.label {
		font-size: .9rem;
	}
}
</style>
