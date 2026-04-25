<template>
	<Card class="milestone-details">
		<div
			v-if="isLoading"
			class="milestone-details__loading"
		>
			{{ $t('misc.loading') }}
		</div>

		<template v-else>
			<div class="milestone-details__header">
				<div>
					<div class="milestone-details__eyebrow">
						<span
							class="milestone-details__swatch"
							:style="{backgroundColor: milestone.hexColor || undefined}"
						/>
						{{ $t('project.gantt.milestoneDetails') }}
					</div>
					<h3 class="milestone-details__title">
						{{ milestone.name }}
					</h3>
				</div>
				<button
					type="button"
					class="milestone-details__close"
					:aria-label="$t('misc.close')"
					@click="emit('close')"
				>
					&times;
				</button>
			</div>

			<div class="milestone-details__meta">
				<div class="milestone-metric">
					<span class="milestone-metric__label">{{ $t('project.gantt.dueDate') }}</span>
					<strong>{{ dueDateLabel }}</strong>
				</div>
				<div class="milestone-metric">
					<span class="milestone-metric__label">{{ $t('project.gantt.totalCards') }}</span>
					<strong>{{ tasks.length }}</strong>
				</div>
				<div class="milestone-metric">
					<span class="milestone-metric__label">{{ $t('project.gantt.completion') }}</span>
					<strong>{{ completionPercent }}%</strong>
				</div>
			</div>

			<div
				v-if="milestone.users.length > 0"
				class="milestone-details__users"
			>
				<span class="milestone-metric__label">{{ $t('project.gantt.owners') }}</span>
				<div class="milestone-details__user-list">
					<User
						v-for="user in milestone.users"
						:key="user.id"
						:user="user"
						:avatar-size="24"
						show-username
					/>
				</div>
			</div>
			
			<div
				v-if="tasks.length === 0"
				class="milestone-details__empty"
			>
				{{ $t('project.gantt.noMilestoneTasks') }}
			</div>

			<template v-else>

				<div class="milestone-details__progress">
					<div class="milestone-details__section-title">
						{{ $t('project.gantt.statusProgress') }}
					</div>
					<div class="progress-stack">
						<div
							v-for="item in progressItems"
							:key="item.key"
							class="progress-stack__segment"
							:style="{
								backgroundColor: item.color,
								inlineSize: `${item.percentage}%`,
							}"
						/>
					</div>
				</div>

				<div class="milestone-details__effort">
					<div class="milestone-details__section-title">
						{{ $t('project.gantt.effortBreakdown') }}
					</div>
					<div
						v-if="milestoneEffortSummary.total > 0"
						class="effort-summary-list"
						:title="milestoneEffortSummaryTitle"
					>
						<span class="effort-summary effort-summary-total">
							{{ milestoneEffortSummary.completed }} / {{ milestoneEffortSummary.total }}
						</span>
						<span
							v-for="assignee in milestoneEffortSummary.assignees"
							:key="assignee.user.id"
							class="effort-summary assignee-effort-summary"
						>
							<User
								:user="assignee.user"
								:avatar-size="18"
								:show-username="false"
								:is-inline="true"
							/>
							<span>{{ assignee.completed }} / {{ assignee.total }}</span>
						</span>
					</div>
					<div
						v-else
						class="milestone-details__empty"
					>
						{{ $t('project.gantt.noMilestoneEffort') }}
					</div>
				</div>
				
				<div class="milestone-details__chart">
					<div class="milestone-details__section-title">
						{{ $t('project.gantt.burndown') }}
					</div>
					<svg
						class="milestone-burndown"
						viewBox="0 0 640 220"
						role="img"
						:aria-label="$t('project.gantt.burndownChartLabel', {milestone: milestone.name})"
					>
						<defs>
							<linearGradient
								id="milestone-burndown-fill"
								x1="0"
								y1="0"
								x2="0"
								y2="1"
							>
								<stop
									offset="0%"
									stop-color="var(--primary)"
									stop-opacity=".18"
								/>
								<stop
									offset="100%"
									stop-color="var(--primary)"
									stop-opacity=".02"
								/>
							</linearGradient>
						</defs>
						<line
							v-for="gridLine in gridLines"
							:key="gridLine"
							:x1="CHART_PADDING"
							:x2="CHART_WIDTH - CHART_PADDING"
							:y1="gridLine"
							:y2="gridLine"
							class="milestone-burndown__grid"
						/>
						<path
							:d="idealPath"
							class="milestone-burndown__ideal"
						/>
						<path
							:d="actualAreaPath"
							class="milestone-burndown__area"
						/>
						<path
							:d="actualPath"
							class="milestone-burndown__actual"
						/>
						<circle
							v-if="lastPoint"
							:cx="lastPoint.x"
							:cy="lastPoint.y"
							r="5"
							class="milestone-burndown__point"
						/>
						<text
							x="24"
							y="24"
							class="milestone-burndown__value"
						>
							{{ statusCounts.done }}/{{ tasks.length }} {{ $t('project.gantt.cardsComplete') }}
						</text>
						<text
							x="24"
							y="44"
							class="milestone-burndown__label"
						>
							{{ rangeLabel }}
						</text>
					</svg>
				</div>

				<div class="milestone-details__statuses">
					<div class="milestone-details__section-title">
						{{ $t('project.gantt.statusBreakdown') }}
					</div>
					<div class="status-grid">
						<div
							v-for="item in statusItems"
							:key="item.key"
							class="status-card"
						>
							<div
								class="status-card__count"
								:style="{borderColor: item.color}"
							>
								{{ item.count }}
							</div>
							<div
								class="status-card__bar"
								:style="{backgroundColor: item.color}"
							/>
							<div class="status-card__label">
								{{ item.label }}
							</div>
						</div>
					</div>
				</div>

			</template>
		</template>
	</Card>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useI18n} from 'vue-i18n'
