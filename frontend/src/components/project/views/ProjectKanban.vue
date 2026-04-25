<template>
	<ProjectWrapper
		class="project-kanban"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="filter-container">
				<FilterPopup
					v-if="!isSavedFilter(project)"
					v-model="params"
					:view-id="viewId"
					:project-id="projectId"
					@update:modelValue="updateFilters"
				/>
			</div>
		</template>

		<template #default>
			<div
				class="kanban-view"
				:class="{'has-open-task': selectedTaskId !== null}"
			>
				<aside
					class="kanban-detail-sidebar"
					:class="{'is-open': selectedTaskId !== null}"
					:style="sidebarStyle"
				>
					<TaskDetailView
						v-if="selectedTaskId !== null"
						:task-id="selectedTaskId"
						display-mode="sidebar"
						@close="closeTaskDetails"
						@taskDeleted="closeTaskDetails"
						@taskDuplicated="openTask"
					/>
				</aside>
				<div
					v-if="selectedTaskId !== null && !isMobile"
					class="kanban-detail-resizer"
					:class="{'is-resizing': isResizingSidebar}"
					role="separator"
					aria-orientation="vertical"
					aria-label="Resize task details sidebar"
					@mousedown="startSidebarResize"
				/>

				<div
					:class="{ 'is-loading': loading && !oneTaskUpdating}"
					class="kanban kanban-content kanban-bucket-container loader-container"
				>
					<div
						v-if="hasSelectedTasks"
						class="kanban-bulk-actions"
					>
						<div class="kanban-bulk-actions__summary">
							<strong>{{ $t('project.kanban.selectedCards', {count: selectedTasks.length}) }}</strong>
							<XButton
								variant="tertiary"
								:shadow="false"
								@click="clearSelectedTasks"
							>
								{{ $t('project.kanban.clearSelection') }}
							</XButton>
						</div>
						<div class="kanban-bulk-actions__buttons">
							<XButton
								variant="secondary"
								:shadow="false"
								@click="markSelectedAsDone"
							>
								{{ $t('project.kanban.markSelectedDone') }}
							</XButton>
							<XButton
								variant="secondary"
								:shadow="false"
								@click="markSelectedAsForReview"
							>
								{{ $t('project.kanban.markSelectedForReview') }}
							</XButton>
							<XButton
								variant="secondary"
								:shadow="false"
								@click="markSelectedAsUndone"
							>
								{{ $t('project.kanban.markSelectedUndone') }}
							</XButton>
							<XButton
								variant="secondary"
								:shadow="false"
								@click="openBulkActionDialog('assignee')"
							>
								{{ $t('project.kanban.assignSelectedUser') }}
							</XButton>
							<XButton
								variant="secondary"
								:shadow="false"
								@click="openBulkActionDialog('milestone')"
							>
								{{ $t('project.kanban.assignSelectedMilestone') }}
							</XButton>
							<XButton
								class="has-text-danger"
								variant="secondary"
								:shadow="false"
								@click="openBulkActionDialog('delete')"
							>
								{{ $t('project.kanban.deleteSelectedCards') }}
							</XButton>
							<XButton
								variant="secondary"
								:shadow="false"
								@click="openBulkActionDialog('bucket')"
							>
								{{ $t('project.kanban.moveSelectedBucket') }}
							</XButton>
						</div>
					</div>

					<draggable
						v-bind="DRAG_OPTIONS"
						:model-value="buckets"
						group="buckets"
						:disabled="!canWrite || newTaskInputFocused"
						tag="ul"
						:item-key="({id}: IBucket) => `bucket${id}`"
						:component-data="bucketDraggableComponentData"
						@update:modelValue="updateBuckets"
						@end="updateBucketPosition"
						@start="() => dragBucket = true"
					>
						<template #item="{element: bucket, index: bucketIndex }">
							<div
								class="bucket"
								:class="{'is-collapsed': collapsedBuckets[bucket.id]}"
							>
								<div
									class="bucket-header"
									@click="() => unCollapseBucket(bucket)"
								>
									<span
										v-if="bucket.id !== 0 && view?.doneBucketId === bucket.id"
										v-tooltip="$t('project.kanban.doneBucketHint')"
										class="icon is-small has-text-success mie-2"
										@click.stop="() => collapseBucket(bucket)"
									>
										<Icon icon="check-double" />
									</span>
									<div class="title-wrapper">
										<h2
											class="title input"
											:contenteditable="(bucketTitleEditable && canWrite && !collapsedBuckets[bucket.id]) ? true : undefined"
											:spellcheck="false"
											@keydown.enter.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
											@keydown.esc.prevent.stop="!$event.isComposing && ($event.target as HTMLElement).blur()"
											@blur="saveBucketTitle(bucket.id, ($event.target as HTMLElement).textContent as string)"
											@click="focusBucketTitle"
										>
											{{ bucket.title }}
										</h2>
									</div>
									<span
										class="effort-summary-list"
										:title="getBucketEffortSummaryTitle(bucket.id)"
									>
										<span class="effort-summary effort-summary-total">
											{{ bucketEffortSummary[bucket.id]?.remaining ?? 0 }} / {{ bucketEffortSummary[bucket.id]?.total ?? 0 }}
										</span>
										<span
											v-for="assignee in bucketEffortSummary[bucket.id]?.assignees ?? []"
											:key="`${bucket.id}-${assignee.user.id}`"
											class="effort-summary assignee-effort-summary"
										>
											<User
												:user="assignee.user"
												:avatar-size="18"
												:show-username="false"
												:is-inline="true"
											/>
											<span>{{ assignee.remaining }} / {{ assignee.total }}</span>
										</span>
									</span>
									<span
										v-if="bucket.limit > 0 || alwaysShowBucketTaskCount"
										:class="{'is-max': bucket.limit > 0 && bucket.count >= bucket.limit}"
										class="limit"
									>
										{{ bucket.limit > 0 ? `${bucket.count}/${bucket.limit}` : bucket.count }}
									</span>
									<div class="bucket-actions">
										<XButton
											v-tooltip="hiddenDoneBuckets[bucket.id] ? $t('project.kanban.showDoneCards') : $t('project.kanban.hideDoneCards')"
											:shadow="false"
											:icon="hiddenDoneBuckets[bucket.id] ? 'eye' : 'eye-slash'"
											variant="secondary"
											@click.stop="toggleHiddenDoneCards(bucket.id)"
										/>
										<XButton
											v-if="!(bucket.limit > 0 || alwaysShowBucketTaskCount)"
											v-tooltip="bucket.limit > 0 && bucket.count >= bucket.limit ? $t('project.kanban.bucketLimitReached') : ''"
											:shadow="false"
											icon="plus"
											variant="secondary"
											:disabled="bucket.limit > 0 && bucket.count >= bucket.limit"
											@click="toggleShowNewTaskInput(bucket.id)"
										/>
									</div>
									<Dropdown
										v-if="canWrite && !collapsedBuckets[bucket.id]"
										class="is-right options"
										trigger-icon="ellipsis-v"
										:trigger-label="$t('project.kanban.bucketOptions')"
										@close="() => showSetLimitInput = false"
									>
										<div
											v-if="showSetLimitInput"
											class="field has-addons"
										>
											<div class="control">
												<input
													ref="bucketLimitInputRef"
													v-focus.always
													:value="bucket.limit"
													class="input"
													type="number"
													min="0"
													@keyup.esc="() => showSetLimitInput = false"
													@keyup.enter="() => {setBucketLimit(bucket.id, true); showSetLimitInput = false}"
													@input="setBucketLimit(bucket.id)"
												>
											</div>
											<div class="control">
												<XButton
													v-cy="'setBucketLimit'"
													:disabled="bucket.limit < 0"
													:icon="['far', 'save']"
													:shadow="false"
													@click="() => {setBucketLimit(bucket.id, true); showSetLimitInput = false}"
												/>
											</div>
										</div>
										<DropdownItem
											v-else
											@click.stop="showSetLimitInput = true"
										>
											{{
												$t('project.kanban.limit', {limit: bucket.limit > 0 ? bucket.limit : $t('project.kanban.noLimit')})
											}}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.doneBucketHintExtended')"
											:icon-class="{'has-text-success': bucket.id === view?.doneBucketId}"
											icon="check-double"
											@click.stop="toggleDoneBucket(bucket)"
										>
											{{ $t('project.kanban.doneBucket') }}
										</DropdownItem>
										<DropdownItem
											v-tooltip="$t('project.kanban.defaultBucketHint')"
											:icon-class="{'has-text-primary': bucket.id === view?.defaultBucketId}"
											icon="th"
											@click.stop="toggleDefaultBucket(bucket)"
										>
											{{ $t('project.kanban.defaultBucket') }}
										</DropdownItem>
										<DropdownItem
											icon="angles-up"
											@click.stop="() => collapseBucket(bucket)"
										>
											{{ $t('project.kanban.collapse') }}
										</DropdownItem>
										<DropdownItem
											v-tooltip="buckets.length <= 1 ? $t('project.kanban.deleteLast') : ''"
											class="has-text-danger"
											:class="{'is-disabled': buckets.length <= 1}"
											icon-class="has-text-danger"
											icon="trash-alt"
											@click.stop="() => deleteBucketModal(bucket.id)"
										>
											{{ $t('misc.delete') }}
										</DropdownItem>
									</Dropdown>
								</div>

								<draggable
									v-bind="DRAG_OPTIONS"
									:handle="taskDragHandle"
									:delay="isTouchDevice ? 300 : 1000"
									:model-value="bucket.tasks"
									:group="{name: 'tasks', put: shouldAcceptDrop(bucket) && !dragBucket}"
									:disabled="!canWrite"
									:data-bucket-index="bucketIndex"
									tag="ul"
									:item-key="(task: ITask) => `bucket${bucket.id}-task${task.id}`"
									:component-data="getTaskDraggableTaskComponentData(bucket)"
									@update:modelValue="(tasks) => updateTasks(bucket.id, tasks)"
									@start="handleTaskDragStart"
									@end="updateTaskPosition"
								>
									<template #footer>
										<div
											v-if="canCreateTasks"
											class="bucket-footer"
										>
											<div
												v-if="showNewTaskInput === bucket.id"
												class="field"
											>
												<div
													class="control"
													:class="{'is-loading': loading || taskLoading}"
												>
													<input
														v-model="newTaskText"
														v-focus.always
														class="input"
														:disabled="loading || taskLoading || undefined"
														:placeholder="$t('project.kanban.addTaskPlaceholder')"
														type="text"
														@focusout="toggleShowNewTaskInput(bucket.id)"
														@focusin="() => newTaskInputFocused = true"
														@keyup.enter="addTaskToBucket(bucket.id)"
														@keyup.esc="toggleShowNewTaskInput(bucket.id)"
													>
												</div>
												<p
													v-if="newTaskError[bucket.id] && newTaskText === ''"
													class="help is-danger"
												>
													{{ $t('project.create.addTitleRequired') }}
												</p>
											</div>
										</div>
									</template>

									<template #item="{element: task}">
										<div
											v-show="!hiddenDoneBuckets[bucket.id] || !isTaskCompleted(task)"
											class="task-item card-item"
											:class="{
												'is-detail-selected': selectedTaskId === task.id,
												'is-multi-selected': selectedTaskIds.includes(task.id),
											}"
											:data-task-id="task.id"
										>
											<span
												v-if="canWrite && isTouchDevice"
												class="handle"
												@click="openTask(task)"
												@touchstart.passive="onHandleTouchStart"
												@touchmove.passive="onHandleTouchMove"
											/>
											<KanbanCard
												:key="`${task.id}-${task.updated?.getTime?.() ?? 0}-${task.title}`"
												class="kanban-card"
												:task="task"
												:loading="taskUpdating[task.id] ?? false"
												:project-id="projectId"
												open-behavior="emit"
												:selectable="canWrite"
												:selected="selectedTaskIds.includes(task.id)"
												@open="openTask"
												@toggleSelected="toggleTaskSelection"
												@taskCompletedRecurring="handleRecurringTaskCompletion"
											/>
										</div>
									</template>
								</draggable>
							</div>
						</template>
					</draggable>

					<div
						v-if="canWrite && !loading && buckets.length > 0"
						class="bucket new-bucket"
					>
						<input
							v-if="showNewBucketInput"
							v-model="newBucketTitle"
							v-focus.always
							:class="{'is-loading': loading}"
							:disabled="loading || undefined"
							class="input"
							:placeholder="$t('project.kanban.addBucketPlaceholder')"
							type="text"
							@blur="() => showNewBucketInput = false"
							@keyup.enter="createNewBucket"
							@keyup.esc="($event.target as HTMLInputElement).blur()"
						>
						<XButton
							v-else
							:shadow="false"
							class="is-transparent is-fullwidth has-text-centered"
							variant="secondary"
							icon="plus"
							@click="() => showNewBucketInput = true"
						>
							{{ $t('project.kanban.addBucket') }}
						</XButton>
					</div>
				</div>

				<Modal
					:enabled="showBucketDeleteModal"
					@close="showBucketDeleteModal = false"
					@submit="deleteBucket()"
				>
					<template #header>
						<span>{{ $t('project.kanban.deleteHeaderBucket') }}</span>
					</template>

					<template #text>
						<p>
							{{ $t('project.kanban.deleteBucketText1') }}<br>
							{{ $t('project.kanban.deleteBucketText2') }}
						</p>
					</template>
				</Modal>

				<Modal
					:enabled="bulkAction !== null"
					@close="closeBulkActionDialog"
					@submit="submitBulkAction"
				>
					<template #header>
						<span>{{ bulkActionDialogTitle }}</span>
					</template>

					<template #text>
						<div class="bulk-action-dialog">
							<p>
								{{ $t('project.kanban.selectedCards', {count: selectedTasks.length}) }}
							</p>

							<Multiselect
								v-if="bulkAction === 'assignee'"
								v-model="bulkAssignee"
								:loading="projectUserService.loading"
								:placeholder="$t('task.assignee.placeholder')"
								:search-results="foundProjectUsers"
								label="name"
								:select-placeholder="$t('task.assignee.selectPlaceholder')"
								:autocomplete-enabled="false"
								@search="findProjectUsers"
							>
								<template #searchResult="{option: user}">
									<User
										:avatar-size="24"
										:show-username="true"
										:user="user"
									/>
								</template>
							</Multiselect>

							<EditMilestone
								v-else-if="bulkAction === 'milestone'"
								v-model="bulkMilestone"
								:project-id="projectIdWithFallback"
							/>

							<div
								v-else-if="bulkAction === 'bucket'"
								class="field"
							>
								<div class="control">
									<div class="select is-fullwidth">
										<select v-model.number="bulkBucketId">
											<option
												disabled
												:value="0"
											>
												{{ $t('project.kanban.selectBucket') }}
											</option>
											<option
												v-for="bucket in buckets"
												:key="bucket.id"
												:value="bucket.id"
											>
												{{ bucket.title }}
											</option>
										</select>
									</div>
								</div>
							</div>

							<p
								v-else-if="bulkAction === 'delete'"
								class="has-text-danger"
							>
								{{ $t('project.kanban.deleteSelectedCardsConfirm') }}
							</p>
						</div>
					</template>
				</Modal>
			</div>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, nextTick, onUnmounted, ref, shallowReactive, watch, toRef} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {useRouteQuery} from '@vueuse/router'
