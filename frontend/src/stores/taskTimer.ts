import {computed, ref} from 'vue'
import {acceptHMRUpdate, defineStore} from 'pinia'

import TaskTimerService from '@/services/taskTimer'
import type {ITask} from '@/modelTypes/ITask'
import type {ITaskTimeTrackingTimer} from '@/modelTypes/ITaskTimeTracking'

type TimerTaskContext = Pick<ITask, 'id' | 'title'>

export const useTaskTimerStore = defineStore('task-timer', () => {
	const currentTimer = ref<ITaskTimeTrackingTimer | null>(null)
	const isLoading = ref(false)

	const isRunning = computed(() => currentTimer.value?.status === 'running')
	const isSnoozed = computed(() => currentTimer.value?.status === 'snoozed')

	async function loadCurrent() {
		isLoading.value = true
		try {
			currentTimer.value = await new TaskTimerService().getCurrent()
			return currentTimer.value
		} finally {
			isLoading.value = false
		}
	}

	function hydrateCurrentTask(task: TimerTaskContext) {
		if (!currentTimer.value || currentTimer.value.taskId !== task.id) {
			return
		}

		currentTimer.value = {
			...currentTimer.value,
			task: {
				...(currentTimer.value.task ?? {}),
				...task,
			} as ITask,
		}
	}

	async function start(task: ITask['id'] | TimerTaskContext) {
		isLoading.value = true
		try {
			const taskId = typeof task === 'number' ? task : task.id
			currentTimer.value = await new TaskTimerService().start(taskId)
			if (typeof task !== 'number') {
				hydrateCurrentTask(task)
			}
			return currentTimer.value
		} finally {
			isLoading.value = false
		}
	}

	async function stop(taskId: ITask['id']) {
		isLoading.value = true
		try {
			const entry = await new TaskTimerService().stop(taskId)
			currentTimer.value = null
			return entry
		} finally {
			isLoading.value = false
		}
	}

	async function adjust(deltaSeconds: number) {
		isLoading.value = true
		try {
			currentTimer.value = await new TaskTimerService().adjust(deltaSeconds)
			return currentTimer.value
		} finally {
			isLoading.value = false
		}
	}

	return {
		currentTimer,
		isLoading,
		isRunning,
		isSnoozed,
		hydrateCurrentTask,
		loadCurrent,
		start,
		stop,
		adjust,
	}
})

if (import.meta.hot) {
	import.meta.hot.accept(acceptHMRUpdate(useTaskTimerStore, import.meta.hot))
}
