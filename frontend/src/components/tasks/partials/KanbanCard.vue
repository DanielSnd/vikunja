<template>
	<div
		class="task card-style loader-container draggable"
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': !colorIsDark(color),
			'has-custom-background-color': color ?? undefined,
		}"
		:style="{'background-color': color ?? undefined}"
		:data-task-id="task.id"
		:data-project-id="task.projectId"
		:data-is-overdue="isOverdue || undefined"
		@click.exact="openTaskDetail()"
		@click.ctrl="() => toggleTaskDone(task)"
		@click.meta="() => toggleTaskDone(task)"
	>
		<!-- Card Main Content Area -->
		<div 
			class="card-main"
			:class="{'has-cover-image': coverImageBlobUrl}"
			:style="coverImageBlobUrl ? `background-image: url('${coverImageBlobUrl}')` : undefined"
		>
			<!-- Card Header -->
			<div class="card-header">
				<!-- Top badges row -->
				<div class="card-badges">
					<span
						v-if="task.dueDate > 0"
						v-tooltip="formatDateLong(task.dueDate)"
						class="badge due-date"
					>
						<span class="icon">
							<Icon :icon="['far', 'calendar-alt']" />
						</span>
						<time :datetime="formatISO(task.dueDate)">
							{{ formatDisplayDate(task.dueDate) }}
						</time>
					</span>
				</div>

				<!-- Task Title -->
				<h3 class="card-title">{{ task.title }}</h3>
				
				<!-- Labels -->
				<Labels
					v-if="task.labels.length > 0"
					:labels="task.labels"
					class="card-labels"
				/>

				<!-- Project Title (if different) -->
				<span
					v-if="projectTitle"
					class="project-badge"
				>
					{{ projectTitle }}
				</span>

				<!-- Progress Bar -->
				<ProgressBar
					v-if="task.percentDone > 0"
					class="card-progress"
					:value="task.percentDone * 100"
				/>
			</div>
		</div>

		<!-- Card Footer -->
		<div 
			class="card-footer"
			:class="`status-${task.done ? 4 : (task.status || 0)}`"
		>
			<div class="footer-left">
				<!-- Task ID / Done indicator -->
				<span class="task-id-badge">
					<Done
						class="kanban-card__done"
						:is-done="task.done"
						variant="small"
					/>
					<Blocked
						class="kanban-card__blocked"
						:is-blocked="task.status === 2"
						variant="small"
					/>
					<Review
						class="kanban-card__review"
						:is-review="task.status === 3"
						variant="small"
					/>
					<!-- <template v-if="task.identifier === ''">
						#{{ task.index }}
					</template>
					<template v-else>
						{{ task.identifier }}
					</template>
					<span
						v-if="showTaskPosition"
						class="tw:text-red-600 tw:ps-2"
					>
						{{ task.position }}
					</span> -->
					<!-- Priority -->
					<EffortLabel
						v-if="task.effort"
						:effort="task.effort"
						:done="task.done"
						class="effort-indicator"
					/>
				</span>

				<!-- Metadata Icons -->
				<div class="metadata-icons">
					<span
						v-if="task.attachments.length > 0"
						v-tooltip="$t('task.attachment.attachments')"
						class="meta-icon"
					>
						<Icon icon="paperclip" />
						<span class="meta-count">{{ task.attachments.length }}</span>
					</span>
					<span
						v-if="!isEditorContentEmpty(task.description)"
						v-tooltip="$t('task.hasDescription')"
						class="meta-icon"
					>
						<Icon icon="align-left" />
					</span>
					<span
						v-if="task.repeatAfter.amount > 0"
						v-tooltip="$t('task.repeat.repeat')"
						class="meta-icon"
					>
						<Icon icon="history" />
					</span>
					<CommentCount
						:task="task"
						class="meta-icon"
					/>
					<ChecklistSummary
						:task="task"
						class="meta-icon checklist"
					/>
				</div>
			</div>

			<div class="footer-right">
				<!-- Priority -->
				<PriorityLabel
					v-if="task.priority"
					:priority="task.priority"
					:done="task.done"
					class="priority-indicator"
				/>
				
				<!-- Assignees -->
				<AssigneeList
					v-if="task.assignees.length > 0"
					:assignees="task.assignees"
					:avatar-size="28"
					class="card-assignees"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed, ref, watch} from 'vue'
import {useRouter} from 'vue-router'

import {useGlobalNow} from '@/composables/useGlobalNow'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import EffortLabel from '@/components/tasks/partials/EffortLabel.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import Done from '@/components/misc/Done.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from './ChecklistSummary.vue'
import CommentCount from './CommentCount.vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'

import {formatDateLong, formatDisplayDate, formatISO} from '@/helpers/time/formatDate'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useTaskStore} from '@/stores/tasks'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import {playPopSound} from '@/helpers/playPop'
import {isEditorContentEmpty} from '@/helpers/editorContentEmpty'
import {useProjectStore} from '@/stores/projects'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import Blocked from '@/components/misc/Blocked.vue'
import Review from '@/components/misc/Review.vue'