import {useDebounceFn, useMediaQuery} from '@vueuse/core'
import {useI18n} from 'vue-i18n'
import draggable from 'zhyswan-vuedraggable'
import {klona} from 'klona/lite'

import {PERMISSIONS as Permissions} from '@/constants/permissions'
import {STATUSES} from '@/constants/priorities'
import BucketModel from '@/models/bucket'
import TaskModel from '@/models/task'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IMilestone} from '@/modelTypes/IMilestone'
import type {IUser} from '@/modelTypes/IUser'

import {useBaseStore} from '@/stores/base'
import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useAuthStore} from '@/stores/auth'

import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import FilterPopup from '@/components/project/partials/FilterPopup.vue'
import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'
import EditMilestone from '@/components/tasks/partials/EditMilestone.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import TaskDetailView from '@/views/tasks/TaskDetailView.vue'
import Dropdown from '@/components/misc/Dropdown.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import User from '@/components/misc/User.vue'

import {
	type CollapsedBuckets,
	getCollapsedBucketState,
	saveCollapsedBucketState,
} from '@/helpers/saveCollapsedBucketState'
import {
	type HiddenDoneBuckets,
	getHiddenDoneBucketState,
	saveHiddenDoneBucketState,
} from '@/helpers/saveHiddenDoneBucketState'
import {calculateItemPosition} from '@/helpers/calculateItemPosition'
import {objectToSnakeCase} from '@/helpers/case'
import {AuthenticatedHTTPFactory} from '@/helpers/fetcher'
import {getDisplayName} from '@/models/user'

