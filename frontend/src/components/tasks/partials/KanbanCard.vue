<template>
	<div
		ref="cardRef"
		class="task card-style loader-container draggable"
		:class="{
			'is-loading': loadingInternal || loading,
			'draggable': !(loadingInternal || loading),
			'has-light-text': !colorIsDark(color),
			'has-custom-background-color': color ?? undefined,
			'is-tilting': tiltActive,
			'is-selected': selected,
		}"
		:style="cardStyle"
		:data-task-id="task.id"
		:data-project-id="task.projectId"
		:data-is-overdue="isOverdue || undefined"
		@click="handleCardClick"
		@mouseenter="handleMouseEnter"
		@mousemove="handleMouseMove"
		@mouseleave="resetTilt"
	>
		<label
			v-if="selectable"
			class="selection-toggle"
			@click.stop
			@mousedown.stop
		>
			<input
				:checked="selected"
				type="checkbox"
				@click.stop
				@change.stop="toggleSelected"
			>
			<span class="selection-toggle__indicator" />
		</label>
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
				<h3 class="card-title">
					{{ task.title }}
				</h3>
				
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

				<span
					v-if="task.timeTrackingTotal || isTaskTimerRunning"
					v-tooltip="timeTrackingTooltip"
					class="header-time-tracking"
				>
					<TimeTrackingIndicator
						:total-seconds="task.timeTrackingTotal"
						:active="isTaskTimerRunning"
					/>
				</span>

				<span
					v-if="task.milestone"
					v-tooltip="task.milestone.name"
					class="milestone-diamond"
					:style="task.milestone.hexColor ? {backgroundColor: task.milestone.hexColor} : undefined"
				/>
			</div>
		</div>

		<!-- Card Footer -->
		<div 
			class="card-footer"
			:class="`status-${task.done ? 4 : (task.status || 0)}`"
		>
			<div class="footer-left">
				<span
					v-if="task.done"
					v-tooltip="$t('task.attributes.done')"
					class="done-indicator"
					:aria-label="$t('task.attributes.done')"
				>
					<Icon icon="check" />
				</span>

				<!-- Task ID / Done indicator -->
				<span
					v-tooltip="$t('project.kanban.cardEffortShortcutHint')"
					class="task-id-badge shortcut-zone shortcut-zone--effort"
					@mouseenter="hoveredFooterAction = 'effort'"
					@mouseleave="clearHoveredFooterAction('effort')"
				>
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
					<span
						v-else
						class="shortcut-placeholder shortcut-placeholder--effort"
						aria-hidden="true"
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
				</div>
			</div>

			<div class="footer-center">
				<div
					ref="assigneeTriggerRef"
					v-tooltip="task.assignees.length === 0 ? $t('project.kanban.cardOwnerShortcutHint') : undefined"
					class="shortcut-zone shortcut-zone--assignee"
					@mouseenter="hoveredFooterAction = 'assignee'"
					@mouseleave="clearHoveredFooterAction('assignee')"
				>
					<AssigneeList
						v-if="task.assignees.length > 0"
						:assignees="task.assignees"
						:avatar-size="38"
						class="card-assignees"
					/>
					<span
						v-else
						class="shortcut-placeholder shortcut-placeholder--assignee"
						aria-hidden="true"
					/>
				</div>

				<div
					v-if="isAssigneeEditorOpen"
					ref="assigneeEditorPopupRef"
					class="assignee-editor-popup"
					@click.stop
					@mousedown.stop
				>
					<EditAssignees
						ref="assigneeEditorRef"
						v-model="editableAssignees"
						:task-id="task.id"
						:project-id="task.projectId"
						:show-project-users-on-empty="true"
						:autofocus="true"
						:silent-updates="true"
					/>
				</div>
			</div>

			<div class="footer-right">
				<ChecklistSummary
					:task="task"
					class="meta-icon checklist"
				/>

				<!-- Priority -->
				<PriorityLabel
					v-if="task.priority"
					:priority="task.priority"
					:done="task.done"
					class="priority-indicator"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {computed, nextTick, onBeforeUnmount, onMounted, ref, watch} from 'vue'
