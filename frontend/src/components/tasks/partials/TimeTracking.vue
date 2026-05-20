<template>
	<div class="time-tracking">
		<div class="time-tracking__header">
			<div>
				<h3>{{ $t('task.timeTracking.title') }}</h3>
				<p class="time-tracking__total">
					{{ $t('task.timeTracking.total') }}: {{ formatDuration(totalTrackedSeconds) }}
				</p>
			</div>
			<div class="time-tracking__header-actions">
				<XButton
					v-if="canWrite"
					variant="secondary"
					:icon="isTaskTimerRunning ? 'stop' : ['far', 'clock']"
					:disabled="isAnotherTaskTimerRunning"
					@click="toggleTaskTimer()"
				>
					{{ timerButtonLabel }}
				</XButton>
				<XButton
					v-if="canWrite && isTaskTimerRunning"
					variant="secondary"
					icon="minus"
					@click="adjustTimer(-300)"
				/>
				<XButton
					v-if="canWrite && isTaskTimerRunning"
					variant="secondary"
					icon="plus"
					@click="adjustTimer(300)"
				/>
				<XButton
					v-if="canWrite"
					variant="secondary"
					icon="plus"
					@click="toggleAddForm"
				>
					{{ $t('task.timeTracking.add') }}
				</XButton>
			</div>
		</div>

		<div
			v-if="summary.length > 0"
			class="time-tracking__summary"
		>
			<div
				v-for="item in summary"
				:key="item.user.id"
				class="time-tracking__summary-item"
			>
				<span>{{ getDisplayName(item.user) }}</span>
				<strong>{{ formatDuration(item.timeSpent) }}</strong>
			</div>
		</div>

		<div
			v-if="showAddForm"
			class="time-tracking__editor"
		>
			<input
				v-model.number="draftMinutes"
				min="5"
				step="5"
				type="number"
			>
			<input
				v-model="draftTrackedAt"
				type="datetime-local"
			>
			<XButton @click="createEntry()">
				{{ $t('misc.save') }}
			</XButton>
		</div>

		<div
			v-if="entries.length === 0"
			class="time-tracking__empty"
		>
			{{ $t('task.timeTracking.empty') }}
		</div>

		<div
			v-for="entry in entries"
			:key="entry.id"
			class="time-tracking__entry"
		>
			<div
				v-if="editingId === entry.id"
				class="time-tracking__editor"
			>
				<input
					v-model.number="editMinutes"
					min="5"
					step="5"
					type="number"
				>
				<input
					v-model="editTrackedAt"
					type="datetime-local"
				>
				<XButton @click="saveEdit(entry.id)">
					{{ $t('misc.save') }}
				</XButton>
			</div>

			<div
				v-else
				class="time-tracking__entry-row"
			>
				<div class="time-tracking__entry-main">
					<BaseButton
						class="time-tracking__duration"
						@click="beginEdit(entry)"
					>
						{{ formatDuration(entry.timeSpent) }}
					</BaseButton>

					<div
						v-if="entry.user.id === authStore.info.id"
						class="time-tracking__actions"
					>
						<BaseButton @click="adjustEntry(entry, -300)">
							<Icon icon="minus" />
						</BaseButton>
						<BaseButton @click="adjustEntry(entry, 300)">
							<Icon icon="plus" />
						</BaseButton>
						<BaseButton @click="deleteEntry(entry.id)">
							<Icon icon="trash-alt" />
						</BaseButton>
					</div>
				</div>

				<div class="time-tracking__entry-meta">
					<User
						:user="entry.user"
						:avatar-size="24"
						is-inline
					/>
					<span>{{ formatDisplayDate(entry.trackedAt) }}</span>
					<span
						v-if="entry.isSnoozed"
						class="time-tracking__snoozed"
					>
						<Icon icon="moon" />
					</span>
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, ref, shallowReactive, watch} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import User from '@/components/misc/User.vue'
import TaskTimeTrackingService from '@/services/taskTimeTracking'
import {TaskTimeTrackingModel} from '@/models/taskTimeTracking'
import type {ITaskTimeTracking, ITaskTimeTrackingSummary} from '@/modelTypes/ITaskTimeTracking'
import {formatDisplayDate} from '@/helpers/time/formatDate'
import {formatDuration} from '@/helpers/time/formatDuration'
import {getDisplayName} from '@/models/user'
import {useAuthStore} from '@/stores/auth'
import {useTaskTimerStore} from '@/stores/taskTimer'
import {success} from '@/message'
import {useI18n} from 'vue-i18n'

const props = withDefaults(defineProps<{
	taskId: number
	taskTitle?: string
	canWrite?: boolean
}>(), {
	taskTitle: '',
	canWrite: true,
})

const emit = defineEmits<{
	summaryChanged: [payload: {total: number, summary: ITaskTimeTrackingSummary[]}],
}>()

const authStore = useAuthStore()
const taskTimerStore = useTaskTimerStore()
const taskTimeTrackingService = shallowReactive(new TaskTimeTrackingService())
const {t} = useI18n({useScope: 'global'})

const entries = ref<ITaskTimeTracking[]>([])
const showAddForm = ref(false)
const draftMinutes = ref(30)
const draftTrackedAt = ref(toDateTimeLocalString(new Date()))
const editingId = ref<number | null>(null)
const editMinutes = ref(0)
const editTrackedAt = ref('')
const currentTimer = computed(() => taskTimerStore.currentTimer)
const isTaskTimerRunning = computed(() => currentTimer.value?.status === 'running' && currentTimer.value.taskId === props.taskId)
const isAnotherTaskTimerRunning = computed(() => currentTimer.value?.status === 'running' && currentTimer.value.taskId !== props.taskId)
const timerButtonLabel = computed(() => isTaskTimerRunning.value ? t('task.timeTracking.stopTimer') : t('task.timeTracking.startTimer'))