import {isSavedFilter, useSavedFilter} from '@/services/savedFilter'
import {useTaskDragToProject} from '@/composables/useTaskDragToProject'
import {useWebSocket} from '@/composables/useWebSocket'
import {error, success} from '@/message'
import {useProjectStore} from '@/stores/projects'
import type {TaskFilterParams} from '@/services/taskCollection'
import type {IProjectView} from '@/modelTypes/IProjectView'
import TaskPositionService from '@/services/taskPosition'
import TaskPositionModel from '@/models/taskPosition'
import {i18n} from '@/i18n'
import ProjectViewService from '@/services/projectViews'
import ProjectViewModel from '@/models/projectView'
import TaskBucketService from '@/services/taskBucket'
import TaskAssigneeService from '@/services/taskAssignee'
import TaskBucketModel from '@/models/taskBucket'
import TaskService from '@/services/task'
import ProjectUserService from '@/services/projectUsers'

const props = defineProps<{
	isLoadingProject: boolean,
	projectId: number,
	viewId: IProjectView['id'],
}>()

const projectId = toRef(props, 'projectId')

const DRAG_OPTIONS = {
	// sortable options
	animation: 150,
	ghostClass: 'ghost',
	dragClass: 'task-dragging',
	delayOnTouchOnly: true,
	delay: 1000,
} as const

const MIN_SCROLL_HEIGHT_PERCENT = 0.25
const DEFAULT_DETAIL_SIDEBAR_WIDTH = 720
const MIN_DETAIL_SIDEBAR_WIDTH = 420
const MAX_DETAIL_SIDEBAR_WIDTH = 960

const {t} = useI18n({useScope: 'global'})

const baseStore = useBaseStore()
const kanbanStore = useKanbanStore()
const taskStore = useTaskStore()
const projectStore = useProjectStore()
const authStore = useAuthStore()
const {subscribe, connected: wsConnected} = useWebSocket()

const alwaysShowBucketTaskCount = computed(() => authStore.settings.frontendSettings.alwaysShowBucketTaskCount)
const {handleTaskDropToProject} = useTaskDragToProject()
const taskPositionService = ref(new TaskPositionService())
const taskBucketService = ref(new TaskBucketService())

// Saved filter composable for accessing filter data
const savedFilter = useSavedFilter(() => isSavedFilter({id: projectId.value}) ? projectId.value : undefined).filter

const taskContainerRefs = ref<{ [id: IBucket['id']]: HTMLElement }>({})
const bucketLimitInputRef = ref<HTMLInputElement | null>(null)

const drag = ref(false)
const dragBucket = ref(false)
const sourceBucket = ref(0)

const showBucketDeleteModal = ref(false)
const bucketToDelete = ref(0)
const bucketTitleEditable = ref(false)

const newTaskText = ref('')
const showNewTaskInput = ref<IBucket['id'] | null>(null)

const newBucketTitle = ref('')
const showNewBucketInput = ref(false)
const newTaskError = ref<{ [id: IBucket['id']]: boolean }>({})
const newTaskInputFocused = ref(false)

const showSetLimitInput = ref(false)
const collapsedBuckets = ref<CollapsedBuckets>({})
const hiddenDoneBuckets = ref<HiddenDoneBuckets>({})
const selectedTaskIds = ref<ITask['id'][]>([])
const projectUserService = shallowReactive(new ProjectUserService())
const foundProjectUsers = ref<IUser[]>([])
const bulkAssignee = ref<IUser | null>(null)
const bulkMilestone = ref<IMilestone | null>(null)
const bulkBucketId = ref<IBucket['id']>(0)
const bulkAction = ref<'assignee' | 'milestone' | 'bucket' | 'delete' | null>(null)
const bulkActionLoading = ref(false)

// We're using this to show the loading animation only at the task when updating it
const taskUpdating = ref<{ [id: ITask['id']]: boolean }>({})
const oneTaskUpdating = ref(false)

// URL-synchronized filter parameters
const filter = useRouteQuery('filter')
const s = useRouteQuery('s')

const params = ref<TaskFilterParams>({
	sort_by: [],
	order_by: [],
	filter: '',
	filter_include_nulls: false,
	s: '',
})

watch([filter, s], ([filterValue, sValue]) => {
	params.value.filter = filterValue ?? ''
	params.value.s = sValue ?? ''
}, { immediate: true })

function updateFilters(newParams: TaskFilterParams) {
	// Update all params
	params.value = { ...newParams }
	
	// Sync only filter and s to URL
	filter.value = newParams.filter || undefined
	s.value = newParams.s || undefined
}

const getTaskDraggableTaskComponentData = computed(() => (bucket: IBucket) => {
	return {
		ref: (el: HTMLElement) => setTaskContainerRef(bucket.id, el),
		onScroll: (event: Event) => handleTaskContainerScroll(bucket.id, event.target as HTMLElement),
		type: 'transition-group',
		name: !drag.value ? 'move-card' : null,
		class: [
			'tasks',
			{'dragging-disabled': !canWrite.value},
		],
	}
})

const bucketDraggableComponentData = computed(() => ({
	type: 'transition-group',
	name: !dragBucket.value ? 'move-bucket' : null,
	class: [
		'kanban-bucket-container',
		{'dragging-disabled': !canWrite.value},
	],
}))
const project = computed(() => projectId.value ? projectStore.projects[projectId.value] : null)
const view = computed(() => project.value?.views.find(v => v.id === props.viewId) as IProjectView || null)
const canWrite = computed(() => baseStore.currentProject?.maxPermission > Permissions.READ && view.value.bucketConfigurationMode === 'manual')
const canCreateTasks = computed(() => canWrite.value && projectId.value > 0)