const props = withDefaults(defineProps<{
	task: ITask,
	projectId: IProject['id'],
	loading?: boolean,
}>(), {
	loading: false,
})

const emit = defineEmits<{
	'taskCompletedRecurring': [task: ITask]
}>()

const router = useRouter()

const loadingInternal = ref(false)

const color = computed(() => getHexColor(props.task.hexColor))

const projectStore = useProjectStore()

const projectTitle = computed(() => {
	if (props.projectId === props.task.projectId) {
		return
	}
	
	const project = projectStore.projects[props.task.projectId]
	return project?.title
})

const showTaskPosition = computed(() => window.DEBUG_TASK_POSITION)

const {now} = useGlobalNow()
const isOverdue = computed(() => (
	!props.task.done &&
	props.task.dueDate !== null &&
	props.task.dueDate.getTime() > 0 &&
	props.task.dueDate.getTime() <= now.value.getTime()
))

async function toggleTaskDone(task: ITask) {
	const isRecurringTask = task.repeatAfter.amount > 0 || task.repeatMode === TASK_REPEAT_MODES.REPEAT_MODE_MONTH
	const wasBeingMarkedDone = !task.done
	
	loadingInternal.value = true
	try {
		const updatedTask = await useTaskStore().update({
			...task,
			done: !task.done,
		})

		if (updatedTask.done) {
			playPopSound()
		}
		
		// Emit event if this was a recurring task being marked as done
		if (isRecurringTask && wasBeingMarkedDone && updatedTask.done) {
			emit('taskCompletedRecurring', updatedTask)
		}
	} finally {
		loadingInternal.value = false
	}
}