const summary = computed<ITaskTimeTrackingSummary[]>(() => {
	const grouped = new Map<number, ITaskTimeTrackingSummary>()
	for (const entry of entries.value) {
		const current = grouped.get(entry.user.id)
		if (current) {
			current.timeSpent += entry.timeSpent
			continue
		}
		grouped.set(entry.user.id, {
			user: entry.user,
			timeSpent: entry.timeSpent,
		})
	}

	return [...grouped.values()].sort((a, b) => getDisplayName(a.user).localeCompare(getDisplayName(b.user)))
})

const totalTrackedSeconds = computed(() => summary.value.reduce((total, item) => total + item.timeSpent, 0))

watch([summary, totalTrackedSeconds], () => {
	emit('summaryChanged', {
		total: totalTrackedSeconds.value,
		summary: summary.value,
	})
}, {immediate: true})

watch(() => props.taskId, () => {
	loadEntries()
}, {immediate: true})

async function loadEntries() {
	entries.value = await taskTimeTrackingService.getAll({taskId: props.taskId}, {order_by: 'desc'})
}

function toggleAddForm() {
	showAddForm.value = !showAddForm.value
	if (showAddForm.value) {
		draftTrackedAt.value = toDateTimeLocalString(new Date())
	}
}

async function createEntry() {
	const entry = new TaskTimeTrackingModel({
		taskId: props.taskId,
		timeSpent: draftMinutes.value * 60,
		trackedAt: parseDateTimeLocalValue(draftTrackedAt.value),
	})
	entries.value.unshift(await taskTimeTrackingService.create(entry))
	showAddForm.value = false
}

function beginEdit(entry: ITaskTimeTracking) {
	editingId.value = entry.id
	editMinutes.value = Math.max(5, Math.round(entry.timeSpent / 60))
	editTrackedAt.value = toDateTimeLocalString(entry.trackedAt)
}

async function saveEdit(entryId: number) {
	const existing = entries.value.find(entry => entry.id === entryId)
	if (!existing) {
		return
	}

	const updated = await taskTimeTrackingService.update({
		...existing,
		taskId: props.taskId,
		timeSpent: editMinutes.value * 60,
		trackedAt: parseDateTimeLocalValue(editTrackedAt.value),
	})
	const index = entries.value.findIndex(entry => entry.id === entryId)
	entries.value.splice(index, 1, updated)
	editingId.value = null
}

async function adjustEntry(entry: ITaskTimeTracking, deltaSeconds: number) {
	const nextTimeSpent = Math.max(300, entry.timeSpent + deltaSeconds)
	const updated = await taskTimeTrackingService.update({
		...entry,
		taskId: props.taskId,
		timeSpent: nextTimeSpent,
	})
	const index = entries.value.findIndex(existing => existing.id === entry.id)
	entries.value.splice(index, 1, updated)
}

async function deleteEntry(entryId: number) {
	const entry = entries.value.find(existing => existing.id === entryId)
	if (!entry) {
		return
	}

	await taskTimeTrackingService.delete({
		...entry,
		taskId: props.taskId,
	})
	entries.value = entries.value.filter(existing => existing.id !== entryId)
}

function toDateTimeLocalString(date: Date) {
	const offset = date.getTimezoneOffset()
	const local = new Date(date.getTime() - offset * 60 * 1000)
	return local.toISOString().slice(0, 16)
}

function parseDateTimeLocalValue(value: string) {
	if (!value) {
		return new Date()
	}

	const parsed = new Date(value)
	if (Number.isNaN(parsed.getTime())) {
		return new Date()
	}

	return parsed
}

async function toggleTaskTimer() {
	if (isTaskTimerRunning.value) {
		await taskTimerStore.stop(props.taskId)
		success({message: t('task.timeTracking.timerStopped')})
		await loadEntries()
		return
	}

	await taskTimerStore.start({
		id: props.taskId,
		title: props.taskTitle,
	})
	success({message: t('task.timeTracking.timerStarted')})
}

async function adjustTimer(deltaSeconds: number) {
	await taskTimerStore.adjust(deltaSeconds)
}
</script>

<style lang="scss" scoped>
.time-tracking {
	display: flex;
	flex-direction: column;
	gap: 1rem;
}

.time-tracking__header,
.time-tracking__entry-row,
.time-tracking__entry-main,
.time-tracking__summary-item,
.time-tracking__editor,
.time-tracking__actions,
.time-tracking__header-actions {
	display: flex;
	align-items: center;
	gap: .75rem;
}

.time-tracking__header {
	justify-content: space-between;
}

.time-tracking__summary {
	display: grid;
	gap: .5rem;
}

.time-tracking__summary-item,
.time-tracking__entry {
	padding: .75rem 1rem;
	border-radius: .75rem;
	background: var(--grey-100);
}

.time-tracking__entry {
	display: flex;
	flex-direction: column;
	gap: .75rem;
}

.time-tracking__entry-row {
	justify-content: space-between;
	flex-wrap: wrap;
}

.time-tracking__entry-main {
	flex: 1 1 auto;
	min-width: 0;
}

.time-tracking__entry-meta {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: .75rem;
	margin-inline-start: auto;
	color: var(--grey-600);
	font-size: .875rem;
	text-align: right;
}

.time-tracking__duration {
	font-weight: 700;
}

.time-tracking__entry-meta :deep(.user) {
	align-items: center;
}

.time-tracking__entry-meta :deep(.username) {
	white-space: nowrap;
}

.time-tracking__empty,
.time-tracking__total {
	color: var(--grey-600);
}

.time-tracking__snoozed {
	color: rebeccapurple;
}
</style>
