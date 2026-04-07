<template>
	<div
		v-if="isBlocked"
		class="is-blocked"
		:class="{ 'is-blocked--small': variant === 'small' }"
	>
		{{ $t('task.attributes.status_blocked') }}
	</div>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{
	isBlocked?: boolean
	variant?: 'default' | 'small'
}>(), {
	isBlocked: false,
	variant: 'default',
})
</script>


<style lang="scss" scoped>
.is-blocked {
  background: var(--warning);
  color: var(--white);
  padding: .5rem;
  font-weight: bold;
  line-height: 1;
  border-radius: 4px;
  text-align: center;
  margin: 0px 10px;
}

.is-blocked--small {
	padding: .2rem .3rem;
}
</style>