const isTouchDevice = ref(false)
if (typeof window !== 'undefined') {
	isTouchDevice.value = !window.matchMedia('(hover: hover) and (pointer: fine)').matches
}
const taskDragHandle = computed(() => isTouchDevice.value ? '.handle' : undefined)

const router = useRouter()
const route = useRoute()
const touchStartY = ref(0)
const selectedTaskIdQuery = useRouteQuery('taskId')
const isMobile = useMediaQuery('(max-width: 768px)')
const detailSidebarWidth = ref(DEFAULT_DETAIL_SIDEBAR_WIDTH)
const isResizingSidebar = ref(false)

const selectedTaskId = computed<number | null>(() => {
	const value = selectedTaskIdQuery.value
	if (value === undefined || value === null || value === '') {
		return null
	}

	const parsed = Number(value)
	return Number.isNaN(parsed) ? null : parsed
})

const sidebarStyle = computed(() => {
	if (selectedTaskId.value === null) {
		return undefined
	}

	if (isMobile.value) {
		return undefined
	}

	return {
		'--kanban-detail-width': `${detailSidebarWidth.value}px`,
	}
})

function openTask(task: ITask) {
	if (selectedTaskId.value === task.id) {
		closeTaskDetails()
		return
	}

	router.replace({
		name: route.name as string,
		params: route.params,
		query: {
			...route.query,
			taskId: String(task.id),
		},
	})
}

function closeTaskDetails() {
	const {taskId, ...query} = route.query
	void taskId

	router.replace({
		name: route.name as string,
		params: route.params,
		query,
	})
}

function clampSidebarWidth(width: number): number {
	return Math.min(MAX_DETAIL_SIDEBAR_WIDTH, Math.max(MIN_DETAIL_SIDEBAR_WIDTH, width))
}

function handleSidebarResize(event: MouseEvent) {
	if (!isResizingSidebar.value) {
		return
	}

	const viewportWidth = window.innerWidth
	const proposedWidth = document.dir === 'rtl'
		? event.clientX
		: viewportWidth - event.clientX
	detailSidebarWidth.value = clampSidebarWidth(proposedWidth)
}

function stopSidebarResize() {
	if (!isResizingSidebar.value) {
		return
	}

	isResizingSidebar.value = false
	document.removeEventListener('mousemove', handleSidebarResize)
	document.removeEventListener('mouseup', stopSidebarResize)
	document.body.style.userSelect = ''
	document.body.style.cursor = ''
}

function startSidebarResize(event: MouseEvent) {
	if (isMobile.value) {
		return
	}

	event.preventDefault()
	isResizingSidebar.value = true
	document.addEventListener('mousemove', handleSidebarResize)
	document.addEventListener('mouseup', stopSidebarResize)
	document.body.style.userSelect = 'none'
	document.body.style.cursor = 'col-resize'
}

onUnmounted(() => {
	unsubscribeKanbanWs?.()
	stopSidebarResize()
})

function onHandleTouchStart(e: TouchEvent) {
	touchStartY.value = e.touches[0].clientY
}

function onHandleTouchMove(e: TouchEvent) {
	if (drag.value) return

	const currentY = e.touches[0].clientY
	const deltaY = touchStartY.value - currentY
	const scrollContainer = (e.target as HTMLElement).closest('.tasks') as HTMLElement | null
	if (scrollContainer) {
		scrollContainer.scrollTop += deltaY
		touchStartY.value = currentY
	}
}

const buckets = computed(() => kanbanStore.buckets)
const loading = computed(() => kanbanStore.isLoading)
const projectIdWithFallback = computed<number>(() => project.value?.id || projectId.value)
const selectedTasks = computed(() => {
	const taskMap = new Map<ITask['id'], ITask>(
		buckets.value.flatMap(bucket => bucket.tasks.map(task => [task.id, task] as const)),
	)
	return selectedTaskIds.value
		.map(taskId => taskMap.get(taskId))
		.filter((task): task is ITask => typeof task !== 'undefined')
})
const hasSelectedTasks = computed(() => selectedTasks.value.length > 0)
const bulkActionDialogTitle = computed(() => {
	switch (bulkAction.value) {
		case 'assignee':
			return t('project.kanban.assignSelectedUser')
		case 'milestone':
			return t('project.kanban.assignSelectedMilestone')
		case 'bucket':
			return t('project.kanban.moveSelectedBucket')
		case 'delete':
			return t('project.kanban.deleteSelectedCards')
		default:
			return ''
	}
})
type AssigneeEffortSummary = {
	user: IUser,
	remaining: number,
	total: number,
}

type BucketEffortSummary = {
	remaining: number,
	total: number,
	assignees: AssigneeEffortSummary[],
}

const bucketEffortSummary = computed<Record<IBucket['id'], BucketEffortSummary>>(() => {
	return Object.fromEntries(buckets.value.map(bucket => {
		const total = bucket.tasks.reduce((sum, task) => sum + (Number(task.effort) || 0), 0)
		const remaining = bucket.tasks.reduce((sum, task) => sum + (Number(task.effort) || 0), 0) - bucket.tasks.reduce((sum, task) => {
			return isTaskCompleted(task) ? sum : sum + (Number(task.effort) || 0)
		}, 0)
		const assigneeSummaryMap = new Map<IUser['id'], AssigneeEffortSummary>()

		bucket.tasks.forEach(task => {
			const effort = Number(task.effort) || 0
			if (effort === 0 || task.assignees.length === 0) {
				return
			}

			const completedEffort = isTaskCompleted(task) ? effort : 0

			task.assignees.forEach(user => {
				const existingSummary = assigneeSummaryMap.get(user.id)

				if (existingSummary) {
					existingSummary.total += effort
					existingSummary.remaining += completedEffort
					return
				}

				assigneeSummaryMap.set(user.id, {
					user,
					total: effort,
					remaining: completedEffort,
				})
			})
		})

		const assignees = Array.from(assigneeSummaryMap.values())
			.sort((a, b) => getDisplayName(a.user).localeCompare(getDisplayName(b.user)))

		return [bucket.id, {remaining, total, assignees}]
	}))
})

function getBucketEffortSummaryTitle(bucketId: IBucket['id']) {
	const summary = bucketEffortSummary.value[bucketId]
	if (!summary) {
		return '0 / 0'
	}

	const parts = [`${summary.remaining} / ${summary.total}`]

	summary.assignees.forEach(assignee => {
		parts.push(`${getDisplayName(assignee.user)} ${assignee.remaining} / ${assignee.total}`)
	})

	return parts.join(' • ')
}

const taskLoading = computed(() => taskStore.isLoading || taskPositionService.value.loading)
const skipRealtimeReloadUntil = ref(0)
const kanbanWsEvent = computed(() => {
	if (projectId.value === undefined || Number(projectId.value) === 0 || props.viewId === 0) {
		return null
	}

	return `project.${projectId.value}.view.${props.viewId}.kanban.changed`
})
const reloadKanbanFromRealtime = useDebounceFn(() => {
	if (projectId.value === undefined || Number(projectId.value) === 0) {
		return
	}

	if (Date.now() < skipRealtimeReloadUntil.value) {
		return
	}

	kanbanStore.loadBucketsForProject(projectId.value, props.viewId, params.value)
}, 300)

function suppressKanbanRealtimeReload(duration = 2000) {
	skipRealtimeReloadUntil.value = Date.now() + duration
}