import dayjs from 'dayjs'

import Card from '@/components/misc/Card.vue'
import User from '@/components/misc/User.vue'

import type {IMilestone} from '@/modelTypes/IMilestone'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

import {STATUSES} from '@/constants/priorities'
import {getDisplayName} from '@/models/user'

const props = defineProps<{
	milestone: IMilestone
	tasks: ITask[]
	isLoading?: boolean
}>()

const emit = defineEmits<{
	(e: 'close'): void
}>()

const {t, locale} = useI18n({useScope: 'global'})

const CHART_WIDTH = 640
const CHART_HEIGHT = 220
const CHART_PADDING = 24

const dateFormatter = computed(() => new Intl.DateTimeFormat(locale.value, {
	month: 'short',
	day: 'numeric',
	year: 'numeric',
}))

const dueDateLabel = computed(() => {
	if (!props.milestone.milestoneDate) {
		return t('project.gantt.noDueDate')
	}

	return dateFormatter.value.format(props.milestone.milestoneDate)
})

const statusCounts = computed(() => {
	return props.tasks.reduce<Record<'unset' | 'started' | 'blocked' | 'review' | 'done', number>>((acc, task) => {
		if (task.done || task.status === STATUSES.DONE) {
			acc.done++
			return acc
		}

		switch (task.status) {
			case STATUSES.BLOCKED:
				acc.blocked++
				break
			case STATUSES.REVIEW:
				acc.review++
				break
			case STATUSES.IN_PROGRESS:
				acc.started++
				break
			default:
				acc.unset++
		}

		return acc
	}, {
		unset: 0,
		started: 0,
		blocked: 0,
		review: 0,
		done: 0,
	})
})

const completionPercent = computed(() => {
	if (props.tasks.length === 0) {
		return 0
	}

	return Math.round(statusCounts.value.done / props.tasks.length * 100)
})