import {useI18n} from 'vue-i18n'
import {useRouter} from 'vue-router'

import {useGlobalNow} from '@/composables/useGlobalNow'

import PriorityLabel from '@/components/tasks/partials/PriorityLabel.vue'
import EffortLabel from '@/components/tasks/partials/EffortLabel.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import Labels from '@/components/tasks/partials/Labels.vue'
import ChecklistSummary from './ChecklistSummary.vue'
import CommentCount from './CommentCount.vue'
import TimeTrackingIndicator from './TimeTrackingIndicator.vue'

import {getHexColor} from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'
import type {IUser} from '@/modelTypes/IUser'
import {SUPPORTED_IMAGE_SUFFIX} from '@/models/attachment'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'

import {formatDateLong, formatDisplayDate, formatISO} from '@/helpers/time/formatDate'
import {formatDuration} from '@/helpers/time/formatDuration'
import {colorIsDark} from '@/helpers/color/colorIsDark'
import {useTaskStore} from '@/stores/tasks'
import {useTaskTimerStore} from '@/stores/taskTimer'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import {playPopSound} from '@/helpers/playPop'
import {useProjectStore} from '@/stores/projects'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'

const props = withDefaults(defineProps<{
	task: ITask,
	projectId: IProject['id'],
	loading?: boolean,
	openBehavior?: 'route' | 'emit',
	selectable?: boolean,
	selected?: boolean,
}>(), {
	loading: false,
	openBehavior: 'route',
	selectable: false,
	selected: false,
})

const emit = defineEmits<{
	'taskCompletedRecurring': [task: ITask],
	'open': [task: ITask],
	'toggleSelected': [task: ITask],
}>()

const router = useRouter()
const {t} = useI18n()

const loadingInternal = ref(false)
const cardRef = ref<HTMLElement | null>(null)
const assigneeTriggerRef = ref<HTMLElement | null>(null)
const assigneeEditorPopupRef = ref<HTMLElement | null>(null)
const assigneeEditorRef = ref<{focus: () => void} | null>(null)
const hoveredFooterAction = ref<'effort' | 'assignee' | null>(null)
const isAssigneeEditorOpen = ref(false)
const editableAssignees = ref<IUser[]>([])

const MAX_TILT_DEG = 7
const tiltActive = ref(false)
const rotateX = ref(0)
const rotateY = ref(0)
const glareX = ref(50)
const glareY = ref(50)