let unsubscribeKanbanWs: (() => void) | null = null

watch(
	() => ({
		params: params.value,
		projectId: projectId.value,
		viewId: props.viewId,
	}),
	({params, projectId, viewId}) => {
		if (projectId === undefined || Number(projectId) === 0) {
			return
		}
		collapsedBuckets.value = getCollapsedBucketState(projectId)
		hiddenDoneBuckets.value = getHiddenDoneBucketState(projectId, viewId)
		kanbanStore.loadBucketsForProject(projectId, viewId, params)
	},
	{
		immediate: true,
		deep: true,
	},
)

watch(buckets, (currentBuckets) => {
	const availableTaskIds = new Set(currentBuckets.flatMap(bucket => bucket.tasks.map(task => task.id)))
	selectedTaskIds.value = selectedTaskIds.value.filter(taskId => availableTaskIds.has(taskId))
}, {deep: true})

watch(
	() => taskStore.lastUpdatedTask,
	(updatedTask) => {
		if (updatedTask?.projectId === projectId.value) {
			suppressKanbanRealtimeReload()
		}
	},
)

watch(kanbanWsEvent, (eventName) => {
	unsubscribeKanbanWs?.()
	unsubscribeKanbanWs = null

	if (!eventName) {
		return
	}

	unsubscribeKanbanWs = subscribe(eventName, (msg) => {
		if (msg.event === eventName) {
			reloadKanbanFromRealtime()
		}
	})
}, {immediate: true})

watch(wsConnected, (isConnected, wasConnected) => {
	if (wasConnected && !isConnected) {
		reloadKanbanFromRealtime()
	}
})

function setTaskContainerRef(id: IBucket['id'], el: HTMLElement) {
	if (!el) return
	taskContainerRefs.value[id] = el
}

function handleTaskContainerScroll(id: IBucket['id'], el: HTMLElement) {
	if (!el) {
		return
	}
	const scrollTopMax = el.scrollHeight - el.clientHeight
	const threshold = el.scrollTop + el.scrollTop * MIN_SCROLL_HEIGHT_PERCENT
	if (scrollTopMax > threshold) {
		return
	}

	kanbanStore.loadNextTasksForBucket(
		projectId.value,
		props.viewId,
		params.value,
		id,
	)
}

function toggleTaskSelection(task: ITask) {
	if (selectedTaskIds.value.includes(task.id)) {
		selectedTaskIds.value = selectedTaskIds.value.filter(taskId => taskId !== task.id)
		return
	}

	selectedTaskIds.value = [...selectedTaskIds.value, task.id]
}

function clearSelectedTasks() {
	selectedTaskIds.value = []
}

function openBulkActionDialog(action: 'assignee' | 'milestone' | 'bucket' | 'delete') {
	bulkAction.value = action
	bulkAssignee.value = null
	bulkMilestone.value = null
	bulkBucketId.value = 0
	foundProjectUsers.value = []
}

function closeBulkActionDialog() {
	bulkAction.value = null
	bulkAssignee.value = null
	bulkMilestone.value = null
	bulkBucketId.value = 0
	foundProjectUsers.value = []
}

async function reloadKanbanAfterBulkAction() {
	await kanbanStore.loadBucketsForProject(projectIdWithFallback.value, props.viewId, params.value)
}

async function runBulkAction(
	action: () => Promise<void>,
	message: string,
) {
	if (!hasSelectedTasks.value || bulkActionLoading.value) {
		return
	}

	bulkActionLoading.value = true
	suppressKanbanRealtimeReload()

	try {
		await action()
		success({message})
		clearSelectedTasks()
		closeBulkActionDialog()
		await reloadKanbanAfterBulkAction()
	} catch (e) {
		error(e)
	} finally {
		bulkActionLoading.value = false
	}
}

async function bulkUpdateSelectedTasks(
	values: Partial<ITask>,
	fields: string[],
	message: string,
) {
	await runBulkAction(async () => {
		await AuthenticatedHTTPFactory().post('/tasks/bulk', {
			task_ids: selectedTaskIds.value,
			fields,
			values: objectToSnakeCase(values),
		})
	}, message)
}

async function findProjectUsers(query: string) {
	const response = await projectUserService.getAll({projectId: projectIdWithFallback.value}, {s: query}) as IUser[]
	foundProjectUsers.value = response.map(user => ({
		...user,
		name: getDisplayName(user),
	}))
}

async function markSelectedAsDone() {
	await bulkUpdateSelectedTasks(
		{done: true},
		['done'],
		t('project.kanban.selectedDoneSuccess'),
	)
}

async function markSelectedAsForReview() {
	await bulkUpdateSelectedTasks(
		{done: false, status: STATUSES.REVIEW},
		['done', 'status'],
		t('project.kanban.selectedReviewSuccess'),
	)
}

async function markSelectedAsUndone() {
	await runBulkAction(async () => {
		const taskService = new TaskService()

		await Promise.all(selectedTasks.value.map(task => {
			const nextStatus = task.status === STATUSES.DONE ? STATUSES.UNSET : task.status

			return taskService.update(new TaskModel({
				...task,
				done: false,
				status: nextStatus,
			}))
		}))
	}, t('project.kanban.selectedUndoneSuccess'))
}

async function assignSelectedTasksToUser() {
	if (bulkAssignee.value === null) {
		error({message: t('project.kanban.selectUser')})
		return
	}

	await runBulkAction(async () => {
		const taskAssigneeService = new TaskAssigneeService()

		await Promise.all(selectedTasks.value.map(task => {
			if (task.assignees.some(assignee => assignee.id === bulkAssignee.value?.id)) {
				return Promise.resolve(task)
			}

			return taskAssigneeService.create({
				taskId: task.id,
				userId: bulkAssignee.value?.id ?? 0,
			})
		}))
	}, t('project.kanban.selectedUserAssignedSuccess'))
}

async function assignSelectedTasksToMilestone() {
	if (bulkMilestone.value === null) {
		error({message: t('project.kanban.selectMilestone')})
		return
	}

	await bulkUpdateSelectedTasks(
		{milestoneId: bulkMilestone.value.id},
		['milestone_id'],
		t('project.kanban.selectedMilestoneAssignedSuccess'),
	)
}

async function moveSelectedTasksToBucket() {
	if (bulkBucketId.value === 0) {
		error({message: t('project.kanban.selectBucket')})
		return
	}

	await runBulkAction(async () => {
		for (const task of selectedTasks.value) {
			await taskBucketService.value.update(new TaskBucketModel({
				taskId: task.id,
				bucketId: bulkBucketId.value,
				projectViewId: props.viewId,
				projectId: projectIdWithFallback.value,
			}))
		}
	}, t('project.kanban.selectedBucketMovedSuccess'))
}

async function deleteSelectedTasks() {
	await runBulkAction(async () => {
		const taskService = new TaskService()
		await Promise.all(selectedTasks.value.map(task => taskService.delete(task)))
	}, t('project.kanban.selectedCardsDeletedSuccess'))
}

async function submitBulkAction() {
	switch (bulkAction.value) {
		case 'assignee':
			await assignSelectedTasksToUser()
			return
		case 'milestone':
			await assignSelectedTasksToMilestone()
			return
		case 'bucket':
			await moveSelectedTasksToBucket()
			return
		case 'delete':
			await deleteSelectedTasks()
			return
	}
}

function updateTasks(bucketId: IBucket['id'], tasks: IBucket['tasks']) {
	const bucket = kanbanStore.getBucketById(bucketId)

	if (bucket === undefined) {
		return
	}

	kanbanStore.setBucketById({
		...bucket,
		tasks,
	})
}

