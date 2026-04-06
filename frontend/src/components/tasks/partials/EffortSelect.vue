<template>
	<div class="select">
		<select
			v-model.number="effort"
			:disabled="disabled || undefined"
		>
			<option
				v-for="option in EFFORT_OPTIONS"
				:key="option"
				:value="option"
			>
				{{ option }}
			</option>
		</select>
	</div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
	disabled?: boolean
}>(), {
	disabled: false,
})

const effort = defineModel<number>({ required: true })

const EFFORT_OPTIONS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as const
</script>
