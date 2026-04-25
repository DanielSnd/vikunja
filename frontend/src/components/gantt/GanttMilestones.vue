<template>
	<div
		class="gantt-milestones"
		:aria-label="$t('project.gantt.milestonesRow')"
		:style="{width: `${totalWidth}px`}"
	>
		<button
			v-for="milestone in visibleMilestones"
			:key="milestone.id"
			class="gantt-milestone"
			:class="{'gantt-milestone--active': milestone.id === selectedMilestoneId}"
			type="button"
			:aria-pressed="milestone.id === selectedMilestoneId"
			:aria-label="$t('project.gantt.milestoneLabel', {milestone: milestone.name})"
			:style="{left: `${getMilestoneX(milestone)}px`}"
			@click="emit('select', milestone.id)"
		>
			<div
				class="gantt-milestone-line"
				:style="{height: `${height}px`, backgroundColor: milestone.hexColor || undefined}"
			/>
			<div
				class="gantt-milestone-label"
				:style="{color: milestone.hexColor || undefined}"
			>
				<span
					class="gantt-milestone-diamond"
					:style="{backgroundColor: milestone.hexColor || undefined}"
				/>
				<span>
					{{ milestone.name }}
				</span>
			</div>
		</button>
	</div>
</template>

<script setup lang="ts">
import {computed} from 'vue'

import type {IMilestone} from '@/modelTypes/IMilestone'

import {MILLISECONDS_A_DAY} from '@/constants/date'
import {roundToNaturalDayBoundary} from '@/helpers/time/roundToNaturalDayBoundary'

const props = defineProps<{
	milestones: IMilestone[]
	dateFromDate: Date
	dateToDate: Date
	dayWidthPixels: number
	height: number
	totalWidth: number
	selectedMilestoneId?: IMilestone['id'] | null
}>()

const emit = defineEmits<{
	(e: 'select', milestoneId: IMilestone['id']): void
}>()

const visibleMilestones = computed(() => {
	return props.milestones.filter(milestone => {
		if (!milestone.milestoneDate) {
			return false
		}

		const milestoneDate = roundToNaturalDayBoundary(milestone.milestoneDate, true)
		return milestoneDate >= props.dateFromDate && milestoneDate <= props.dateToDate
	})
})

function getMilestoneX(milestone: IMilestone) {
	const milestoneDate = roundToNaturalDayBoundary(milestone.milestoneDate, true)
	const diff = Math.ceil((milestoneDate.getTime() - props.dateFromDate.getTime()) / MILLISECONDS_A_DAY)
	return diff * props.dayWidthPixels
}
</script>

<style lang="scss" scoped>
.gantt-milestones {
	position: absolute;
	inset-block-start: 1.9rem;
	inset-inline-start: 0;
	inset-inline-end: 0;
	z-index: 12;
}

.gantt-milestone {
	position: absolute;
	inset-block-start: 0;
	display: flex;
	flex-direction: column;
	align-items: center;
	border: 0;
	background: transparent;
	padding: 0;
	transform: translateX(-50%);
	cursor: pointer;
}

.gantt-milestone--active {
	.gantt-milestone-diamond {
		box-shadow: 0 0 0 4px color-mix(in srgb, var(--white) 40%, transparent);
	}
}

.gantt-milestone-line {
	inline-size: 2px;
	background-color: var(--primary);
	opacity: .7;
}

.gantt-milestone-label {
	position: absolute;
	inset-block-start: -1.1rem;
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	padding: .125rem .5rem;
	background: color-mix(in srgb, var(--white) 90%, transparent);
	box-shadow: var(--shadow-sm);
	font-size: .75rem;
	font-weight: 600;
	white-space: nowrap;
}

.gantt-milestone-diamond {
	inline-size: .8rem;
	block-size: 0.8rem;
	border: 2px solid var(--white);
	border-radius: 2px;
	background: var(--primary);
	box-shadow: var(--shadow-sm);
	transform: rotate(45deg);
}
</style>