async function updateTaskPosition(e) {
	drag.value = false
	oneTaskUpdating.value = true
	suppressKanbanRealtimeReload()

	// Check if dropped on a sidebar project
	const {moved} = await handleTaskDropToProject(e, (task) => {
		kanbanStore.removeTaskInBucket(task)
	})

	if (moved) {
		return
	}

	// If dropped outside kanban
	if (!e.to.dataset.bucketIndex) {
		return
	}

	// While we could just pass the bucket index in through the function call, this would not give us the
	// new bucket id when a task has been moved between buckets, only the new bucket. Using the data-bucket-id
	// of the drop target works all the time.
	const bucketIndex = parseInt(e.to.dataset.bucketIndex)

	const newBucket = buckets.value[bucketIndex]

	// HACK:
	// this is a hacky workaround for a known problem of vue.draggable.next when using the footer slot
	// the problem: https://github.com/SortableJS/vue.draggable.next/issues/108
	// This hack doesn't remove the problem that the ghost item is still displayed below the footer
	// It just makes releasing the item possible.

	// The newIndex of the event doesn't count in the elements of the footer slot.
	// This is why in case the length of the tasks is identical with the newIndex
	// we have to remove 1 to get the correct index.
	const newTaskIndex = newBucket.tasks.length === e.newIndex
		? e.newIndex - 1
		: e.newIndex

	const task = newBucket.tasks[newTaskIndex]
	const oldBucket = buckets.value.find(b => b.id === sourceBucket.value)
	const taskBefore = newBucket.tasks[newTaskIndex - 1] ?? null
	const taskAfter = newBucket.tasks[newTaskIndex + 1] ?? null
	taskUpdating.value[task.id] = true

	const newTask = klona(task) // cloning the task to avoid pinia store manipulation
	newTask.bucketId = newBucket.id
	const position = calculateItemPosition(
		taskBefore !== null ? taskBefore.position : null,
		taskAfter !== null ? taskAfter.position : null,
	)
	
	let bucketHasChanged = false
	if (
		oldBucket !== undefined && // This shouldn't actually be `undefined`, but let's play it safe.
		newBucket.id !== oldBucket.id
	) {
		kanbanStore.setBucketById({
			...oldBucket,
			count: oldBucket.count - 1,
		})
		kanbanStore.setBucketById({
			...newBucket,
			count: newBucket.count + 1,
		})
		bucketHasChanged = true
	}

	try {
		const newPosition = new TaskPositionModel({
			position,
			projectViewId: props.viewId,
			taskId: newTask.id,
		})
		await taskPositionService.value.update(newPosition)
		newTask.position = position
		
		if(bucketHasChanged) {
			const updatedTaskBucket = await taskBucketService.value.update(new TaskBucketModel({
				taskId: newTask.id,
				bucketId: newTask.bucketId,
				projectViewId: props.viewId,
				projectId: projectIdWithFallback.value,
			}))
			Object.assign(newTask, updatedTaskBucket.task)
			if (updatedTaskBucket.bucketId !== newTask.bucketId) {
				kanbanStore.moveTaskToBucket(newTask, updatedTaskBucket.bucketId)
			}
			newTask.bucketId = updatedTaskBucket.bucketId
			if (updatedTaskBucket.bucket) {
				kanbanStore.setBucketById(updatedTaskBucket.bucket, false)
			}
		}
		kanbanStore.setTaskInBucket(newTask)

		// Make sure the first and second task don't both get position 0 assigned
		if (newTaskIndex === 0 && taskAfter !== null && taskAfter.position === 0) {
			const taskAfterAfter = newBucket.tasks[newTaskIndex + 2] ?? null
			const newTaskAfter = klona(taskAfter) // cloning the task to avoid pinia store manipulation
			newTaskAfter.bucketId = newBucket.id
			newTaskAfter.position = calculateItemPosition(
				0,
				taskAfterAfter !== null ? taskAfterAfter.position : null,
			)

			await taskStore.update(newTaskAfter)
		}
	} finally {
		taskUpdating.value[task.id] = false
		oneTaskUpdating.value = false
	}
}

function toggleShowNewTaskInput(bucketId: IBucket['id']) {
	if (loading.value || taskLoading.value) {
		return
	}
	showNewTaskInput.value = showNewTaskInput.value === bucketId 
		? null
		: bucketId
	newTaskInputFocused.value = false
}

async function addTaskToBucket(bucketId: IBucket['id']) {
	if (newTaskText.value === '') {
		newTaskError.value[bucketId] = true
		return
	}
	newTaskError.value[bucketId] = false
	suppressKanbanRealtimeReload()

	const task = await taskStore.createNewTask({
		title: newTaskText.value,
		bucketId,
		projectId: projectIdWithFallback.value,
	})
	newTaskText.value = ''
	kanbanStore.addTaskToBucket(task)
	scrollTaskContainerToTop(bucketId)

	const bucket = kanbanStore.getBucketById(bucketId)
	if (bucket && bucket.limit && bucket.count >= bucket.limit) {
		toggleShowNewTaskInput(bucketId)
	}
}

function scrollTaskContainerToTop(bucketId: IBucket['id']) {
	const bucketEl = taskContainerRefs.value[bucketId]
	if (!bucketEl) {
		return
	}
	bucketEl.scrollTop = 0
}

async function createNewBucket() {
	if (newBucketTitle.value === '') {
		return
	}

	await kanbanStore.createBucket(new BucketModel({
		title: newBucketTitle.value,
		projectId: projectIdWithFallback.value,
		projectViewId: props.viewId,
	}))
	newBucketTitle.value = ''
}

function deleteBucketModal(bucketId: IBucket['id']) {
	if (buckets.value.length <= 1) {
		return
	}

	bucketToDelete.value = bucketId
	showBucketDeleteModal.value = true
}

async function deleteBucket() {
	try {
		await kanbanStore.deleteBucket({
			bucket: new BucketModel({
				id: bucketToDelete.value,
				projectId: projectIdWithFallback.value,
				projectViewId: props.viewId,
			}),
			params: params.value,
		})
		success({message: t('project.kanban.deleteBucketSuccess')})
	} finally {
		showBucketDeleteModal.value = false
	}
}

/** This little helper allows us to drag a bucket around at the title without focusing on it right away. */
async function focusBucketTitle(e: Event) {
	bucketTitleEditable.value = true
	await nextTick()
	const target = e.target as HTMLInputElement
	target.focus()
}

async function saveBucketTitle(bucketId: IBucket['id'], bucketTitle: string) {
	
	const bucket = kanbanStore.getBucketById(bucketId)
	if (bucket?.title === bucketTitle) {
		bucketTitleEditable.value = false
		return
	}
	
	await kanbanStore.updateBucket({
		id: bucketId,
		title: bucketTitle,
		projectId: projectId.value,
	})
	success({message: i18n.global.t('project.kanban.bucketTitleSavedSuccess')})
	bucketTitleEditable.value = false
}

function updateBuckets(value: IBucket[]) {
	// (1) buckets get updated in store and tasks positions get invalidated
	kanbanStore.setBuckets(value)
}

function handleRecurringTaskCompletion() {
	// Only reload if we're in a saved filter and the filter contains date fields
	if (!isSavedFilter(project.value)) {
		return
	}

	const filterContainsDateFields = savedFilter.value?.filters?.filter?.includes('due_date') ||
		savedFilter.value?.filters?.filter?.includes('start_date') ||
		savedFilter.value?.filters?.filter?.includes('end_date')
		
	if (filterContainsDateFields) {
		// Reload the kanban board to refresh tasks that now match/don't match the filter
		kanbanStore.loadBucketsForProject(projectId.value, props.viewId, params.value)
	}
}

