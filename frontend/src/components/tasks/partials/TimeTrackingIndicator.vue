<template>
	<span
		v-if="totalSeconds > 0"
		class="time-tracking-indicator"
		:title="tooltip"
	>
		<svg
			class="time-tracking-indicator__svg"
			viewBox="0 0 24 24"
			aria-hidden="true"
		>
			<rect
				class="time-tracking-indicator__fill"
				x="4"
				:y="20 - fillHeight"
				width="16"
				:height="fillHeight"
			/>
			<circle
				class="time-tracking-indicator__outline"
				cx="12"
				cy="12"
				r="8"
			/>
			<path
				class="time-tracking-indicator__hands"
				d="M12 8v4l2.75 1.75"
			/>
		</svg>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {formatDuration} from '@/helpers/time/formatDuration'

const props = defineProps<{
	totalSeconds: number
}>()

const fillRatio = computed(() => {
	return Math.max(0, Math.min(1, 1 - Math.exp(-props.totalSeconds / 14400)))
})

const fillHeight = computed(() => fillRatio.value * 16)

const tooltip = computed(() => formatDuration(props.totalSeconds))
</script>

<style lang="scss" scoped>
.time-tracking-indicator {
	display: inline-flex;
	inline-size: 1rem;
	block-size: 1rem;
	align-items: center;
	justify-content: center;
}

.time-tracking-indicator__svg {
	inline-size: 100%;
	block-size: 100%;
	overflow: visible;
}

.time-tracking-indicator__fill {
	fill: var(--warning);
	clip-path: circle(8px at 12px 12px);
}

.time-tracking-indicator__outline {
	fill: none;
	stroke: currentColor;
	stroke-width: 2;
}

.time-tracking-indicator__hands {
	fill: none;
	stroke: currentColor;
	stroke-linecap: round;
	stroke-linejoin: round;
	stroke-width: 2;
}
</style>
