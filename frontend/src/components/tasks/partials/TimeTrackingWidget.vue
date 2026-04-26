<template>
	<div
		v-if="timer"
		class="time-tracking-widget"
		:class="{'time-tracking-widget--snoozed': timer.status === 'snoozed'}"
	>
		<RouterLink
			:to="{name: 'task.detail', params: {id: timer.taskId}}"
			class="time-tracking-widget__task"
		>
			<span class="icon">
				<Icon :icon="timer.status === 'snoozed' ? 'moon' : 'clock'" />
			</span>
			<span class="time-tracking-widget__title">{{ timer.task?.title }}</span>
		</RouterLink>
		<span class="time-tracking-widget__time">{{ elapsedLabel }}</span>
		<div
			v-if="timer.status === 'running'"
			class="time-tracking-widget__actions"
		>
			<BaseButton @click="adjust(-300)">
				<Icon icon="minus" />
			</BaseButton>
			<BaseButton @click="adjust(300)">
				<Icon icon="plus" />
			</BaseButton>
			<BaseButton @click="stopTimer()">
				<Icon icon="stop" />
			</BaseButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import {computed, onMounted, ref, watch} from 'vue'
import {useIntervalFn} from '@vueuse/core'

import BaseButton from '@/components/base/BaseButton.vue'
import {useTaskTimerStore} from '@/stores/taskTimer'
import {formatDurationCompact} from '@/helpers/time/formatDuration'
import {setTitle} from '@/helpers/setTitle'

const taskTimerStore = useTaskTimerStore()
const baseTitle = ref<string | null>(null)
const now = ref(new Date())

useIntervalFn(() => {
	now.value = new Date()
}, 1000, {immediate: true})

const timer = computed(() => taskTimerStore.currentTimer)
const elapsedSeconds = computed(() => {
	if (!timer.value) {
		return 0
	}

	const end = timer.value.status === 'running'
		? now.value
		: (timer.value.stoppedAt ?? now.value)

	return Math.max(0, (end.getTime() - timer.value.startedAt.getTime()) / 1000)
})
const elapsedLabel = computed(() => formatDurationCompact(elapsedSeconds.value))

onMounted(() => {
	taskTimerStore.loadCurrent()
})

watch([timer, elapsedLabel], ([newTimer, newLabel], [oldTimer]) => {
	if (newTimer && !oldTimer) {
		baseTitle.value = document.title
	}

	if (!newTimer) {
		document.title = baseTitle.value ?? 'Vikunja'
		return
	}

	setTitle(`${newLabel} | ${newTimer.task?.title ?? ''}`.trim())
})

async function adjust(deltaSeconds: number) {
	await taskTimerStore.adjust(deltaSeconds)
}

async function stopTimer() {
	if (!timer.value) {
		return
	}

	await taskTimerStore.stop(timer.value.taskId)
}
</script>

<style lang="scss" scoped>
.time-tracking-widget {
	position: fixed;
	inset-inline-start: 1rem;
	inset-block-end: 1rem;
	z-index: 40;
	display: flex;
	align-items: center;
	gap: .75rem;
	background: color-mix(in srgb, var(--primary) 92%, black 8%);
	color: var(--white);
	padding: .75rem 1rem;
	border-radius: 1rem;
	box-shadow: 0 1rem 2rem rgb(0 0 0 / 15%);
}

.time-tracking-widget--snoozed {
	background: color-mix(in srgb, rebeccapurple 88%, black 12%);
}

.time-tracking-widget__task {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	color: inherit;
	min-inline-size: 0;
}

.time-tracking-widget__title {
	max-inline-size: 14rem;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.time-tracking-widget__time {
	font-weight: 700;
	font-variant-numeric: tabular-nums;
}

.time-tracking-widget__actions {
	display: inline-flex;
	gap: .25rem;
}

.time-tracking-widget :deep(button) {
	color: inherit;
}
</style>