function prefersReducedMotion() {
	return typeof window !== 'undefined' &&
		window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

const cardStyle = computed(() => ({
	'background-color': color.value ?? undefined,
	'--kanban-card-glare-x': `${glareX.value}%`,
	'--kanban-card-glare-y': `${glareY.value}%`,
	transform: prefersReducedMotion()
		? undefined
		: `perspective(900px) rotateX(${rotateX.value}deg) rotateY(${rotateY.value}deg) translateY(${tiltActive.value ? -4 : 0}px)`,
}))

function updateTiltStyles(clientX: number, clientY: number) {
	if (!cardRef.value) {
		return
	}

	const {left, top, width, height} = cardRef.value.getBoundingClientRect()
	if (width === 0 || height === 0) {
		return
	}

	const x = (clientX - left) / width
	const y = (clientY - top) / height
	rotateY.value = Number(((x - 0.5) * MAX_TILT_DEG * 2).toFixed(2))
	rotateX.value = Number(((0.5 - y) * MAX_TILT_DEG * 2).toFixed(2))
	glareX.value = Number((x * 100).toFixed(2))
	glareY.value = Number((y * 100).toFixed(2))
}

function handleMouseEnter(event: MouseEvent) {
	if (prefersReducedMotion()) {
		return
	}

	tiltActive.value = true
	updateTiltStyles(event.clientX, event.clientY)
}

function handleMouseMove(event: MouseEvent) {
	if (prefersReducedMotion()) {
		return
	}

	tiltActive.value = true
	updateTiltStyles(event.clientX, event.clientY)
}

function resetTilt() {
	tiltActive.value = false
	rotateX.value = 0
	rotateY.value = 0
	glareX.value = 50
	glareY.value = 50
}

const color = computed(() => getHexColor(props.task.hexColor))

const projectStore = useProjectStore()
const taskTimerStore = useTaskTimerStore()

const projectTitle = computed(() => {
	if (props.projectId === props.task.projectId) {
		return
	}
	
	const project = projectStore.projects[props.task.projectId]
	return project?.title
})

const {now} = useGlobalNow()
const isOverdue = computed(() => (
	!props.task.done &&
	props.task.dueDate !== null &&
	props.task.dueDate.getTime() > 0 &&
	props.task.dueDate.getTime() <= now.value.getTime()
))

const isTaskTimerRunning = computed(() => (
	taskTimerStore.currentTimer?.status === 'running' &&
	taskTimerStore.currentTimer.taskId === props.task.id
))

const timeTrackingTooltip = computed(() => {
	if (isTaskTimerRunning.value) {
		if (props.task.timeTrackingTotal > 0) {
			return t('task.timeTracking.activeWithTime', {time: formatDuration(props.task.timeTrackingTotal)})
		}

		return t('task.timeTracking.active')
	}

	return t('task.timeTracking.totalWithTime', {time: formatDuration(props.task.timeTrackingTotal)})
})

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
	if (props.openBehavior === 'emit') {
		emit('open', props.task)
		return
	}

	router.push({
		name: 'task.detail',
		params: {id: props.task.id},
		state: {backdropView: router.currentRoute.value.fullPath},
	})
}

function handleCardClick(event: MouseEvent) {
	if (event.altKey && props.selectable) {
		toggleSelected()
		return
	}

	if (event.ctrlKey || event.metaKey) {
		void toggleTaskDone(props.task)
		return
	}

	openTaskDetail()
}