// TODO: fix type
function updateBucketPosition(e: { newIndex: number }) {
	// (2) bucket positon is changed
	dragBucket.value = false

	const bucket = buckets.value[e.newIndex]
	const bucketBefore = buckets.value[e.newIndex - 1] ?? null
	const bucketAfter = buckets.value[e.newIndex + 1] ?? null

	kanbanStore.updateBucket({
		id: bucket.id,
		projectId: projectId.value,
		position: calculateItemPosition(
			bucketBefore !== null ? bucketBefore.position : null,
			bucketAfter !== null ? bucketAfter.position : null,
		),
	})
}

async function saveBucketLimit(bucketId: IBucket['id'], limit: number) {
	if (limit < 0) {
		return
	}

	await kanbanStore.updateBucket({
		...kanbanStore.getBucketById(bucketId),
		projectId: projectId.value,
		limit,
	})
	success({message: t('project.kanban.bucketLimitSavedSuccess')})
}

const setBucketLimitCancel = ref<number | null>(null)

async function setBucketLimit(bucketId: IBucket['id'], now: boolean = false) {
	const limit = parseInt(bucketLimitInputRef.value?.value || '')

	if (setBucketLimitCancel.value !== null) {
		clearTimeout(setBucketLimitCancel.value)
	}

	if (now) {
		return saveBucketLimit(bucketId, limit)
	}

	setBucketLimitCancel.value = setTimeout(saveBucketLimit, 2500, bucketId, limit)
}

function shouldAcceptDrop(bucket: IBucket) {
	return (
		// When dragging from a bucket who has its limit reached, dragging should still be possible
		bucket.id === sourceBucket.value ||
		// If there is no limit set, dragging & dropping should always work
		bucket.limit === 0 ||
		// Disallow dropping to buckets which have their limit reached
		bucket.count < bucket.limit
	)
}

function dragstart(bucket: IBucket) {
	drag.value = true
	sourceBucket.value = bucket.id
}

function handleTaskDragStart(e) {
	const taskId = parseInt(e.item.dataset.taskId, 10)
	const bucketIndex = parseInt(e.from.dataset.bucketIndex, 10)
	const bucket = buckets.value[bucketIndex]
	const task = bucket?.tasks.find(t => t.id === taskId)

	if (task) {
		taskStore.setDraggedTask(task)
	}
	dragstart(bucket)
}

async function toggleDefaultBucket(bucket: IBucket) {
	const defaultBucketId = view.value?.defaultBucketId === bucket.id
		? 0
		: bucket.id

	const projectViewService = new ProjectViewService()
	const updatedView = await projectViewService.update(new ProjectViewModel({
		...view.value,
		defaultBucketId,
	}))

	const views = project.value.views.map(v => v.id === view.value?.id ? updatedView : v)
	const updatedProject = {
		...project.value,
		views,
	}

	projectStore.setProject(updatedProject)

	success({message: t('project.kanban.defaultBucketSavedSuccess')})
}

async function toggleDoneBucket(bucket: IBucket) {
	const doneBucketId = view.value?.doneBucketId === bucket.id
		? 0
		: bucket.id
	
	const projectViewService = new ProjectViewService()
	const updatedView = await projectViewService.update(new ProjectViewModel({
		...view.value,
		doneBucketId,
	}))

	const views = project.value.views.map(v => v.id === view.value?.id ? updatedView : v)
	const updatedProject = {
		...project.value,
		views,
	}
	
	projectStore.setProject(updatedProject)
	
	success({message: t('project.kanban.doneBucketSavedSuccess')})
}

function collapseBucket(bucket: IBucket) {
	collapsedBuckets.value[bucket.id] = true
	saveCollapsedBucketState(projectIdWithFallback.value, collapsedBuckets.value)
}

function unCollapseBucket(bucket: IBucket) {
	if (!collapsedBuckets.value[bucket.id]) {
		return
	}

	collapsedBuckets.value[bucket.id] = false
	saveCollapsedBucketState(projectIdWithFallback.value, collapsedBuckets.value)
}

function toggleHiddenDoneCards(bucketId: IBucket['id']) {
	hiddenDoneBuckets.value[bucketId] = !hiddenDoneBuckets.value[bucketId]
	saveHiddenDoneBucketState(projectIdWithFallback.value, props.viewId, hiddenDoneBuckets.value)
}

function isTaskCompleted(task: ITask) {
	return task.done || task.status === STATUSES.DONE
}
</script>

<style lang="scss" scoped>
.control.is-loading {
  &::after {
    inset-block-start: 30%;
    inset-inline-end: 50%;
    transform: translate(-50%, 0);

	--loader-border-color: var(--grey-500);
  }
}
</style>


<style lang="scss">
$ease-out: all .3s cubic-bezier(0.23, 1, 0.32, 1);
$bucket-width: 300px;
$bucket-header-height: 60px;
$bucket-right-margin: 1rem;
$crazy-height-calculation: '100vh - 4.5rem - 1.5rem - 1rem - 1.5rem - 11px';
$crazy-height-calculation-tasks: '#{$crazy-height-calculation} - 1rem - 2.5rem - 2rem - #{$button-height} - 1rem';
$filter-container-height: '1rem - #{$switch-view-height}';