function openTaskDetail() {
	router.push({
		name: 'task.detail',
		params: {id: props.task.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

const coverImageBlobUrl = ref<string | null>(null)

async function maybeDownloadCoverImage() {
	if (!props.task.coverImageAttachmentId) {
		coverImageBlobUrl.value = null
		return
	}

	const attachment = props.task.attachments.find(a => a.id === props.task.coverImageAttachmentId)
	if (!attachment || !SUPPORTED_IMAGE_SUFFIX.some((suffix) => attachment.file.name.toLowerCase().endsWith(suffix))) {
		return
	}

	const attachmentService = new AttachmentService()
	coverImageBlobUrl.value = await attachmentService.getBlobUrl(attachment, PREVIEW_SIZE.LG)
}

watch(
	() => props.task.coverImageAttachmentId,
	maybeDownloadCoverImage,
	{immediate: true},
)
</script>

<style lang="scss" scoped>
$task-background: var(--white);

.task {
	-webkit-touch-callout: none;
	user-select: none;
	cursor: pointer;
	display: flex;
	flex-direction: column;
	font-size: .9rem;
	border-radius: $radius;
	background: $task-background;
	overflow: hidden;
	transition: all 0.2s ease;
	
	// Card-style enhancements
	&.card-style {
		box-shadow: 
			0 1px 3px rgba(0, 0, 0, 0.08),
			0 1px 2px rgba(0, 0, 0, 0.06);
		border: 1px solid var(--grey-200);
		
		&:hover {
			box-shadow: 
				0 4px 12px rgba(0, 0, 0, 0.1),
				0 2px 4px rgba(0, 0, 0, 0.08);
			transform: translateY(-2px);
		}
	}

	&.loader-container.is-loading::after {
		inline-size: 1.5rem;
		block-size: 1.5rem;
		inset-block-start: calc(50% - .75rem);
		inset-inline-start: calc(50% - .75rem);
		border-width: 2px;
	}

	&.is-moving {
		opacity: .5;
	}

	&[data-is-overdue] .due-date {
		color: var(--danger);
		background-color: var(--danger-light);
	}
}

// Card Main Content
.card-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
}

// Cover Image
.card-cover {
	width: 100%;
	height: 120px;
	overflow: hidden;
	background: var(--grey-100);
	
	img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
}

// Card Header
.card-header {
	padding: 0.75rem;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	flex: 1;
}

.card-badges {
	display: flex;
	gap: 0.5rem;
	flex-wrap: wrap;
	margin-bottom: 0.25rem;
}

.badge {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	padding: 0.125rem 0.5rem;
	background: var(--grey-100);
	border-radius: calc($radius / 1.5);
	font-size: 0.75rem;
	line-height: 1.5;
	
	.icon {
		font-size: 0.85rem;
	}
}

.due-date {
	color: var(--text-secondary);
}

.card-title {
	font-family: $family-sans-serif;
	font-size: 0.9rem;
	font-weight: 500;
	line-height: 1.4;
	word-break: break-word;
	margin: 0;
	color: var(--text-primary);
}

.card-labels {
	:deep(.tag) {
		margin: 0.25rem 0.25rem 0 0;
		font-size: 0.7rem;
		padding: 0.125rem 0.5rem;
	}
}

.project-badge {
	display: inline-flex;
	align-items: center;
	padding: 0.25rem 0.5rem;
	background: var(--primary-light);
	color: var(--primary);
	border-radius: calc($radius / 1.5);
	font-size: 0.75rem;
	font-weight: 500;
	width: fit-content;
}
.status-badge {
	display: inline-flex;
	align-items: center;
	flex-shrink: 0;
}

.status-dot {
	width: 8px;
	height: 8px;
	border-radius: 50%;
		
	// More vibrant status colors
	&.status-0 {
		background:hsl(210, 2.5%, 31.4%);
	}

	&.status-1 {
		background: hsl(220, 60%, 90%); // navy blue
		border-top-color: hsl(220, 60%, 75%);
	}

	&.status-2 {
		background: hsl(0, 60%, 90%); // red
		border-top-color: hsl(0, 60%, 75%);
	}

	&.status-3 {
		background: hsl(180, 50%, 90%); // teal
		border-top-color: hsl(180, 50%, 75%);
	}

	&.status-4 {
		background: hsl(140, 50%, 90%); // green
		border-top-color: hsl(140, 50%, 75%);
	}
}

.card-progress {
	margin-top: 0.5rem;
	width: 100%;
	height: 0.375rem;
}

// Card Footer
.card-footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
	padding: 0.5rem 0.75rem;
	background: var(--grey-50);
	border-top: 1px solid var(--grey-200);
	min-height: 48px;

	&.status-0 {
		background: var(--grey-500);
	}
	
	&.status-1 {
		background: hsl(220, 70%, 50%); // navy blue
	}
	
	&.status-2 {
		background: hsl(0, 70%, 50%); // red
	}
	
	&.status-3 {
		background: hsl(180, 70%, 40%); // teal
	}
	
	&.status-4 {
		background: hsl(140, 60%, 40%); // green
	}
}

.footer-left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex: 1;
	min-width: 0;
}

.footer-right {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex-shrink: 0;
}

.task-id-badge {
	display: inline-flex;
	align-items: center;
	padding: 0.125rem 0.375rem;
	// background: var(--grey-200);
	border-radius: calc($radius / 1.5);
	font-size: 0.75rem;
	color: var(--grey-600);
	font-weight: 500;
	flex-shrink: 0;
}

.kanban-card__done {
	margin-inline-end: 0.25rem;
}

.kanban-card__blocked {
	margin-inline-end: 0.25rem;
}

.kanban-card__review {
	margin-inline-end: 0.25rem;
}

.metadata-icons {
	display: flex;
	align-items: center;
	gap: 0.375rem;
	flex-wrap: wrap;
	overflow: hidden;
}

.meta-icon {
	display: inline-flex;
	align-items: center;
	gap: 0.25rem;
	color: var(--grey-600);
	font-size: 0.85rem;
	flex-shrink: 0;
	
	.icon {
		font-size: 1rem;
	}
	
	.meta-count {
		font-size: 0.75rem;
		font-weight: 500;
	}
}

:deep(.comment-count) {
	background: transparent;
	padding: 0;
	color: var(--grey-600);
}

.checklist {
	:deep(.checklist-summary) {
		padding: 0;
		color: var(--grey-600);
	}
}

.priority-indicator {
	font-size: 0.75rem;
	padding: 0.25rem 0.5rem;
	border-radius: calc($radius / 1.5);
	
	.icon {
		height: 1rem;
		padding: 0;
		margin: 0;
	}
}

.card-assignees {
	:deep(.user) {
		margin: 0;
		// border: 2px solid var(--white);
		
		&:not(:first-child) {
			margin-left: -0.5rem;
		}
		
		img {
			width: 28px;
			height: 28px;
		}
	}
}

// Custom Background Color Variants
.has-custom-background-color {
	color: hsl(215, 27.9%, 16.9%);

	.card-footer {
		background: rgba(255, 255, 255, 0.15);
		border-top-color: rgba(0, 0, 0, 0.1);
	}

	.badge,
	.task-id-badge {
		background: rgba(255, 255, 255, 0.2);
		color: hsl(215, 27.9%, 16.9%);
	}

	.project-badge {
		background: rgba(255, 255, 255, 0.25);
		color: hsl(215, 27.9%, 16.9%);
	}

	.meta-icon {
		color: hsl(216.9, 19.1%, 26.7%);
	}
}

.has-light-text {
	color: var(--white);

	.card-title {
		color: var(--white);
	}

	.card-footer {
		background: rgba(0, 0, 0, 0.15);
		border-top-color: rgba(255, 255, 255, 0.1);
	}

	.badge,
	.task-id-badge {
		background: rgba(0, 0, 0, 0.2);
		color: var(--white);
	}

	.project-badge {
		background: rgba(0, 0, 0, 0.25);
		color: var(--white);
	}

	.meta-icon {
		color: hsl(220, 13%, 91%);
		
		.icon svg {
			fill: var(--white);
		}
	}

	:deep(.checklist-summary) {
		color: hsl(220, 13%, 91%);
	}
}
</style>