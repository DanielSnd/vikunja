<template>
	<section
		v-if="authStore.authenticated"
		class="hand-tasks mbs-4"
	>
		<div class="hand-tasks__header">
			<div>
				<h3>{{ $t('home.hand.title') }}</h3>
				<p class="hand-tasks__subtitle">
					{{ $t(showExcludedStatuses ? 'home.hand.showingAll' : 'home.hand.showingActive') }}
				</p>
			</div>
			<BaseButton
				v-tooltip="$t(showExcludedStatuses ? 'home.hand.hideUnsetAndDone' : 'home.hand.showUnsetAndDone')"
				class="hand-tasks__toggle"
				:aria-label="$t(showExcludedStatuses ? 'home.hand.hideUnsetAndDone' : 'home.hand.showUnsetAndDone')"
				@click="toggleExcludedStatuses"
			>
				<Icon :icon="showExcludedStatuses ? 'eye-slash' : 'eye'" />
			</BaseButton>
		</div>

		<div
			:class="{'is-loading': isLoading}"
			class="hand-tasks__content loader-container"
		>
			<p
				v-if="!isLoading && handTasks.length === 0"
				class="hand-tasks__empty"
			>
				{{ $t('home.hand.empty') }}
			</p>

			<ul
				v-else
				class="hand-tasks__grid"
			>
				<li
					v-for="task in handTasks"
					:key="task.id"
					class="hand-tasks__card task-item card-item"
				>
					<KanbanCard
						class="kanban-card"
						:task="task"
						:project-id="0"
					/>
				</li>
			</ul>
		</div>
	</section>
</template>

<script lang="ts" setup>
import {ref, watch} from 'vue'

import BaseButton from '@/components/base/BaseButton.vue'
import Icon from '@/components/misc/Icon'
import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'

import {STATUSES} from '@/constants/priorities'
import type {ITask} from '@/modelTypes/ITask'
import TaskService from '@/services/task'
import type {TaskFilterParams} from '@/services/taskCollection'
import {useAuthStore} from '@/stores/auth'

const authStore = useAuthStore()

const taskService = new TaskService()
const handTasks = ref<ITask[]>([])
const isLoading = ref(false)
const showExcludedStatuses = ref(false)
const MAX_HAND_CARDS = 6

function toggleExcludedStatuses() {
	showExcludedStatuses.value = !showExcludedStatuses.value
}

function getBaseTaskFilter() {
	if (!authStore.info?.username) {
		return ''
	}

	return `assignees = '${authStore.info.username}'`
}

async function fetchHandTasks(filter: string, limit?: number): Promise<ITask[]> {
	const params: TaskFilterParams = {
		sort_by: ['due_date', 'id'],
		order_by: ['asc', 'desc'],
		filter,
		filter_include_nulls: false,
		s: '',
	}

	if (typeof limit !== 'undefined') {
		params.per_page = limit
	}

	return taskService.getAll({}, params)
}

async function loadHandTasks() {
	if (!authStore.authenticated || !authStore.info?.username) {
		handTasks.value = []
		return
	}

	const baseFilter = getBaseTaskFilter()
	const activeFilter = `${baseFilter} && status > ${STATUSES.UNSET} && status < ${STATUSES.DONE}`

	isLoading.value = true
	try {
		const visibleTasks = await fetchHandTasks(activeFilter, MAX_HAND_CARDS)

		if (visibleTasks.length < MAX_HAND_CARDS) {
			const unsetTasks = await fetchHandTasks(
				`${baseFilter} && status = ${STATUSES.UNSET}`,
				MAX_HAND_CARDS - visibleTasks.length,
			)
			visibleTasks.push(...unsetTasks)
		}

		if (showExcludedStatuses.value && visibleTasks.length < MAX_HAND_CARDS) {
			const doneTasks = await fetchHandTasks(
				`${baseFilter} && status = ${STATUSES.DONE}`,
				MAX_HAND_CARDS - visibleTasks.length,
			)
			visibleTasks.push(...doneTasks)
		}

		handTasks.value = visibleTasks
	} finally {
		isLoading.value = false
	}
}

watch(
	() => [authStore.authenticated, authStore.info?.username, showExcludedStatuses.value],
	loadHandTasks,
	{immediate: true},
)
</script>

<style scoped lang="scss">
.hand-tasks {
	&__header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		margin-block-end: 1rem;
	}

	&__subtitle {
		margin: 0.25rem 0 0;
		color: var(--grey-500);
		font-size: 0.875rem;
	}

	&__toggle {
		flex-shrink: 0;
	}

	&__content {
		min-block-size: 6rem;
	}

	&__empty {
		margin: 0;
		padding: 1rem 0;
		color: var(--grey-500);
	}

	&__grid {
		margin: 0;
		padding: 0;
		list-style: none;
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		justify-content: flex-start;
		align-items: flex-start;
	}

	&__card {
		margin: 0;
		background-color: transparent;
		padding: 0;
		position: relative;
		border-radius: 8px;
		min-inline-size: 0;
		overflow: visible;
		inline-size: 200px;
		flex: 0 0 200px;
	}

	:deep(.kanban-card) {
		inline-size: 100%;
		max-inline-size: 200px;
	}
}
</style>