function toggleSelected() {
	emit('toggleSelected', props.task)
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

watch(
	() => props.loading,
	(isLoading) => {
		if (isLoading) {
			resetTilt()
		}
	},
)

watch(
	() => props.task.assignees,
	(assignees) => {
		editableAssignees.value = [...assignees]
	},
	{immediate: true, deep: true},
)

function clearHoveredFooterAction(action: 'effort' | 'assignee') {
	if (hoveredFooterAction.value === action) {
		hoveredFooterAction.value = null
	}
}

async function updateTaskEffort(effort: number) {
	if (loadingInternal.value || props.loading || props.task.effort === effort) {
		return
	}

	loadingInternal.value = true
	try {
		await useTaskStore().update({
			...props.task,
			effort,
		})
	} finally {
		loadingInternal.value = false
	}
}

async function openAssigneeEditor() {
	if (loadingInternal.value || props.loading) {
		return
	}

	isAssigneeEditorOpen.value = true
	hoveredFooterAction.value = 'assignee'

	await nextTick()
	await assigneeEditorRef.value?.focus()
}

function closeAssigneeEditor() {
	isAssigneeEditorOpen.value = false
}

function isTypingTarget(target: EventTarget | null) {
	return target instanceof HTMLElement && (
		target.isContentEditable ||
		['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName)
	)
}

function handleDocumentClick(event: MouseEvent) {
	if (!isAssigneeEditorOpen.value) {
		return
	}

	const target = event.target
	if (
		target instanceof Node &&
		(
			assigneeEditorPopupRef.value?.contains(target) ||
			assigneeTriggerRef.value?.contains(target)
		)
	) {
		return
	}

	closeAssigneeEditor()
}

async function handleWindowKeydown(event: KeyboardEvent) {
	if (loadingInternal.value || props.loading) {
		return
	}

	if (event.isComposing || event.repeat) {
		return
	}

	if (isAssigneeEditorOpen.value) {
		if (event.key === 'Escape') {
			event.preventDefault()
			closeAssigneeEditor()
		}
		return
	}

	if (isTypingTarget(event.target)) {
		return
	}

	const digitMatch = event.code.match(/^(?:Digit|Numpad)([0-9])$/)
	const effortDigit = digitMatch?.[1] ?? (/^[0-9]$/.test(event.key) ? event.key : null)
	if (hoveredFooterAction.value === 'effort' && effortDigit !== null) {
		event.preventDefault()
		await updateTaskEffort(Number(effortDigit))
		return
	}

	if (hoveredFooterAction.value === 'assignee' && event.key.toLowerCase() === 'o') {
		event.preventDefault()
		await openAssigneeEditor()
	}
}

onMounted(() => {
	document.addEventListener('keydown', handleWindowKeydown, true)
	document.addEventListener('mousedown', handleDocumentClick)
})

onBeforeUnmount(() => {
	document.removeEventListener('keydown', handleWindowKeydown, true)
	document.removeEventListener('mousedown', handleDocumentClick)
})
</script>

<style lang="scss" scoped>
$task-background: var(--white);

.task {
	--kanban-card-rotate-x: 0deg;
	--kanban-card-rotate-y: 0deg;
	--kanban-card-glare-x: 50%;
	--kanban-card-glare-y: 50%;

	-webkit-touch-callout: none;
	user-select: none;
	cursor: pointer;
	display: flex;
	flex-direction: column;
	position: relative;
	isolation: isolate;
	font-size: .9rem;
	border-radius: $radius;
	background: $task-background;
	overflow: visible;
	transform: perspective(900px) rotateX(var(--kanban-card-rotate-x)) rotateY(var(--kanban-card-rotate-y)) translateY(0);
	transform-style: preserve-3d;
	will-change: transform, box-shadow;
	transition:
		transform 0.18s ease,
		box-shadow 0.18s ease,
		border-color 0.18s ease;
	
	// Card-style enhancements
	&.card-style {
		box-shadow: 
			0 1px 3px rgba(0, 0, 0, 0.08),
			0 1px 2px rgba(0, 0, 0, 0.06);
		border: 1px solid var(--grey-200);
		
		&:hover {
			box-shadow: 
				0 18px 30px rgba(15, 23, 42, 0.16),
				0 6px 14px rgba(15, 23, 42, 0.1);
			transform: perspective(900px) rotateX(var(--kanban-card-rotate-x)) rotateY(var(--kanban-card-rotate-y)) translateY(-2px);
		}

		&.is-tilting {
			transform: perspective(900px) rotateX(var(--kanban-card-rotate-x)) rotateY(var(--kanban-card-rotate-y)) translateY(-4px);
		}
	}

	&:hover .selection-toggle,
	&.is-selected .selection-toggle,
	.selection-toggle:focus-within {
		opacity: 1;
		pointer-events: auto;
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

	&::before {
		content: '';
		position: absolute;
		inset: 0;
		background:
			radial-gradient(
				circle at var(--kanban-card-glare-x) var(--kanban-card-glare-y),
				rgba(255, 255, 255, 0.24),
				rgba(255, 255, 255, 0.08) 18%,
				transparent 48%
			);
		opacity: 0;
		pointer-events: none;
		transition: opacity 0.18s ease;
	}

	&.is-tilting::before {
		opacity: 1;
	}
}

.selection-toggle {
	position: absolute;
	inset-block-start: -.7rem;
	inset-inline-start: -.7rem;
	z-index: 8;
	opacity: 0;
	pointer-events: none;
	transition: opacity 0.18s ease;
	inline-size: 2rem;
	block-size: 2rem;
	display: grid;
	place-items: center;
	filter: drop-shadow(0 8px 16px rgba(15, 23, 42, 0.18));

	input {
		cursor: pointer;
		position: absolute;
		inset: 0;
		margin: 0;
		opacity: 0;
	}

	&__indicator {
		inline-size: 2rem;
		block-size: 2rem;
		border-radius: 999px;
		border: 2px solid color-mix(in srgb, var(--primary) 70%, var(--grey-300));
		background:
			radial-gradient(circle at 30% 30%, rgba(255, 255, 255, 0.95), rgba(255, 255, 255, 0.72) 45%, rgba(255, 255, 255, 0.35) 100%),
			var(--white);
		box-shadow:
			0 10px 20px rgba(15, 23, 42, 0.14),
			inset 0 1px 0 rgba(255, 255, 255, 0.85);
		display: grid;
		place-items: center;
		transition:
			transform 0.18s ease,
			box-shadow 0.18s ease,
			background-color 0.18s ease,
			border-color 0.18s ease;

		&::after {
			content: '';
			inline-size: .55rem;
			block-size: 1rem;
			border: solid transparent;
			border-width: 0 .18rem .18rem 0;
			transform: rotate(45deg) scale(0.7);
			opacity: 0;
			transition:
				opacity 0.18s ease,
				transform 0.18s ease,
				border-color 0.18s ease;
		}
	}

	&:hover &__indicator {
		transform: scale(1.06);
		box-shadow:
			0 14px 24px rgba(15, 23, 42, 0.18),
			inset 0 1px 0 rgba(255, 255, 255, 0.92);
	}

	input:checked + &__indicator {
		border-color: color-mix(in srgb, var(--primary) 88%, white);
		background:
			radial-gradient(circle at 30% 30%, color-mix(in srgb, white 48%, var(--primary) 52%), var(--primary));
		box-shadow:
			0 14px 28px color-mix(in srgb, var(--primary) 26%, transparent),
			inset 0 1px 0 rgba(255, 255, 255, 0.4);
	}

	input:checked + &__indicator::after {
		opacity: 1;
		transform: rotate(45deg) scale(1);
		border-color: var(--white);
	}
}

// Card Main Content
.card-main {
	flex: 1;
	display: flex;
	flex-direction: column;
	min-height: 0;
	position: relative;
	z-index: 1;
	overflow: hidden;
	border-start-start-radius: inherit;
	border-start-end-radius: inherit;
	transform: translateZ(18px);
	background: hsl(216, 19.2%, 20.4%);

	&.has-cover-image {
		background-position: center;
		background-repeat: no-repeat;
		background-size: cover;

		.card-title,
		.badge,
		.project-badge,
		.meta-icon,
		:deep(.tag),
		:deep(.checklist-summary),
		:deep(.comment-count) {
			text-shadow: 0 0 2px rgba(0, 0, 0, 1), 0 1px 8px rgba(0, 0, 0, 1);
		}
	}
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
	position: relative;
	padding: 0.75rem;
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	flex: 1;
	padding-bottom: 1.25rem;
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
	text-shadow: 0 1px 10px rgb(0,0,0),0 2px 4px rgb(0,0,0,0.1);
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
	background: hsl(221.3, 13.6%, 23.1%);
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

.header-time-tracking {
	position: absolute;
	inset-inline-start: .875rem;
	inset-block-end: .625rem;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	padding: 0;
	color: var(--grey-700);
}

.milestone-diamond {
	position: absolute;
	inset-inline-end: .875rem;
	inset-block-end: .855rem;
	inline-size: .8rem;
	block-size: 0.8rem;
	background: var(--primary);
	border: 2px solid rgba(255, 255, 255, 0.9);
	border-radius: 2px;
	box-shadow: 0 1px 4px rgba(0, 0, 0, 0.28);
	transform: rotate(45deg);
}

// Card Footer
.card-footer {
	--kanban-footer-color: var(--grey-50);

	position: relative;
	z-index: 1;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 0.5rem;
	padding: 0.65rem 0.75rem 0.5rem;
	background: var(--grey-50);
	border-top: 1px solid var(--grey-200);
	min-height: 48px;
	transform: translateZ(28px);

	&.status-0 {
		--kanban-footer-color: #4a5967;
		background:#4a5967;
		border-top-color: hsl(200, 4%, 43%);
	}
	
	&.status-1 {
		--kanban-footer-color:#3c71c2;
		background: #3c71c2; // navy blue
		border-top-color: hsl(220, 82%, 78%);
	}
	
	&.status-2 {
		--kanban-footer-color: hsl(0, 70%, 50%);
		background: hsl(0, 70%, 50%); // red
		border-top-color: hsl(0, 82%, 78%);

	}
	
	&.status-3 {
		--kanban-footer-color: hsl(180, 70%, 40%);
		background: hsl(180, 70%, 40%); // teal
		border-top-color: hsl(180, 82%, 78%);

	}
	
	&.status-4 {
		--kanban-footer-color: hsl(140, 60%, 40%);
		background: hsl(140, 60%, 40%); // green
		border-top-color: hsl(140, 82%, 78%);

	}
}

.footer-left {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	flex: 1;
	min-width: 0;
}

.footer-center {
	position: absolute;
	inset-inline-start: 50%;
	inset-block-start: 0;
	transform: translate(-50%, -26%);
	z-index: 2;
	display: flex;
	justify-content: center;
	pointer-events: none;

	.card-assignees {
		pointer-events: auto;
	}
}

.shortcut-zone {
	position: relative;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	border-radius: calc($radius / 1.5);
	pointer-events: auto;
}

.shortcut-zone--effort {
	min-inline-size: 2.5rem;
	min-block-size: 1.5rem;
}

.shortcut-zone--assignee {
	min-inline-size: 4.5rem;
	min-block-size: 2.75rem;
	padding-inline: 0.25rem;
}

.shortcut-placeholder {
	display: inline-block;
	border-radius: 999px;
}

.shortcut-placeholder--effort {
	inline-size: 2rem;
	block-size: 1.1rem;
}

.shortcut-placeholder--assignee {
	inline-size: 4rem;
	block-size: 2rem;
}

.assignee-editor-popup {
	position: absolute;
	inset-block-end: calc(100% + 0.5rem);
	inset-inline-start: 50%;
	transform: translateX(-50%);
	inline-size: min(18rem, calc(100vw - 2rem));
	padding: 0.5rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--white);
	box-shadow:
		0 18px 30px rgba(15, 23, 42, 0.16),
		0 6px 14px rgba(15, 23, 42, 0.1);
	pointer-events: auto;
}

.footer-right {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	margin-inline-start: auto;
	flex-shrink: 0;
}

.task-id-badge {
	display: inline-flex;
	align-items: center;
	padding: 0.01rem 0.025rem;
	// background: var(--grey-200);
	border-radius: calc($radius / 1.5);
	font-size: 1.1rem;
	color: var(--grey-600);
	font-weight: 500;
	flex-shrink: 0;
}

.done-indicator {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	width: 2rem;
	height: 2rem;
	border-radius: 12px;
	background: hsl(123.2, 62.4%, 29.2%);
	color: #b7e0c5;
	flex-shrink: 0;
	box-shadow: 0 3px 8px rgba(15, 23, 42, 0.18);

	.icon {
		font-size: 1.1rem;
	}
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
	:deep(.assignee) {
		&:not(:first-child) {
			margin-inline-start: -0.75rem;
		}
	}

	:deep(.user) {
		margin: 0;
		
		&:not(:first-child) {
			margin-left: -0.5rem;
		}

		img {
			width: 46px;
			height: 46px;
			border: 3px solid var(--kanban-footer-color);
			box-shadow: 0 3px 6px rgba(15, 23, 42, 0.2);
		}
	}

	:deep(.user img) {
		border-color: var(--kanban-footer-color);
		
		margin-inline-end: 0;
	}
}

// Custom Background Color Variants
.has-custom-background-color {
	color: hsl(215, 27.9%, 16.9%);

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

@media (hover: none) {
	.task {
		transform: none;
	}

	.task.card-style:hover,
	.task.card-style.is-tilting {
		transform: translateY(-2px);
	}

	.task::before {
		display: none;
	}
}

@media (prefers-reduced-motion: reduce) {
	.task {
		transform: none;
		transition:
			box-shadow 0.18s ease,
			border-color 0.18s ease;
	}

	.task.card-style:hover,
	.task.card-style.is-tilting {
		transform: translateY(-2px);
	}

	.task::before {
		display: none;
	}
}
</style>