const statusItems = computed(() => [
	{
		key: 'blocked',
		label: t('task.status.blocked'),
		count: statusCounts.value.blocked,
		color: 'var(--danger)',
	},
	{
		key: 'review',
		label: t('task.status.review'),
		count: statusCounts.value.review,
		color: '#28c3d4',
	},
	{
		key: 'started',
		label: t('task.status.in_progress'),
		count: statusCounts.value.started,
		color: 'var(--primary)',
	},
	{
		key: 'done',
		label: t('task.status.done'),
		count: statusCounts.value.done,
		color: 'var(--success)',
	},
	{
		key: 'unset',
		label: t('task.status.unset'),
		count: statusCounts.value.unset,
		color: 'var(--grey-400)',
	},
	{
		key: 'total',
		label: t('project.gantt.totalCards'),
		count: props.tasks.length,
		color: 'var(--grey-200)',
	},
])

const progressItems = computed(() => {
	if (props.tasks.length === 0) {
		return []
	}

	return statusItems.value
		.filter(item => item.key !== 'total' && item.count > 0)
		.map(item => ({
			...item,
			percentage: item.count / props.tasks.length * 100,
		}))
})

type AssigneeEffortSummary = {
	user: IUser
	completed: number
	total: number
}

type MilestoneEffortSummary = {
	completed: number
	total: number
	assignees: AssigneeEffortSummary[]
}

function isTaskCompleted(task: ITask) {
	return task.done || task.status === STATUSES.DONE
}

const milestoneEffortSummary = computed<MilestoneEffortSummary>(() => {
	const total = props.tasks.reduce((sum, task) => sum + (Number(task.effort) || 0), 0)
	const completed = props.tasks.reduce((sum, task) => {
		return isTaskCompleted(task) ? sum + (Number(task.effort) || 0) : sum
	}, 0)
	const assigneeSummaryMap = new Map<IUser['id'], AssigneeEffortSummary>()

	props.tasks.forEach(task => {
		const effort = Number(task.effort) || 0
		if (effort === 0 || task.assignees.length === 0) {
			return
		}

		const completedEffort = isTaskCompleted(task) ? effort : 0

		task.assignees.forEach(user => {
			const existingSummary = assigneeSummaryMap.get(user.id)
			if (existingSummary) {
				existingSummary.total += effort
				existingSummary.completed += completedEffort
				return
			}

			assigneeSummaryMap.set(user.id, {
				user,
				total: effort,
				completed: completedEffort,
			})
		})
	})

	const assignees = Array.from(assigneeSummaryMap.values())
		.sort((a, b) => getDisplayName(a.user).localeCompare(getDisplayName(b.user)))

	return {
		completed,
		total,
		assignees,
	}
})

const milestoneEffortSummaryTitle = computed(() => {
	if (milestoneEffortSummary.value.total === 0) {
		return '0 / 0'
	}

	const parts = [`${milestoneEffortSummary.value.completed} / ${milestoneEffortSummary.value.total}`]

	milestoneEffortSummary.value.assignees.forEach(assignee => {
		parts.push(`${getDisplayName(assignee.user)} ${assignee.completed} / ${assignee.total}`)
	})

	return parts.join(' • ')
})

