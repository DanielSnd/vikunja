<template>
	<div
		class="gantt-milestones"
		:aria-label="$t('project.gantt.milestonesRow')"
		:style="{width: `${totalWidth}px`}"
	>
		<div
			v-for="milestone in visibleMilestones"
			:key="milestone.id"
			class="gantt-milestone"
			:style="{left: `${getMilestoneX(milestone)}px`}"
		>
			<div
				class="gantt-milestone-line"
				:style="{height: `${height}px`, backgroundColor: milestone.hexColor || undefined}"
			/>
			<div
				class="gantt-milestone-label"
				:style="{borderColor: milestone.hexColor || undefined, color: milestone.hexColor || undefined}"
			>
				<span
					class="gantt-milestone-dot"
					:style="{backgroundColor: milestone.hexColor || undefined}"
				/>
				<span>
					{{ milestone.name }}
				</span>
			</div>
		</div>
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
	inset-block-start: 0;
	inset-inline-start: 0;
	inset-inline-end: 0;
	pointer-events: none;
	z-index: 3;
}

.gantt-milestone {
	position: absolute;
	inset-block-start: 0;
}

.gantt-milestone-line {
	inline-size: 2px;
	background-color: var(--primary);
	opacity: .7;
}

.gantt-milestone-label {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: .5rem;
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	padding: .125rem .5rem;
	border: 1px solid var(--primary);
	border-radius: 999px;
	background: color-mix(in srgb, var(--white) 90%, transparent);
	box-shadow: var(--shadow-sm);
	font-size: .75rem;
	font-weight: 600;
	white-space: nowrap;
}

.gantt-milestone-dot {
	inline-size: .5rem;
	block-size: .5rem;
	border-radius: 50%;
	background: var(--primary);
}
</style>