.kanban {
		overflow-x: hidden;
		overflow-y: auto;
		block-size: calc(#{$crazy-height-calculation});
		margin: 0 -1.5rem;
		padding: 0 1.5rem;

		&:focus, .bucket .tasks:focus {
			box-shadow: none;
		}

		@media screen and (max-width: $tablet) {
			block-size: calc(#{$crazy-height-calculation} - #{$filter-container-height} + 9px);
			margin: 0 -0.5rem;
		}

		&-bucket-container {
			display: block;
		}

	.ghost {
		position: relative;

		* {
			opacity: 0;
		}

		&::after {
			content: '';
			position: absolute;
			display: block;
			inset-block-start: 0.25rem;
			inset-inline-end: 0.5rem;
			inset-block-end: 0.25rem;
			inset-inline-start: 0.5rem;
			border: 3px dashed var(--grey-300);
			border-radius: $radius;
		}
	}

	.bucket {
		border-radius: $radius;
		position: relative;
		margin: 0 0 1rem 0;
		padding: 0.75rem;
		display: flex;
		flex-direction: column;

		.tasks {
			display: grid;
			grid-template-columns: repeat(auto-fill,minmax(200px,0.1fr));
			gap: 1rem;
			margin-top: 0.75rem;
		}

		.task-item {
			background-color: transparent; // Remove background from wrapper
			padding: 0; // Remove padding from wrapper
			position: relative;
			border-radius: 8px;
			min-inline-size: 0;
			overflow: visible;
			
			&.card-item {
				// Card-specific styling
				width: 200px; // Adjust based on your needs
				margin-bottom: 0.5rem;
			}

			.handle {
				position: absolute;
				inset: 0;
				z-index: 1;
				opacity: 0;
				touch-action: none;
				-webkit-touch-callout: none;
				user-select: none;
			}
		}

		// Card styling
		.kanban-card {
			background-color: var(--grey-100);
			border-radius: 15px !important;
			box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12), 0 1px 2px rgba(0, 0, 0, 0.24);
			transition: all 0.3s cubic-bezier(.25,.8,.25,1);
			overflow: visible;
			display: flex;
			flex-direction: column;
			
			&:hover {
				box-shadow: 0 4px 6px rgba(0, 0, 0, 0.16), 0 3px 6px rgba(0, 0, 0, 0.23);
				transform: translateY(-2px);
			}
			
			// Card header
			.card-header {
				padding: 0.75rem;
				border-bottom: 1px solid var(--grey-200);
				min-height: 180px;
			}
			
			// Card title
			.card-title {
				font-size: 22px;
				font-weight: 500;
				line-height: 1.4;
				color: var(--text-primary);
				word-wrap: break-word;
				overflow-wrap: break-word;
			}
			
			// Card footer
			.card-footer {
				padding: 0.5rem 0.75rem;
				//background-color: var(--grey-50);
				border-top: 1px solid var(--grey-200);
				display: flex;
				align-items: center;
				justify-content: space-between;
				gap: 0.5rem;
				min-height: 48px;
				border-radius: 1px 1px 10px 10px !important;
			}
			
			// Card metadata (effort, assignee, priority)
			.card-metadata {
				display: flex;
				align-items: center;
				gap: 0.5rem;
				flex-wrap: wrap;
			}
			
			.card-effort {
				display: flex;
				align-items: center;
				gap: 0.25rem;
				font-size: 14px;
				color: var(--text-secondary);
				
				svg {
				width: 20px;
				height: 20px;
				}
			}
			
			.card-assignee {
				width: 24px;
				height: 24px;
				border-radius: 50%;
				display: flex;
				align-items: center;
				justify-content: center;
				font-size: 10px;
				font-weight: 600;
				color: white;
			}
			
			.card-priority {
				svg {
				width: 20px;
				height: 20px;
				}
			}
		}

		.no-move {
			transition: transform 0s;
		}

		h2 {
			font-size: 1.6rem;
			margin: 0;
			font-weight: 600 !important;
		}

		&.new-bucket {
			background: transparent;

			.button {
				background: var(--grey-100);
				inline-size: 100%;
			}
		}

		&.is-collapsed {
			cursor: pointer;

			.tasks, .bucket-footer {
				display: none;
			}
		}
	}

	.bucket-header {
		background-color: transparent;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: .5rem;
		block-size: $bucket-header-height;

		.icon.has-text-success {
			cursor: pointer;
		}

		.limit {
			padding: 0 .5rem;
			font-weight: bold;

			&.is-max {
				color: var(--danger);
			}
		}

		.bucket-actions {
			display: inline-flex;
			align-items: center;
			gap: .5rem;
			margin-inline-start: .5rem;
		}

		.title.input {
			block-size: auto;
			padding: .4rem .5rem;
			display: inline-block;
			cursor: pointer;
		}

		.title-wrapper {
			display: flex;
			align-items: center;
			min-inline-size: 0;
			gap: .5rem;
			flex: 1 1 auto;
			padding-right: .75rem;
		}

		.effort-summary-list {
			display: inline-flex;
			align-items: center;
			gap: .5rem;
			flex: 0 1 auto;
			min-inline-size: 0;
			margin-inline-start: auto;
			margin-inline-end: .75rem;
			white-space: nowrap;
		}

		.effort-summary {
			display: inline-flex;
			align-items: center;
			gap: .35rem;
			font-size: .95rem;
			color: var(--text-light);
			background: #222;
			padding: 8px 10px;
			border-radius: 12px;
			white-space: nowrap;
			flex-shrink: 0;
		}

		.effort-summary-total {
			font-size: 1.15rem;
		}

		.assignee-effort-summary {
			:deep(.user) {
				display: inline-flex;
				align-items: center;
			}

			:deep(.avatar) {
				margin-inline-end: 0;
			}
		}

		@media screen and (max-width: $tablet) {
			.effort-summary-list {
				display: none;
			}
		}
	}

	:deep(.dropdown-trigger) {
		padding: .5rem;
	}

	.bucket-footer {
		position: sticky;
		inset-block-end: 0;
		block-size: min-content;
		padding: .5rem;
		background-color: transparent;
		border-end-start-radius: $radius;
		border-end-end-radius: $radius;
		transform: none;

		.button {
			background-color: transparent;

			&:hover {
				background-color: var(--white);
			}
		}
		}
}

.kanban-view {
	display: flex;
	gap: 1rem;
	min-inline-size: 0;
	align-items: flex-start;

	@media screen and (max-width: $tablet) {
		flex-direction: column;
	}
}

.kanban-bulk-actions {
	position: sticky;
	inset-block-start: 0;
	z-index: 6;
	display: flex;
	flex-wrap: wrap;
	justify-content: flex-end;
	gap: .75rem;
	margin-block-end: 1rem;
	padding: .75rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: color-mix(in srgb, var(--white) 94%, var(--primary) 6%);
	box-shadow: var(--shadow-xs);

	&__summary,
	&__buttons {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: .5rem;
	}

	&__summary {
		margin-inline-end: auto;
	}
}

.bulk-action-dialog {
	display: grid;
	gap: 1rem;
	text-align: start;
}

.kanban-detail-sidebar {
	flex: 0 0 0;
	inline-size: 0;
	min-inline-size: 0;
	opacity: 0;
	overflow: hidden;
	border-radius: $radius;
	background: var(--site-background);
	border: 1px solid transparent;
	transition:
		flex-basis .25s ease,
		inline-size .25s ease,
		opacity .2s ease,
		border-color .2s ease;

	&.is-open {
		flex-basis: var(--kanban-detail-width, min(34rem, 36vw));
		inline-size: var(--kanban-detail-width, min(34rem, 36vw));
		opacity: 1;
		overflow-y: auto;
		block-size: calc(#{$crazy-height-calculation});
		border-color: var(--grey-200);
		box-shadow: var(--shadow-sm);
	}

	@media screen and (max-width: $desktop) {
		&.is-open {
			flex-basis: min(30rem, 42vw);
			inline-size: min(30rem, 42vw);
		}
	}

	@media screen and (max-width: $tablet) {
		inline-size: 100%;

		&.is-open {
			flex-basis: auto;
			inline-size: 100%;
			block-size: auto;
			max-block-size: 70vh;
		}
	}
}

.kanban-content {
	flex: 1 1 auto;
	min-inline-size: 0;
}

.kanban-detail-resizer {
	flex: 0 0 0.75rem;
	align-self: stretch;
	margin-inline: -0.25rem;
	cursor: col-resize;
	position: relative;

	&::before {
		content: '';
		position: absolute;
		inset-block: 0;
		inset-inline-start: calc(50% - 1px);
		inline-size: 2px;
		background: var(--grey-200);
		transition: background-color $transition-duration ease;
	}

	&:hover::before,
	&.is-resizing::before {
		background: var(--primary);
	}
}

// FIXME: This does not seem to work
.task-dragging {
	transform: rotateZ(3deg);
	transition: transform 0.18s ease;
}

.move-card-move {
	transform: rotateZ(3deg);
	transition: transform $transition-duration;
}

.move-card-leave-from,
.move-card-leave-to,
.move-card-leave-active {
	display: none;
}

.task-item.is-detail-selected .kanban-card,
.task-item.is-multi-selected .kanban-card {
	box-shadow:
		0 0 0 8px color-mix(in srgb, var(--primary) 36%, transparent),
		0 10px 24px rgba(0, 0, 0, 0.14);
	transform: translateY(-2px);
	
	&:hover {
		box-shadow:
			0 0 0 12px color-mix(in srgb, var(--primary) 55%, transparent),
			0 10px 24px rgba(0, 0, 0, 0.14);
	}
}
</style>