const burndownDays = computed(() => {
	if (props.tasks.length === 0) {
		return []
	}

	const relevantDates = [
		props.milestone.created,
		...props.tasks.map(task => task.created),
		...props.tasks.filter(task => task.doneAt).map(task => task.doneAt),
		props.milestone.milestoneDate ?? new Date(),
	].filter(Boolean)

	const start = relevantDates.reduce((earliest, current) => dayjs(current).isBefore(earliest) ? dayjs(current) : earliest, dayjs(relevantDates[0])).startOf('day')
	const end = dayjs(props.milestone.milestoneDate ?? relevantDates.reduce((latest, current) => dayjs(current).isAfter(latest) ? dayjs(current) : latest, dayjs(relevantDates[0]))).startOf('day')

	const days: Array<{date: Date, remaining: number}> = []
	let cursor = start

	while (cursor.isBefore(end) || cursor.isSame(end, 'day')) {
		const completed = props.tasks.filter(task => {
			if (!(task.done || task.status === STATUSES.DONE) || !task.doneAt) {
				return false
			}

			const doneAt = dayjs(task.doneAt)
			return doneAt.isBefore(cursor.endOf('day')) || doneAt.isSame(cursor, 'day')
		}).length

		days.push({
			date: cursor.toDate(),
			remaining: Math.max(props.tasks.length - completed, 0),
		})

		cursor = cursor.add(1, 'day')
	}

	return days
})

const rangeLabel = computed(() => {
	if (burndownDays.value.length === 0) {
		return ''
	}

	const first = burndownDays.value[0].date
	const last = burndownDays.value[burndownDays.value.length - 1].date
	return `${dateFormatter.value.format(first)} - ${dateFormatter.value.format(last)}`
})

const chartPoints = computed(() => {
	if (burndownDays.value.length === 0) {
		return []
	}

	const width = CHART_WIDTH - CHART_PADDING * 2
	const height = CHART_HEIGHT - CHART_PADDING * 2
	const maxRemaining = Math.max(...burndownDays.value.map(point => point.remaining), 1)
	const denominator = Math.max(burndownDays.value.length - 1, 1)

	return burndownDays.value.map((point, index) => ({
		x: CHART_PADDING + width * (index / denominator),
		y: CHART_PADDING + height * (point.remaining / maxRemaining),
	}))
})

const idealPoints = computed(() => {
	if (burndownDays.value.length === 0) {
		return []
	}

	const width = CHART_WIDTH - CHART_PADDING * 2
	const height = CHART_HEIGHT - CHART_PADDING * 2
	const maxRemaining = Math.max(props.tasks.length, 1)
	const denominator = Math.max(burndownDays.value.length - 1, 1)

	return burndownDays.value.map((_, index) => {
		const remaining = props.tasks.length * (1 - index / denominator)
		return {
			x: CHART_PADDING + width * (index / denominator),
			y: CHART_PADDING + height * (remaining / maxRemaining),
		}
	})
})

