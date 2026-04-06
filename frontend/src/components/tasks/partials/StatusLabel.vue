<template>
	<span
		v-if="!done && (showAll || status >= minimumStatus)"
		:class="{
			'negligible': status <= statuses.UNSET,
			'not-so-high': status == statuses.IN_PROGRESS,
			'high-status': status == statuses.BLOCKED
		}"
		class="status-label"
	>
		<span class="icon">
			<Icon
				v-if="status >= statuses.BLOCKED"
				icon="exclamation-circle"
			/>
			<Icon
				v-else
				icon="exclamation"
			/>
		</span>
		<span>
			<template v-if="status === statuses.UNSET">{{ $t('task.status.unset') }}</template>
			<template v-if="status === statuses.IN_PROGRESS">{{ $t('task.status.in_progress') }}</template>
			<template v-if="status === statuses.BLOCKED">{{ $t('task.status.blocked') }}</template>
			<template v-if="status === statuses.REVIEW">{{ $t('task.status.review') }}</template>
			<template v-if="status === statuses.DONE">{{ $t('task.status.done') }}</template>
		</span>
	</span>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {STATUSES as statuses} from '@/constants/priorities'
import {useAuthStore} from '@/stores/auth'
	
withDefaults(defineProps<{
	status: number,
	showAll?: boolean,
	done?: boolean
}>(), {
	showAll: false,
	done: false,
})

const authStore = useAuthStore()

const minimumStatus = computed(() => {
	return statuses.UNSET
})
</script>

<style lang="scss" scoped>
.high-status {
	color: var(--danger);
	inline-size: auto !important; // To override the width set in tasks
}

.not-so-high {
	color: var(--warning);
}

.negligible {
	color: var(--info);
}

.icon {
	vertical-align: top;
	inline-size: auto !important;
	padding-inline-end: .5rem;
}
</style>
