<template>
	<div
		v-if="isReview"
		class="is-review"
		:class="{ 'is-review--small': variant === 'small' }"
	>
		{{ $t('task.attributes.status_review') }}
	</div>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{
	isReview?: boolean
	variant?: 'default' | 'small'
}>(), {
	isReview: false,
	variant: 'default',
})
</script>


<style lang="scss" scoped>
.is-review {
  background: rgb(50, 112, 114);
  color: var(--white);
  padding: .5rem;
  font-weight: bold;
  line-height: 1;
  border-radius: 4px;
  text-align: center;
  margin: 0px 10px;
}

.is-review--small {
	padding: .2rem .3rem;
}
</style>