function toPath(points: Array<{x: number, y: number}>) {
	if (points.length === 0) {
		return ''
	}

	return points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`).join(' ')
}

const actualPath = computed(() => toPath(chartPoints.value))
const idealPath = computed(() => toPath(idealPoints.value))
const actualAreaPath = computed(() => {
	if (chartPoints.value.length === 0) {
		return ''
	}

	const first = chartPoints.value[0]
	const last = chartPoints.value[chartPoints.value.length - 1]
	return `${toPath(chartPoints.value)} L ${last.x} ${CHART_HEIGHT - CHART_PADDING} L ${first.x} ${CHART_HEIGHT - CHART_PADDING} Z`
})

const lastPoint = computed(() => chartPoints.value[chartPoints.value.length - 1] ?? null)

const gridLines = computed(() => {
	const steps = 4
	const chartHeight = CHART_HEIGHT - CHART_PADDING * 2
	return Array.from({length: steps}, (_, index) => CHART_PADDING + chartHeight * (index / (steps - 1)))
})
</script>

<style scoped lang="scss">
.milestone-details {
	margin-block-start: 1rem;
}

.milestone-details__loading {
	color: var(--grey-700);
}

.milestone-details__header {
	display: flex;
	justify-content: space-between;
	align-items: flex-start;
	gap: 1rem;
	margin-block-end: 1rem;
}

.milestone-details__eyebrow {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	color: var(--grey-700);
	font-size: .8rem;
	font-weight: 700;
	letter-spacing: .04em;
	text-transform: uppercase;
}

.milestone-details__swatch {
	inline-size: .8rem;
	block-size: .8rem;
	border-radius: 2px;
}

.milestone-details__title {
	margin: .35rem 0 0;
	font-size: 1.4rem;
	line-height: 1.1;
}

.milestone-details__close {
	border: 0;
	background: transparent;
	color: var(--grey-600);
	cursor: pointer;
	font-size: 1.5rem;
	line-height: 1;
	padding: 0;
}

.milestone-details__meta {
	display: grid;
	grid-template-columns: repeat(3, minmax(0, 1fr));
	gap: .75rem;
	margin-block-end: 1rem;
}

.milestone-metric {
	display: flex;
	flex-direction: column;
	gap: .2rem;
	padding: .9rem 1rem;
	border: 1px solid var(--border);
	border-radius: $radius;
	background: var(--grey-50);
}

.milestone-metric__label,
.milestone-details__section-title {
	color: var(--grey-700);
	font-size: .8rem;
	font-weight: 700;
	text-transform: uppercase;
}

.milestone-details__users {
	margin-block-end: 1.25rem;
}

.milestone-details__user-list {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem 1rem;
	margin-block-start: .5rem;
}

.milestone-details__chart,
.milestone-details__statuses,
.milestone-details__progress,
.milestone-details__effort {
	margin-block-end: 1.25rem;
}

.milestone-burndown {
	inline-size: 100%;
	block-size: auto;
	margin-block-start: .5rem;
	border: 1px solid var(--border);
	border-radius: $radius;
	background:
		linear-gradient(180deg, color-mix(in srgb, var(--primary) 4%, transparent), transparent 55%),
		var(--white);
}

.milestone-burndown__grid {
	stroke: var(--border);
	stroke-width: 1;
}

.milestone-burndown__ideal {
	fill: none;
	stroke: var(--grey-400);
	stroke-dasharray: 8 6;
	stroke-linecap: round;
	stroke-width: 3;
}

.milestone-burndown__area {
	fill: url(#milestone-burndown-fill);
}

.milestone-burndown__actual {
	fill: none;
	stroke: var(--primary);
	stroke-linecap: round;
	stroke-linejoin: round;
	stroke-width: 4;
}

.milestone-burndown__point {
	fill: var(--white);
	stroke: var(--primary);
	stroke-width: 4;
}

.milestone-burndown__value {
	fill: var(--text);
	font-size: 18px;
	font-weight: 700;
}

.milestone-burndown__label {
	fill: var(--grey-700);
	font-size: 14px;
}

.status-grid {
	display: grid;
	grid-template-columns: repeat(6, minmax(0, 1fr));
	gap: .75rem;
	margin-block-start: .5rem;
}

.status-card {
	padding: .8rem;
	border: 1px solid var(--border);
	border-radius: $radius;
	background: var(--white);
	text-align: center;
}

.status-card__count {
	display: flex;
	align-items: center;
	justify-content: center;
	inline-size: 3rem;
	block-size: 3rem;
	margin: 0 auto .5rem;
	border: 2px solid var(--border);
	border-radius: $radius;
	font-size: 1.4rem;
	font-weight: 700;
}

.status-card__bar {
	block-size: .4rem;
	border-radius: 999px;
	margin-block-end: .5rem;
}

.status-card__label {
	font-size: .8rem;
	font-weight: 600;
}

.progress-stack {
	display: flex;
	overflow: hidden;
	block-size: .8rem;
	border-radius: 999px;
	margin-block-start: .5rem;
	background: var(--grey-200);
}

.progress-stack__segment {
	block-size: 100%;
}

.effort-summary-list {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .5rem;
	margin-block-start: .5rem;
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

.milestone-details__empty {
	padding: 1rem;
	border: 1px dashed var(--border);
	border-radius: $radius;
	color: var(--grey-700);
}

@media screen and (max-width: $tablet) {
	.milestone-details__meta {
		grid-template-columns: 1fr;
	}

	.status-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}
}
</style>
