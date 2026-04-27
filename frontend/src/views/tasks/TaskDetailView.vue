<template>
	<div
		ref="taskViewContainer"
		class="loader-container task-view-container"
		:class="{
			'is-loading': isInitialLoading,
			'is-modal': isModal,
			'is-sidebar': isSidebar,
		}"
	>
		<!-- Removing everything until the task is loaded to prevent empty initialization of other components -->
		<div
			v-if="visible"
			class="task-view"
		>
			<BaseButton
				v-if="displayMode === 'page'"
				class="back-button mbs-2"
				@click="lastProject ? router.back() : router.push(projectRoute)"
			>
				<Icon icon="arrow-left" />
				{{ $t('task.detail.back') }}
			</BaseButton>
			<!-- <Heading
				ref="heading"
				:task="task"
				:has-close="displayMode !== 'page'"
				@close="$emit('close')"
			/> -->
			<div
				v-if="project?.id || canWrite"
				class="task-detail-header"
			>
				<h6
					v-if="project?.id"
					class="subtitle"
				>
					<template
						v-for="p in projectStore.getAncestors(project)"
						:key="p.id"
					>
						<a
							v-if="router.options.history.state?.back?.includes('/projects/'+p.id+'/') || false"
							v-shortcut="p.id === project?.id ? 'KeyU' : ''"
							@click="router.back()"
						>
							{{ getProjectTitle(p) }}
						</a>
						<RouterLink
							v-else
							v-shortcut="p.id === project?.id ? 'KeyU' : ''"
							:to="{ name: 'project.index', params: { projectId: p.id } }"
						>
							{{ getProjectTitle(p) }}
						</RouterLink>
						<span
							v-if="p.id !== project?.id"
							class="has-text-grey-light"
						> &gt; </span>
					</template>
					<BucketSelect
						:task="task"
						:can-write="canWrite"
						@update:task="Object.assign(task, $event)"
					/>
				</h6>

				<div
					v-if="canWrite"
					class="task-detail-header-actions d-print-none"
				>
					<BaseButton
						v-tooltip="task.status == STATUSES.UNSET ? $t('task.detail.started') : $t('task.detail.unset')"
						:class="{'is-pending': task.status != STATUSES.UNSET}"
						class="task-header-action-button button--mark-started"
						:aria-label="task.status == STATUSES.UNSET ? $t('task.detail.started') : $t('task.detail.unset')"
						@click="toggleTaskStarted()"
					>
						<Icon icon="play" />
					</BaseButton>
					<BaseButton
						v-tooltip="task.status != STATUSES.REVIEW ? $t('task.detail.review') : $t('task.detail.started')"
						:class="{'is-pending': task.status != STATUSES.REVIEW}"
						class="task-header-action-button button--mark-review"
						:aria-label="task.status != STATUSES.REVIEW ? $t('task.detail.review') : $t('task.detail.started')"
						@click="toggleTaskReview()"
					>
						<Icon icon="circle-exclamation" />
					</BaseButton>
					<BaseButton
						v-tooltip="task.status != STATUSES.BLOCKED ? $t('task.detail.blocked') : $t('task.detail.started')"
						:class="{'is-pending': task.status != STATUSES.BLOCKED}"
						class="task-header-action-button button--mark-blocked"
						:aria-label="task.status != STATUSES.BLOCKED ? $t('task.detail.blocked') : $t('task.detail.started')"
						@click="toggleTaskBlocked()"
					>
						<Icon icon="circle-exclamation" />
					</BaseButton>
					<BaseButton
						v-shortcut="'KeyT'"
						v-tooltip="task.done ? $t('task.detail.undone') : $t('task.detail.done')"
						:class="{'is-pending': !task.done}"
						class="task-header-action-button button--mark-done"
						:aria-label="task.done ? $t('task.detail.undone') : $t('task.detail.done')"
						@click="toggleTaskDone()"
					>
						<Icon icon="check-double" />
					</BaseButton>
					<TaskSubscription
						entity="task"
						:entity-id="task.id"
						:model-value="task.subscription"
						icon-only
						@update:modelValue="sub => task.subscription = sub"
					/>
					<BaseButton
						v-tooltip="taskTimerButtonLabel"
						class="task-header-action-button"
						:aria-label="taskTimerButtonLabel"
						:disabled="isAnotherTaskTimerRunning"
						@click="toggleTaskTimer()"
					>
						<Icon :icon="isTaskTimerRunning ? 'stop' : ['far', 'clock']" />
					</BaseButton>
					<BaseButton
						v-tooltip="$t('task.detail.actions.more')"
						class="task-header-action-button"
						:aria-label="$t('task.detail.actions.more')"
						:aria-expanded="showActionSidebar"
						@click="showActionSidebar = !showActionSidebar"
					>
						<Icon icon="ellipsis-h" />
					</BaseButton>
				</div>
			</div>

			<!-- Content and buttons -->
			<div class="columns mbs-2">
				<!-- Content -->
				<div class="column detail-content">
					<!-- Description -->
					<div
						class="details content description"
						style="padding: 0px 25px 20px 25px; background-color: rgba(51, 51, 68, 0.267); border-radius: 14px; min-height:500px; margin-bottom:20px;
						block-size: auto;
						border-color: var(--grey-200);
						box-shadow: var(--shadow-sm);"
					>
						<Description
							ref="descriptionRef"
							:model-value="task"
							:can-write="canWrite"
							:attachment-upload="attachmentUpload"
							@update:modelValue="Object.assign(task, $event)"
						/>
					</div>
					
					<template v-if="showTaskMetadata">
						<div class="columns details">
							<div class="column">
								<!-- Reactions -->
								<Reactions 
									v-model="task.reactions" 
									entity-kind="tasks"
									:entity-id="task.id"
									class="details"
									:disabled="!canWrite"
								/>
							</div>
							<div
								class="column"
							>
								<!-- Effort -->
								<div class="detail-title">
									<Icon icon="gear" />
									{{ $t('task.attributes.effort') }}
								</div>
								<EffortSelect
									:ref="e => setFieldRef('effort', e)"
									v-model="task.effort"
									:disabled="!canWrite"
									@update:modelValue="setEffort"
								/>
							</div>
							<div
								class="column labels-column"
							>
								<div class="detail-title">
									<span class="icon is-grey">
										<Icon icon="tags" />
									</span>
									{{ $t('task.attributes.labels') }}
								</div>
								<EditLabels
									:ref="e => setFieldRef('labels', e)"
									v-model="task.labels"
									:disabled="!canWrite"
									:task-id="taskId"
									:creatable="!authStore.isLinkShareAuth"
								/>
							</div>
						</div>

						<ChecklistSummary :task="task" />

						<div class="columns details">
							<div
								v-if="activeFields.assignees"
								class="column assignees"
							>
								<!-- Assignees -->
								<div class="detail-title">
									<Icon icon="users" />
									{{ $t('task.attributes.assignees') }}
								</div>
								<EditAssignees
									v-if="canWrite"
									:ref="e => setFieldRef('assignees', e)"
									v-model="task.assignees"
									:project-id="task.projectId"
									:task-id="task.id"
								/>
								<AssigneeList
									v-else
									:assignees="task.assignees"
									class="mbs-2"
								/>
							</div>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.milestone"
									class="column"
								>
									<div class="detail-title">
										<Icon icon="flag-checkered" />
										{{ $t('task.attributes.milestone') }}
									</div>
									<EditMilestone
										:ref="e => setFieldRef('milestone', e)"
										v-model="task.milestone"
										:project-id="task.projectId"
										:disabled="!canWrite"
										@update:modelValue="setMilestone"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.priority"
									class="column"
								>
									<!-- Priority -->
									<div class="detail-title">
										<Icon icon="exclamation-circle" />
										{{ $t('task.attributes.priority') }}
									</div>
									<PrioritySelect
										:ref="e => setFieldRef('priority', e)"
										v-model="task.priority"
										:disabled="!canWrite"
										@update:modelValue="setPriority"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.status"
									class="column"
								>
									<!-- Status -->
									<div class="detail-title">
										<Icon icon="exclamation-circle" />
										{{ $t('task.attributes.status') }}
									</div>
									<StatusSelect
										:ref="e => setFieldRef('status', e)"
										v-model="task.status"
										:disabled="!canWrite"
										@update:modelValue="setStatus"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.dueDate"
									class="column"
								>
									<!-- Due Date -->
									<div class="detail-title">
										<Icon icon="calendar" />
										{{ $t('task.attributes.dueDate') }}
									</div>
									<div class="date-input">
										<Datepicker
											:ref="e => setFieldRef('dueDate', e)"
											v-model="task.dueDate"
											:choose-date-label="$t('task.detail.chooseDueDate')"
											:disabled="isInitialLoading || !canWrite"
											@closeOnChange="saveTask()"
										/>
										<BaseButton
											v-if="task.dueDate && canWrite"
											class="remove"
											@click="() => {task.dueDate = null;saveTask()}"
										>
											<span class="icon is-small">
												<Icon icon="times" />
											</span>
										</BaseButton>
									</div>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.percentDone"
									class="column"
								>
									<!-- Progress -->
									<div class="detail-title">
										<Icon icon="percent" />
										{{ $t('task.attributes.percentDone') }}
									</div>
									<PercentDoneSelect
										:ref="e => setFieldRef('percentDone', e)"
										v-model="task.percentDone"
										:disabled="!canWrite"
										@update:modelValue="setPercentDone"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.startDate"
									class="column"
								>
									<!-- Start Date -->
									<div class="detail-title">
										<Icon icon="play" />
										{{ $t('task.attributes.startDate') }}
									</div>
									<div class="date-input">
										<Datepicker
											:ref="e => setFieldRef('startDate', e)"
											v-model="task.startDate"
											:choose-date-label="$t('task.detail.chooseStartDate')"
											:disabled="isInitialLoading || !canWrite"
											@closeOnChange="saveTask()"
										/>
										<BaseButton
											v-if="task.startDate && canWrite"
											class="remove"
											@click="() => {task.startDate = null;saveTask()}"
										>
											<span class="icon is-small">
												<Icon icon="times" />
											</span>
										</BaseButton>
									</div>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.endDate"
									class="column"
								>
									<!-- End Date -->
									<div class="detail-title">
										<Icon icon="stop" />
										{{ $t('task.attributes.endDate') }}
									</div>
									<div class="date-input">
										<Datepicker
											:ref="e => setFieldRef('endDate', e)"
											v-model="task.endDate"
											:choose-date-label="$t('task.detail.chooseEndDate')"
											:disabled="isInitialLoading || !canWrite"
											@closeOnChange="saveTask()"
										/>
										<BaseButton
											v-if="task.endDate && canWrite"
											class="remove"
											@click="() => {task.endDate = null;saveTask()}"
										>
											<span class="icon is-small">
												<Icon icon="times" />
											</span>
										</BaseButton>
									</div>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.reminders"
									class="column"
								>
									<!-- Reminders -->
									<div class="detail-title">
										<Icon :icon="['far', 'clock']" />
										{{ $t('task.attributes.reminders') }}
									</div>
									<Reminders
										:ref="e => setFieldRef('reminders', e)"
										v-model="task"
										:disabled="!canWrite"
										@update:modelValue="saveTask()"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.repeatAfter"
									class="column"
								>
									<!-- Repeat after -->
									<div class="is-flex is-justify-content-space-between">
										<div class="detail-title">
											<Icon icon="history" />
											{{ $t('task.attributes.repeat') }}
										</div>
										<BaseButton
											v-if="canWrite"
											class="remove"
											@click="removeRepeatAfter"
										>
											<span class="icon is-small">
												<Icon icon="times" />
											</span>
										</BaseButton>
									</div>
									<RepeatAfter
										:ref="e => setFieldRef('repeatAfter', e)"
										v-model="task"
										:disabled="!canWrite"
										@update:modelValue="saveTask()"
									/>
								</div>
							</CustomTransition>
							<CustomTransition
								name="flash-background"
								appear
							>
								<div
									v-if="activeFields.color"
									class="column"
								>
									<!-- Color -->
									<div class="detail-title">
										<Icon icon="fill-drip" />
										{{ $t('task.attributes.color') }}
									</div>
									<ColorPicker
										:ref="e => setFieldRef('color', e)"
										v-model="taskColor"
										menu-position="bottom"
										@update:modelValue="saveTask()"
									/>
								</div>
							</CustomTransition>
						</div>

						<!-- Attachments -->
						<div
							v-if="hasAttachments"
							class="content details attachments-section"
						>
							<BaseButton
								class="foldout-toggle"
								:aria-expanded="showAttachments"
								@click="showAttachments = !showAttachments"
							>
								<span class="foldout-toggle-label">
									<span class="icon is-grey">
										<Icon icon="paperclip" />
									</span>
									{{ $t('task.attachment.title') }} ({{ attachmentCount }})
								</span>
								<Icon :icon="showAttachments ? 'chevron-up' : 'chevron-down'" />
							</BaseButton>

							<Attachments
								v-if="showAttachments"
								:ref="e => { setFieldRef('attachments', e); attachmentsRef = e as any }"
								:edit-enabled="canWrite"
								:show-insert-into-description="canWrite"
								:task="task"
								@taskChanged="({coverImageAttachmentId}) => task.coverImageAttachmentId = coverImageAttachmentId"
								@insertImage="insertAttachmentIntoDescription"
								@update:attachments="onAttachmentsUpdated"
							/>
						</div>

						<!-- Related Tasks -->
						<div
							v-if="activeFields.relatedTasks"
							class="content details mbe-0"
						>
							<h3>
								<span class="icon is-grey">
									<Icon icon="sitemap" />
								</span>
								{{ $t('task.attributes.relatedTasks') }}
							</h3>
							<RelatedTasks
								:ref="e => setFieldRef('relatedTasks', e)"
								:edit-enabled="canWrite"
								:initial-related-tasks="task.relatedTasks"
								:project-id="task.projectId"
								:show-no-relations-notice="true"
								:task-id="taskId"
							/>
						</div>

						<!-- Move Task -->
						<div
							v-if="activeFields.moveProject"
							class="content details"
						>
							<h3>
								<span class="icon is-grey">
									<Icon icon="list" />
								</span>
								{{ $t('task.detail.move') }}
							</h3>
							<div class="field has-addons">
								<div class="control is-expanded">
									<ProjectSearch
										:ref="e => setFieldRef('moveProject', e)"
										:filter="project => project.id !== task.projectId"
										@update:modelValue="changeProject"
									/>
								</div>
							</div>
						</div>
					</template>

					<!-- Time Tracking -->
					<div
						ref="timeTrackingSection"
						class="content details comments-section foldout-section"
						:class="{'foldout-section--open': showTimeTracking}"
					>
						<BaseButton
							class="foldout-toggle"
							:class="{'foldout-toggle--open': showTimeTracking}"
							:aria-expanded="showTimeTracking"
							@click="showTimeTracking = !showTimeTracking"
						>
							<span class="foldout-toggle-label">
								<span class="icon is-grey">
									<Icon :icon="['far', 'clock']" />
								</span>
								{{ showTimeTracking ? $t('task.timeTracking.hide') : $t('task.timeTracking.show') }}
							</span>
							<span class="foldout-toggle-meta">
								<span
									v-if="formattedTimeTrackingTotal"
									class="foldout-toggle-total"
								>
									{{ formattedTimeTrackingTotal }}
								</span>
								<Icon :icon="showTimeTracking ? 'chevron-up' : 'chevron-down'" />
							</span>
						</BaseButton>

						<div
							v-if="showTimeTracking"
							class="foldout-body"
						>
							<TimeTracking
								:task-id="taskId"
								:task-title="task.title"
								:can-write="canWrite"
								@summaryChanged="updateTimeTrackingSummary"
							/>
						</div>
					</div>

					<!-- Comments -->
					<div
						class="content details comments-section foldout-section"
						:class="{'foldout-section--open': showComments}"
					>
						<BaseButton
							class="foldout-toggle"
							:class="{'foldout-toggle--open': showComments}"
							:aria-expanded="showComments"
							@click="showComments = !showComments"
						>
							<span class="foldout-toggle-label">
								<span class="icon is-grey">
									<Icon :icon="['far', 'comments']" />
								</span>
								{{ showComments ? $t('task.comment.hide', {count: commentCount}) : $t('task.comment.show', {count: commentCount}) }}
							</span>
							<Icon :icon="showComments ? 'chevron-up' : 'chevron-down'" />
						</BaseButton>

						<div
							v-if="showComments"
							class="foldout-body"
						>
							<Comments
								:can-write="canWrite"
								:task-id="taskId"
								:project-id="task.projectId"
								:initial-comments="task.comments"
								@countChanged="updateCommentCount"
							/>
						</div>
					</div>

					<!-- Marker element for scroll-to-bottom button visibility -->
					<div
						ref="contentBottomMarker"
						class="content-bottom-marker"
					/>
				</div>
				
				<!-- Task Actions -->
				<div
					v-if="showActionSidebar && (canWrite || isModal)"
					class="column action-buttons d-print-none"
				>
					<template v-if="canWrite">
						<XButton
							v-shortcut="'KeyS'"
							v-tooltip="task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')"
							variant="secondary"
							:aria-label="task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')"
							:icon="task.isFavorite ? 'star' : ['far', 'star']"
							@click="toggleFavorite"
						>
							{{
								task.isFavorite ? $t('task.detail.actions.unfavorite') : $t('task.detail.actions.favorite')
							}}
						</XButton>
						
						<span class="action-heading">{{ $t('task.detail.organization') }}</span>
						
						<XButton
							v-shortcut="'KeyL'"
							v-tooltip="$t('task.detail.actions.label')"
							class="action-group-start"
							variant="secondary"
							:aria-label="$t('task.detail.actions.label')"
							icon="tags"
							@click="setFieldActive('labels')"
						>
							{{ $t('task.detail.actions.label') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.milestone')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.milestone')"
							icon="bullseye"
							@click="setFieldActive('milestone')"
						>
							{{ $t('task.detail.actions.milestone') }}
						</XButton>
						<XButton
							v-shortcut="'KeyP'"
							v-tooltip="$t('task.detail.actions.priority')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.priority')"
							icon="exclamation-circle"
							@click="setFieldActive('priority')"
						>
							{{ $t('task.detail.actions.priority') }}
						</XButton>
						<XButton
							v-shortcut="'KeyS'"
							v-tooltip="$t('task.detail.actions.status')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.status')"
							icon="exclamation-circle"
							@click="setFieldActive('status')"
						>
							{{ $t('task.detail.actions.status') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.percentDone')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.percentDone')"
							icon="percent"
							@click="setFieldActive('percentDone')"
						>
							{{ $t('task.detail.actions.percentDone') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.effort')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.effort')"
							icon="gear"
							@click="setFieldActive('effort')"
						>
							{{ $t('task.detail.actions.effort') }}
						</XButton>
						<XButton
							v-shortcut="'KeyC'"
							v-tooltip="$t('task.detail.actions.color')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.color')"
							icon="fill-drip"
							:icon-color="color"
							@click="setFieldActive('color')"
						>
							{{ $t('task.detail.actions.color') }}
						</XButton>
						
						<span class="action-heading">{{ $t('task.detail.management') }}</span>

						<XButton
							v-shortcut="'KeyA'"
							v-cy="'taskDetail.assign'"
							v-tooltip="$t('task.detail.actions.assign')"
							class="action-group-start"
							variant="secondary"
							:aria-label="$t('task.detail.actions.assign')"
							icon="users"
							@click="setFieldActive('assignees')"
						>
							{{ $t('task.detail.actions.assign') }}
						</XButton>
						<XButton
							v-shortcut="'KeyF'"
							v-tooltip="$t('task.detail.actions.attachments')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.attachments')"
							icon="paperclip"
							@click="openAttachments()"
						>
							{{ $t('task.detail.actions.attachments') }}
						</XButton>
						<XButton
							v-tooltip="taskTimerButtonLabel"
							variant="secondary"
							:aria-label="taskTimerButtonLabel"
							:icon="isTaskTimerRunning ? 'stop' : ['far', 'clock']"
							:disabled="isAnotherTaskTimerRunning"
							@click="toggleTaskTimer()"
						>
							{{ taskTimerButtonLabel }}
						</XButton>
						<XButton
							v-tooltip="$t('task.timeTracking.open')"
							variant="secondary"
							:aria-label="$t('task.timeTracking.open')"
							:icon="['far', 'clock']"
							@click="openTimeTracking()"
						>
							{{ $t('task.timeTracking.open') }}
						</XButton>
						<XButton
							v-shortcut="'KeyR'"
							v-tooltip="$t('task.detail.actions.relatedTasks')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.relatedTasks')"
							icon="sitemap"
							@click="setRelatedTasksActive()"
						>
							{{ $t('task.detail.actions.relatedTasks') }}
						</XButton>
						<XButton
							v-shortcut="'KeyM'"
							v-tooltip="$t('task.detail.actions.moveProject')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.moveProject')"
							icon="list"
							@click="setFieldActive('moveProject')"
						>
							{{ $t('task.detail.actions.moveProject') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.duplicate')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.duplicate')"
							icon="copy"
							@click="duplicateCurrentTask"
						>
							{{ $t('task.detail.actions.duplicate') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.visionBoard')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.visionBoard')"
							icon="object-group"
							@click="openVisionBoard"
						>
							{{ $t('task.detail.actions.visionBoard') }}
						</XButton>

						<span class="action-heading">{{ $t('task.detail.dateAndTime') }}</span>

						<XButton
							v-shortcut="'KeyD'"
							v-tooltip="$t('task.detail.actions.dueDate')"
							class="action-group-start"
							variant="secondary"
							:aria-label="$t('task.detail.actions.dueDate')"
							icon="calendar"
							@click="setFieldActive('dueDate')"
						>
							{{ $t('task.detail.actions.dueDate') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.startDate')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.startDate')"
							icon="play"
							@click="setFieldActive('startDate')"
						>
							{{ $t('task.detail.actions.startDate') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.endDate')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.endDate')"
							icon="stop"
							@click="setFieldActive('endDate')"
						>
							{{ $t('task.detail.actions.endDate') }}
						</XButton>
						<XButton
							v-shortcut="reminderShortcut"
							v-tooltip="$t('task.detail.actions.reminders')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.reminders')"
							:icon="['far', 'clock']"
							@click="setFieldActive('reminders')"
						>
							{{ $t('task.detail.actions.reminders') }}
						</XButton>
						<XButton
							v-tooltip="$t('task.detail.actions.repeatAfter')"
							variant="secondary"
							:aria-label="$t('task.detail.actions.repeatAfter')"
							icon="history"
							@click="setFieldActive('repeatAfter')"
						>
							{{ $t('task.detail.actions.repeatAfter') }}
						</XButton>
						<XButton
							v-shortcut="'Shift+Delete'"
							v-tooltip="$t('task.detail.actions.delete')"
							icon="trash-alt"
							:shadow="false"
							class="is-danger is-outlined has-no-border"
							:aria-label="$t('task.detail.actions.delete')"
							@click="showDeleteModal = true"
						>
							{{ $t('task.detail.actions.delete') }}
						</XButton>
					</template>
				</div>
			</div>
			<!-- Created / Updated [by] -->
			<CreatedUpdated :task="task" />
		</div>

		<BaseButton
			v-if="showScrollToCommentsButton"
			v-tooltip="$t('task.detail.scrollToBottom')"
			class="scroll-to-comments-button d-print-none"
			:aria-label="$t('task.detail.scrollToBottom')"
			@click="scrollToBottom"
		>
			<Icon icon="chevron-down" />
		</BaseButton>

		<Modal
			:enabled="showDeleteModal"
			@close="showDeleteModal = false"
			@submit="deleteTask()"
		>
			<template #header>
				<span>{{ $t('task.detail.delete.header') }}</span>
			</template>

			<template #text>
				<p class="tw:text-balance">
					{{ $t('task.detail.delete.text1') }}
				</p>
				<p class="tw:text-balance">
					{{ $t('task.detail.delete.text2') }}
				</p>
			</template>
		</Modal>
	</div>
</template>

<script lang="ts" setup>
import {ref, reactive, shallowReactive, computed, watch, nextTick, onMounted, onUnmounted} from 'vue'
import {useRouter, useRoute, type RouteLocation, onBeforeRouteLeave} from 'vue-router'
import {useI18n} from 'vue-i18n'
import {unrefElement, useDebounceFn, useElementSize, useIntersectionObserver, useMutationObserver} from '@vueuse/core'
import {klona} from 'klona/lite'

import TaskService from '@/services/task'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {IProject} from '@/modelTypes/IProject'
import type {IMilestone} from '@/modelTypes/IMilestone'

import {PRIORITIES, type Priority} from '@/constants/priorities'
import {STATUSES, type Status} from '@/constants/priorities'
import {PERMISSIONS} from '@/constants/permissions'

import BaseButton from '@/components/base/BaseButton.vue'

// partials
import Attachments from '@/components/tasks/partials/Attachments.vue'
import ChecklistSummary from '@/components/tasks/partials/ChecklistSummary.vue'
import ColorPicker from '@/components/input/ColorPicker.vue'
import Comments from '@/components/tasks/partials/Comments.vue'
import CreatedUpdated from '@/components/tasks/partials/CreatedUpdated.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import Description from '@/components/tasks/partials/Description.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import EditMilestone from '@/components/tasks/partials/EditMilestone.vue'
import ProjectSearch from '@/components/tasks/partials/ProjectSearch.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import EffortSelect from '@/components/tasks/partials/EffortSelect.vue'
import PrioritySelect from '@/components/tasks/partials/PrioritySelect.vue'
import StatusSelect from '@/components/tasks/partials/StatusSelect.vue'
import RelatedTasks from '@/components/tasks/partials/RelatedTasks.vue'
import Reminders from '@/components/tasks/partials/Reminders.vue'
import RepeatAfter from '@/components/tasks/partials/RepeatAfter.vue'
import TimeTracking from '@/components/tasks/partials/TimeTracking.vue'
import TaskSubscription from '@/components/misc/Subscription.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import AssigneeList from '@/components/tasks/partials/AssigneeList.vue'
import BucketSelect from '@/components/tasks/partials/BucketSelect.vue'
import Reactions from '@/components/input/Reactions.vue'

import {generateAttachmentUrl, uploadFile} from '@/helpers/attachments'
import {getProjectTitle} from '@/helpers/getProjectTitle'
import {isAppleDevice} from '@/helpers/isAppleDevice'
import {scrollIntoView} from '@/helpers/scrollIntoView'
import {formatDuration} from '@/helpers/time/formatDuration'
import {TASK_REPEAT_MODES} from '@/types/IRepeatMode'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '@/types/IReminderPeriodRelativeTo'
import {playPopSound} from '@/helpers/playPop'

import {useTaskStore} from '@/stores/tasks'
import {useKanbanStore} from '@/stores/kanban'
import {useProjectStore} from '@/stores/projects'
import {useAuthStore} from '@/stores/auth'
import {useBaseStore} from '@/stores/base'
import {useTaskTimerStore} from '@/stores/taskTimer'

import {useTitle} from '@/composables/useTitle'
import {useTaskDetailShortcuts} from '@/composables/useTaskDetailShortcuts'
import {useWebSocket} from '@/composables/useWebSocket'

import {success} from '@/message'
import type {Action as MessageAction} from '@/message'
import ProjectService from '@/services/project'
import VisionBoardService from '@/services/visionBoard'
import VisionBoardModel from '@/models/visionBoard'

const props = defineProps<{
	taskId: ITask['id'],
	backdropView?: RouteLocation['fullPath'],
	displayMode?: 'page' | 'modal' | 'sidebar',
}>()

const emit = defineEmits<{
	'close': [],
	'taskDeleted': [task: ITask],
	'taskDuplicated': [task: ITask],
}>()

const router = useRouter()
const route = useRoute()
const {t} = useI18n({useScope: 'global'})

const projectStore = useProjectStore()
const taskStore = useTaskStore()
const kanbanStore = useKanbanStore()
const authStore = useAuthStore()
const baseStore = useBaseStore()
const taskTimerStore = useTaskTimerStore()
const {subscribe, connected: wsConnected} = useWebSocket()

const task = ref<ITask>(new TaskModel())
const showComments = ref(false)
const showAttachments = ref(false)
const showTimeTracking = ref(false)
const showActionSidebar = ref(false)
const commentCount = ref(0)
const hasAttachments = computed(() => (task.value.attachments?.length ?? 0) > 0)
const remindersDefaultRelativeTo = computed(() => {
	if (task.value.dueDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE
	}
	if (task.value.startDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.STARTDATE
	}
	if (task.value.endDate) {
		return REMINDER_PERIOD_RELATIVE_TO_TYPES.ENDDATE
	}
	return null
})
const attachmentCount = computed(() => task.value.attachments?.length ?? 0)
const taskNotFound = ref(false)
const taskTitle = computed(() => task.value.title)
useTitle(taskTitle)

const lastProject = computed(() => {
	const backRoute = router.options.history.state?.back
	if (!backRoute || typeof backRoute !== 'string') {
		return null
	}

	const projectMatch = backRoute.match(/\/projects\/(-?\d+)/)
	if (!projectMatch || !projectMatch[1]) {
		return null
	}

	const id = parseInt(projectMatch[1])

	return projectStore.projects[id] ?? null
})

const lastProjectOrTaskProject = computed(() => lastProject.value ?? project.value)

// Use Shift+R on macOS (Alt+R produces special characters depending on keyboard layout)
// Use Alt+r on other platforms
const reminderShortcut = computed(() => isAppleDevice() ? 'Shift+KeyR' : 'Alt+KeyR')

// Match native OS conventions for "delete the selected item"
const deleteShortcut = isAppleDevice() ? 'Backspace' : 'Delete'

onBeforeRouteLeave(async () => {
	if (taskNotFound.value) {
		return
	}

	if (!lastProjectOrTaskProject.value) {
		await new Promise<void>((resolve) => {
			const timeout = setTimeout(() => {
				stop()
				resolve()
			}, 5000) // 5 second timeout
			
			const stop = watch(lastProjectOrTaskProject, (p) => {
				if (p) {
					clearTimeout(timeout)
					stop()
					resolve()
				}
			})
		})
	}

	if (lastProjectOrTaskProject.value) {
		await baseStore.handleSetCurrentProjectIfNotSet(lastProjectOrTaskProject.value)
	}
})

// We doubled the task color property here because verte does not have a real change property, leading
// to the color property change being triggered when the # is removed from it, leading to an update,
// which leads in turn to a change... This creates an infinite loop in which the task is updated, changed,
// updated, changed, updated and so on.
// To prevent this, we put the task color property in a separate value which is set to the task color
// when it is saved and loaded.
const taskColor = ref<ITask['hexColor']>('')

// Used to avoid flashing of empty elements if the task content is not yet loaded.
const visible = ref(false)
const isInitialLoading = computed(() => !visible.value || taskService.loading && !task.value.id)
const skipRealtimeReloadUntil = ref(0)

const project = computed(() => projectStore.projects[task.value.projectId])

const projectRoute = computed(() => ({
	name: 'project.index',
	params: {projectId: task.value.projectId},
	hash: route.hash,
}))

const canWrite = computed(() => (
	task.value.maxPermission !== null &&
	task.value.maxPermission > PERMISSIONS.READ
))

const descriptionRef = ref<InstanceType<typeof Description> | null>(null)

const color = computed(() => {
	const color = task.value.getHexColor
		? task.value.getHexColor()
		: undefined

	return color
})

const displayMode = computed(() => props.displayMode ?? (props.backdropView ? 'modal' : 'page'))
const isModal = computed(() => displayMode.value === 'modal')
const isSidebar = computed(() => displayMode.value === 'sidebar')

async function attachmentUpload(file: File, onSuccess?: (url: string) => void) {
	const uploaded = await uploadFile(props.taskId, file, onSuccess)
	if (uploaded.length > 0) {
		onAttachmentsUpdated([...task.value.attachments, ...uploaded])
	}
	return uploaded
}

function onAttachmentsUpdated(attachments: IAttachment[]) {
	task.value.attachments = attachments
	if (attachments.length === 0) {
		showAttachments.value = false
	}
	kanbanStore.setTaskInBucket({
		...task.value,
		attachments,
	})
}

function insertAttachmentIntoDescription(attachment: IAttachment) {
	descriptionRef.value?.insertImage(generateAttachmentUrl(task.value.id, attachment.id))
}

const heading = ref<HTMLElement | null>(null)

async function scrollToHeading() {
	scrollIntoView(unrefElement(heading))
}

const attachmentsRef = ref<InstanceType<typeof Attachments> | null>(null)
const timeTrackingSection = ref<HTMLElement | null>(null)

const taskViewContainer = ref<HTMLElement | null>(null)
const scrollContainer = ref<HTMLElement | null>(null)
const contentBottomMarker = ref<HTMLElement | null>(null)
const bottomMarkerVisible = ref(true)
const isScrollable = ref(false)

function resolveScrollContainer() {
	let el = taskViewContainer.value

	while (el) {
		const overflowY = getComputedStyle(el).overflowY
		if (['auto', 'scroll', 'overlay'].includes(overflowY)) {
			scrollContainer.value = el
			return
		}
		el = el.parentElement
	}

	scrollContainer.value = (document.scrollingElement as HTMLElement | null) ?? document.documentElement
}

function updateScrollable() {
	const scroller = scrollContainer.value
	if (!scroller) {
		isScrollable.value = false
		return
	}

	isScrollable.value = scroller.scrollHeight > scroller.clientHeight + 1
}

const showScrollToCommentsButton = computed(() => {
	return isScrollable.value && !bottomMarkerVisible.value
})

function scrollToBottom() {
	if (!contentBottomMarker.value) {
		return
	}

	contentBottomMarker.value.scrollIntoView({
		behavior: 'smooth',
		block: 'end',
		inline: 'nearest',
	})
}

function updateCommentCount(count: number) {
	commentCount.value = count
	task.value.commentCount = count
}

useIntersectionObserver(
	contentBottomMarker,
	([entry]) => {
		bottomMarkerVisible.value = entry?.isIntersecting ?? true
	},
	{threshold: 0.1},
)

const debouncedMutationHandler = useDebounceFn(async () => {
	await nextTick()
	resolveScrollContainer()
	updateScrollable()
}, 100)

useMutationObserver(
	taskViewContainer,
	debouncedMutationHandler,
	{subtree: true, childList: true},
)

const {height: scrollContainerHeight} = useElementSize(scrollContainer)
watch(scrollContainerHeight, () => updateScrollable())

onMounted(async () => {
	await nextTick()
	resolveScrollContainer()
	updateScrollable()
	await taskTimerStore.loadCurrent()
})

const taskService = shallowReactive(new TaskService())
const visionBoardService = shallowReactive(new VisionBoardService())

const currentTimer = computed(() => taskTimerStore.currentTimer)
const isTaskTimerRunning = computed(() => currentTimer.value?.status === 'running' && currentTimer.value.taskId === task.value.id)
const isAnotherTaskTimerRunning = computed(() => currentTimer.value?.status === 'running' && currentTimer.value.taskId !== task.value.id)
const taskTimerButtonLabel = computed(() => isTaskTimerRunning.value ? t('task.timeTracking.stopTimer') : t('task.timeTracking.startTimer'))
const showTaskMetadata = computed(() => !canWrite.value || showActionSidebar.value)
const formattedTimeTrackingTotal = computed(() => {
	if ((task.value.timeTrackingTotal ?? 0) <= 0) {
		return null
	}

	return formatDuration(task.value.timeTrackingTotal ?? 0)
})

async function loadTask(id: ITask['id']) {
	const loaded = await taskService.get({id}, {expand: ['reactions', 'comments', 'is_unread', 'buckets', 'time_tracking_summary']})
	Object.assign(task.value, loaded)
	taskTimerStore.hydrateCurrentTask({id: loaded.id, title: loaded.title})
	updateCommentCount(loaded.commentCount ?? loaded.comments?.length ?? 0)
	showComments.value = route.hash.startsWith('#comment-')
	showTimeTracking.value = showTimeTracking.value || (loaded.timeTrackingTotal ?? 0) > 0
	taskColor.value = task.value.hexColor
	setActiveFields()

	if (task.value.isUnread) {
		await taskStore.markTaskAsRead(task.value.id)
		task.value.isUnread = false
	}

	if (lastProject.value) {
		await baseStore.handleSetCurrentProjectIfNotSet(lastProject.value)
	}
}

watch(
	() => taskStore.lastUpdatedTask,
	(updatedTask) => {
		if (updatedTask?.id === task.value.id) {
			skipRealtimeReloadUntil.value = Date.now() + 2000
		}
	},
)

// load task
watch(
	() => props.taskId,
	async (id) => {
		if (id === undefined) {
			return
		}

		try {
			await loadTask(id)
		} catch (e) {
			if (e?.response?.status === 404) {
				taskNotFound.value = true
				router.replace({name: 'not-found'})
				return
			}

			throw e
		} finally {
			await nextTick()
			scrollToHeading()
			resolveScrollContainer()
			updateScrollable()
			visible.value = true
		}
	}, {immediate: true})

const taskWsEvent = computed(() => {
	if (task.value.id === 0 || task.value.projectId === 0) {
		return null
	}

	return `project.${task.value.projectId}.task.${task.value.id}.changed`
})
const reloadTaskFromRealtime = useDebounceFn(() => {
	if (props.taskId === undefined || props.taskId === 0) {
		return
	}

	if (Date.now() < skipRealtimeReloadUntil.value) {
		return
	}

	loadTask(props.taskId).catch((e) => {
		console.warn('Failed to reload task from realtime event:', e)
	})
}, 300)

async function toggleTaskTimer() {
	if (isTaskTimerRunning.value) {
		await taskTimerStore.stop(task.value.id)
		success({message: t('task.timeTracking.timerStopped')})
		await loadTask(task.value.id)
		return
	}

	await taskTimerStore.start({id: task.value.id, title: task.value.title})
	showTimeTracking.value = true
	success({message: t('task.timeTracking.timerStarted')})
}

function openTimeTracking() {
	showTimeTracking.value = true
	nextTick(() => {
		if (timeTrackingSection.value) {
			scrollIntoView(timeTrackingSection.value)
		}
	})
}

function updateTimeTrackingSummary({total, summary}: {total: number, summary: ITask['timeTrackingSummary']}) {
	task.value.timeTrackingTotal = total
	task.value.timeTrackingSummary = summary
}

let unsubscribeTaskWs: (() => void) | null = null

watch(taskWsEvent, (eventName) => {
	unsubscribeTaskWs?.()
	unsubscribeTaskWs = null

	if (!eventName) {
		return
	}

	unsubscribeTaskWs = subscribe(eventName, (msg) => {
		if (msg.event === eventName) {
			reloadTaskFromRealtime()
		}
	})
}, {immediate: true})

watch(wsConnected, (isConnected, wasConnected) => {
	if (wasConnected && !isConnected) {
		reloadTaskFromRealtime()
	}
})

onUnmounted(() => {
	unsubscribeTaskWs?.()
})

type FieldType =
	| 'assignees'
	| 'attachments'
	| 'color'
	| 'dueDate'
	| 'endDate'
	| 'labels'
	| 'milestone'
	| 'moveProject'
	| 'percentDone'
	| 'effort'
	| 'priority'
	| 'status'
	| 'relatedTasks'
	| 'reminders'
	| 'repeatAfter'
	| 'startDate'

const activeFields: { [type in FieldType]: boolean } = reactive({
	assignees: false,
	attachments: false,
	color: false,
	dueDate: false,
	endDate: false,
	labels: false,
	milestone: false,
	moveProject: false,
	percentDone: false,
	effort: false,
	priority: false,
	status: false,
	relatedTasks: false,
	reminders: false,
	repeatAfter: false,
	startDate: false,
})

function setActiveFields() {
	// FIXME: are these lines necessary?
	// task.startDate = task.startDate || null
	// task.endDate = task.endDate || null

	// Set all active fields based on values in the model
	activeFields.assignees = task.value.assignees.length > 0
	activeFields.attachments = task.value.attachments.length > 0
	activeFields.dueDate = task.value.dueDate !== null
	activeFields.endDate = task.value.endDate !== null
	activeFields.labels = task.value.labels.length > 0
	activeFields.milestone = task.value.milestone !== null
	activeFields.percentDone = task.value.percentDone > 0
	activeFields.effort = task.value.effort > 0
	activeFields.priority = task.value.priority !== PRIORITIES.UNSET
	activeFields.status = task.value.status !== STATUSES.UNSET
	activeFields.relatedTasks = Object.keys(task.value.relatedTasks).length > 0
	activeFields.reminders = task.value.reminders.length > 0
	activeFields.repeatAfter = task.value.repeatAfter?.amount > 0 || task.value.repeatMode !== TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	activeFields.startDate = task.value.startDate !== null
}

const activeFieldElements: { [id in FieldType]: HTMLElement | null } = reactive({
	assignees: null,
	attachments: null,
	color: null,
	dueDate: null,
	endDate: null,
	labels: null,
	milestone: null,
	moveProject: null,
	percentDone: null,
	effort: null,
	priority: null,
	status: null,
	relatedTasks: null,
	reminders: null,
	repeatAfter: null,
	startDate: null,
})

function setFieldRef(name, e) {
	activeFieldElements[name] = unrefElement(e)
}

function setFieldActive(fieldName: keyof typeof activeFields) {
	activeFields[fieldName] = true
	nextTick(() => {
		const el = activeFieldElements[fieldName]

		if (!el) {
			return
		}

		el.focus()

		// scroll the field to the center of the screen if not in viewport already
		scrollIntoView(el)
	})
}

function openAttachments() {
	showAttachments.value = true
	activeFields.attachments = true
	nextTick(() => {
		const el = activeFieldElements.attachments
		if (el) {
			scrollIntoView(el)
		}
		attachmentsRef.value?.openFilePicker()
	})
}

async function saveTask(
	currentTask: ITask | null = null,
	undoCallback?: () => void,
) {
	if (currentTask === null) {
		currentTask = klona(task.value)
	}

	if (!canWrite.value) {
		return
	}

	currentTask.hexColor = taskColor.value

	// If no end date is being set, but a start date and due date,
	// use the due date as the end date
	if (
		currentTask.endDate === null &&
		currentTask.startDate !== null &&
		currentTask.dueDate !== null
	) {
		currentTask.endDate = currentTask.dueDate
	}

	const updatedTask = await taskStore.update(currentTask) // TODO: markraw ?
	Object.assign(task.value, updatedTask)
	setActiveFields()

	let actions: MessageAction[] = []
	if (undoCallback) {
		actions = [{
			title: t('task.undo'),
			callback: undoCallback,
		}]
	}
	success({message: t('task.detail.updateSuccess')}, actions)
}

useTaskDetailShortcuts({
	task: () => task.value,
	taskTitle: () => taskTitle.value,
	onSave: saveTask,
})

const showDeleteModal = ref(false)

async function deleteTask() {
	await taskStore.delete(task.value)
	success({message: t('task.detail.deleteSuccess')})

	if (displayMode.value === 'sidebar') {
		emit('taskDeleted', task.value)
		return
	}

	router.push({name: 'project.index', params: {projectId: task.value.projectId}})
}

async function toggleTaskDone() {
	const newTask = {
		...task.value,
		done: !task.value.done,
	}

	if (newTask.done) {
		playPopSound()
	}

	await saveTask(
		newTask,
		toggleTaskDone,
	)
}

async function toggleTaskStarted() {
	const newTask = {
		...task.value,
		status: task.value.status === STATUSES.IN_PROGRESS ? STATUSES.UNSET : STATUSES.IN_PROGRESS,
	}

	await saveTask(
		newTask,
		toggleTaskStarted,
	)
}

async function toggleTaskReview() {
	const newTask = {
		...task.value,
		status: task.value.status === STATUSES.REVIEW ? STATUSES.IN_PROGRESS : STATUSES.REVIEW,
	}

	await saveTask(
		newTask,
		toggleTaskStarted,
	)
}

async function toggleTaskBlocked() {
	const newTask = {
		...task.value,
		status: task.value.status === STATUSES.BLOCKED ? STATUSES.IN_PROGRESS : STATUSES.BLOCKED,
	}

	await saveTask(
		newTask,
		toggleTaskBlocked,
	)
}
async function changeProject(project: IProject | null) {
	if (project === null) {
		return
	}
	kanbanStore.removeTaskInBucket(task.value)
	await saveTask({
		...task.value,
		projectId: project.id,
	})
	baseStore.setCurrentProject(project)
}

async function toggleFavorite() {
	const newTask = await taskStore.toggleFavorite(task.value)
	Object.assign(task.value, newTask)
}

async function duplicateCurrentTask() {
	const duplicatedTask = await taskStore.duplicateTask(task.value.id)
	if (duplicatedTask) {
		success({message: t('task.detail.duplicateSuccess')})

		if (displayMode.value === 'sidebar') {
			emit('taskDuplicated', duplicatedTask)
			return
		}

		router.push({
			name: 'task.detail',
			params: {id: duplicatedTask.id},
		})
	}
}

async function openVisionBoard() {
	const boards = await visionBoardService.getAll({
		projectId: task.value.projectId,
	})

	let board = boards.find(existingBoard => existingBoard.taskId === task.value.id) ?? null
	if (board === null) {
		board = await visionBoardService.create(new VisionBoardModel({
			projectId: task.value.projectId,
			taskId: task.value.id,
			title: task.value.title,
		}))
	}

	let projectWithViews = projectStore.projects[task.value.projectId]
	if (!projectWithViews) {
		const projectService = new ProjectService()
		projectWithViews = await projectService.get({id: task.value.projectId})
		projectStore.setProject(projectWithViews)
	}

	const visionBoardView = projectWithViews.views.find(v => v.viewKind === 'vision_board')
	if (!visionBoardView) {
		return
	}

	await router.push({
		name: 'project.view',
		params: {
			projectId: task.value.projectId,
			viewId: visionBoardView.id,
		},
		query: {
			visionBoardId: String(board.id),
		},
	})
}

async function setPriority(priority: Priority) {
	const newTask: ITask = {
		...task.value,
		priority,
	}

	return saveTask(newTask)
}

async function setMilestone(milestone: IMilestone | null) {
	const newTask: ITask = {
		...task.value,
		milestone,
		milestoneId: milestone?.id ?? 0,
	}

	return saveTask(newTask)
}

async function setStatus(status: Status) {
	const newTask: ITask = {
		...task.value,
		status,
	}

	return saveTask(newTask)
}

async function setPercentDone(percentDone: number) {
	const newTask: ITask = {
		...task.value,
		percentDone,
	}

	return saveTask(newTask)
}

async function setEffort(effort: number) {
	const newTask: ITask = {
		...task.value,
		effort,
	}

	return saveTask(newTask)
}

async function removeRepeatAfter() {
	task.value.repeatAfter.amount = 0
	task.value.repeatMode = TASK_REPEAT_MODES.REPEAT_MODE_DEFAULT
	await saveTask()
}

function setRelatedTasksActive() {
	setFieldActive('relatedTasks')

	// If the related tasks are already available, show the form again
	const el = activeFieldElements['relatedTasks']
	if (!el) {
		return
	}
	for (const child of Array.from(el.children)) {
		if ((child as HTMLElement).id === 'showRelatedTasksFormButton') {
			(child as HTMLElement).click()
			break
		}
	}
}
</script>

<style lang="scss" scoped>
.task-view-container {
	// simulate sass lighten($primary, 30) by increasing lightness 30% to 73%
	--primary-light: hsla(var(--primary-h), var(--primary-s), 73%, var(--primary-a));
	padding-block-end: 0;

	@media screen and (min-width: $desktop) {
		padding-block-end: 1rem;
	}
}

.task-view {
	padding-block-start: 1rem;
	padding-inline: .5rem;
	background-color: #212a37;

	@media screen and (min-width: $desktop) {
		padding: 1rem;
	}
}

.is-sidebar .task-view {
	padding: 1rem;
}

.is-modal .task-view {
	border-radius: $radius;
	padding: 1rem;
	color: var(--text);
	background-color: var(--site-background) !important;

	@media screen and (width <= calc(#{$desktop} + 1px)) {
		border-radius: 0;
	}
}

.task-view * {
	transition: opacity 50ms ease;
}

.is-loading .task-view * {
	opacity: 0;
}


.task-detail-header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 1rem;
	flex-wrap: wrap;
}

.subtitle {
	color: var(--grey-500);
	margin-block-end: 1rem;
	flex: 1 1 auto;
	min-inline-size: 0;

	a {
		color: var(--grey-800);
	}
}

.task-detail-header-actions {
	display: inline-flex;
	align-items: center;
	justify-content: flex-end;
	gap: .5rem;
	flex-wrap: wrap;
	margin-block-end: 1rem;
}

.task-header-action-button,
:deep(.task-detail-header-actions > .button),
:deep(.task-detail-header-actions > .base-button) {
	inline-size: 2.75rem;
	block-size: 2.75rem;
	min-block-size: 2.75rem;
	padding: 0;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	border-radius: $radius;
	background-color: transparent;
	color: var(--grey-700);
	border: 1px solid var(--grey-200);
	box-shadow: none;

	&:hover,
	&:focus {
		background-color: var(--grey-100);
		color: var(--grey-900);
	}
}

.task-header-action-button {
	&.button--mark-done {
		&.is-pending {
			color: var(--success);

			&:hover,
			&:focus {
				background-color: var(--success);
				color: #ffffff;
			}
		}
	}

	&.button--mark-started {
		&.is-pending {
			color: rgb(86, 96, 235);

			&:hover,
			&:focus {
				background-color: rgb(86, 96, 235);
				color: #ffffff;
			}
		}
	}

	&.button--mark-review {
		&.is-pending {
			color: rgb(86, 235, 233);

			&:hover,
			&:focus {
				background-color: rgb(86, 220, 235);
				color: #ffffff;
			}
		}
	}

	&.button--mark-blocked {
		&.is-pending {
			color: rgb(231, 99, 78);

			&:hover,
			&:focus {
				background-color: rgb(231, 99, 78);
				color: #ffffff;
			}
		}
	}
}

h3 .button {
	vertical-align: middle;
}

.icon.is-grey {
	color: var(--grey-400);
}

.date-input {
	display: flex;
	align-items: center;
}

.remove {
	color: var(--danger);
	vertical-align: middle;
	padding-inline-start: .5rem;
	line-height: 1;
}

:deep(.datepicker) {
	inline-size: 100%;

	.show {
		color: var(--text);
		padding: .25rem .5rem;
		transition: background-color $transition;
		border-radius: $radius;
		display: block;
		margin: .1rem 0;
		inline-size: 100%;
		text-align: start;

		&:hover {
			background: var(--white);
		}
	}

	&.disabled .show:hover {
		background: transparent;
	}
}

.details {
	padding-block-end: 0.75rem;
	flex-flow: row wrap;
	margin-block-end: 0;

	.detail-title {
		display: block;
		color: var(--grey-400);
	}

	.none {
		font-style: italic;
	}

	// Break after the 2nd element
	.column:nth-child(2n) {
		page-break-after: always; // CSS 2.1 syntax
		break-after: always; // New syntax
	}

}

.details.labels-list,
.assignees {
	:deep(.multiselect) {
		.input-wrapper {
			&:not(:focus-within, :hover) {
				background: transparent;
				border-color: transparent;
			}
		}
	}
}

.labels-column {
	position: relative;
	overflow: visible;
	z-index: 2;

	:deep(.multiselect) {
		overflow: visible;
	}

	:deep(.search-results) {
		z-index: 150;
	}
}

:deep(.details),
:deep(.heading) {
	.input:not(.has-defaults),
	.textarea,
	.select:not(.has-defaults) select {
		cursor: pointer;
		transition: all $transition-duration;

		&::placeholder {
			color: var(--text-light);
			opacity: 1;
			font-style: italic;
		}

		&:not(:disabled) {
			&:hover,
			&:active,
			&:focus {
				background: var(--scheme-main);
				border-color: var(--border);
				cursor: text;
			}

			&:hover,
			&:active {
				cursor: text;
				border-color: var(--link)
			}
		}
	}

	.select:not(.has-defaults):after {
		opacity: 0;
	}

	.select:not(.has-defaults):hover:after {
		opacity: 1;
	}
}

.attachments {
	margin-block-end: 0;

	table tr:last-child td {
		border-inline-end: none;
	}
}

.action-buttons {
	display: flex;
	flex-direction: column;
	gap: .5rem;
	align-content: flex-start;
	align-items: flex-start;

	@media screen and (min-width: $tablet) {
		position: sticky;
		inset-block-start: $navbar-height + 1.5rem;
		align-self: flex-start;
		flex: 0 0 auto;
		inline-size: auto;
	}

	.button {
		inline-size: 2.75rem;
		block-size: 2.75rem;
		min-block-size: 2.75rem;
		padding: 0;
		justify-content: center;
		flex: 0 0 auto;

		&.has-light-text {
			color: var(--white);
		}

		&.button--mark-done {
			background-color: transparent;
			box-shadow: none;

			&.is-pending {
				color: var(--success);

				&:hover,
				&:focus {
					background-color: var(--success);
					color: #ffffff;
				}
			}
		}
		&.button--mark-started {
			background-color: transparent;
			box-shadow: none;

			&.is-pending {
				color: rgb(86, 96, 235);

				&:hover,
				&:focus {
					background-color: rgb(86, 96, 235);
					color: #ffffff;
				}
			}
		}
		&.button--mark-review {
			background-color: transparent;
			box-shadow: none;

			&.is-pending {
				color: rgb(86, 235, 233);

				&:hover,
				&:focus {
					background-color: rgb(86, 220, 235);
					color: #ffffff;
				}
			}
		}
		&.button--mark-blocked {
			background-color: transparent;
			box-shadow: none;

			&.is-pending {
				color: rgb(231, 99, 78);

				&:hover,
				&:focus {
					background-color: rgb(231, 99, 78);
					color: #ffffff;
				}
			}
		}
	}

	:deep(.button > span:not(.icon)) {
		display: none;
	}
}

.is-modal .action-buttons {
	// we need same top margin for the modal close button 
	@media screen and (min-width: $tablet) {
		inset-block-start: 6.5rem;
	}
	// this is the moment when the fixed close button is outside the modal
	// => we can fill up the space again
	@media screen and (width >= calc(#{$desktop} + 84px)) {
		inset-block-start: 0;
	}
}

.checklist-summary {
	padding-inline-start: .25rem;
}

.detail-content {
	@media print {
		inline-size: 100% !important;
	}
}

.comments-section {
	margin-block-end: 0;
}

.foldout-section {
	border-radius: $radius;
	background: var(--white);
	border: 1px solid transparent;
	padding: 0;
}

.foldout-section--open {
	border-color: var(--grey-200);
	box-shadow: var(--shadow-xs);
	overflow: clip;
}

.foldout-toggle {
	inline-size: 100%;
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: .75rem 1rem;
	border: 1px solid var(--grey-200);
	border-radius: $radius;
	background: var(--white);
	color: var(--text);
}

.foldout-toggle--open {
	border: 0;
	border-radius: 0;
	border-block-end: 1px solid var(--grey-200);
	background: linear-gradient(180deg, var(--grey-50) 0%, var(--white) 100%);
}

.foldout-toggle-label {
	display: inline-flex;
	align-items: center;
	gap: .5rem;
}

.foldout-toggle-meta {
	display: inline-flex;
	align-items: center;
	gap: .75rem;
	color: var(--grey-500);
}

.foldout-toggle-total {
	font-size: .9rem;
	font-variant-numeric: tabular-nums;
}

.foldout-body {
	padding: 1rem;
	background: var(--white);
}

.foldout-body :deep(.time-tracking) {
	padding: 0;
	border: 0;
	background: transparent;
}

.action-heading {
	display: none;
}

.action-group-start {
	margin-block-start: .75rem;
}

.scroll-to-comments-button {
	position: fixed;
	// Position above the keyboard shortcuts button (which is at bottom: calc(1rem - 4px))
	inset-block-end: 2.5rem;
	inset-inline-end: .75rem;
	z-index: 10;
	inline-size: 2rem;
	block-size: 2rem;
	border-radius: 100%;
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0;
	background-color: var(--site-background);
	border: 1px solid var(--grey-300);
	color: var(--grey-500);
	box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
	transition: all $transition;

	&:hover {
		background-color: var(--grey-100);
		color: var(--grey-700);
	}

	@media screen and (max-width: $tablet) {
		// Hide on mobile since keyboard shortcuts button is also hidden
		display: none;
	}
}
</style>

<style lang="scss">
// global style to override position when the modal task detail is active
.modal-content .scroll-to-comments-button {
	inset-block-end: .75rem;
	inset-inline-end: 1rem;
}
</style>
