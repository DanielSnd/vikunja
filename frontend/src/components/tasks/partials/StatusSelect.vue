<template>
	<div class="select">
		<select
			v-model="status"
			:disabled="disabled || undefined"
		>
			<option :value="STATUSES.UNSET">
				{{ $t('task.status.unset') }}
			</option>
			<option :value="STATUSES.IN_PROGRESS">
				{{ $t('task.status.in_progress') }}
			</option>
			<option :value="STATUSES.BLOCKED">
				{{ $t('task.status.blocked') }}
			</option>
			<option :value="STATUSES.REVIEW">
				{{ $t('task.status.review') }}
			</option>
			<option :value="STATUSES.DONE">
				{{ $t('task.status.done') }}
			</option>
		</select>
	</div>
</template>

<script setup lang="ts">
import {STATUSES} from '@/constants/priorities'

withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const status = defineModel<number>({
	required: true,
	default: 0,
})

</script>
