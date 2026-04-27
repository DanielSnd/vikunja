<template>
	<ProjectWrapper
		class="project-vision-board"
		:is-loading-project="isLoadingProject"
		:project-id="projectId"
		:view-id
	>
		<template #header>
			<div class="vision-board-toolbar">
				<div class="vision-board-toolbar__group">
					<label
						class="label"
						for="vision-board-select"
					>
						{{ $t('project.vision_board.selectLabel') }}
					</label>
					<div class="select">
						<select
							id="vision-board-select"
							:model-value="selectedBoardIdValue"
							:disabled="boards.length === 0"
							@change="onBoardSelect"
						>
							<option value="">
								{{ $t('project.vision_board.selectPlaceholder') }}
							</option>
							<option
								v-for="board in boards"
								:key="board.id"
								:value="board.id"
							>
								{{ board.taskTitle || board.title }}
							</option>
						</select>
					</div>
				</div>

				<div class="vision-board-toolbar__actions">
					<BaseButton @click="showCreateForm = !showCreateForm">
						{{ $t('project.vision_board.create') }}
					</BaseButton>
				</div>
			</div>
		</template>

		<template #default>
			<div
				v-if="showCreateForm || boards.length === 0"
				class="vision-board-create"
			>
				<FormField :label="$t('project.vision_board.createTaskLabel')">
					<Multiselect
						v-model="selectedTask"
						:search-results="foundTasks"
						:loading="taskCollectionService.loading"
						:placeholder="$t('project.vision_board.createTaskPlaceholder')"
						label="title"
						@search="findTasks"
					/>
				</FormField>
				<div class="vision-board-create__actions">
					<BaseButton
						:disabled="selectedTask === null"
						@click="createBoard"
					>
						{{ $t('project.vision_board.create') }}
					</BaseButton>
				</div>
			</div>

			<Message
				v-else-if="activeBoard === null"
				class="vision-board-empty-state"
			>
				{{ boards.length > 0 ? $t('project.vision_board.emptySelection') : $t('project.vision_board.noBoards') }}
			</Message>

			<section
				v-else
				class="vision-board-stage"
				:class="{'vision-board-stage--fullscreen': isFullscreen}"
			>
				<header class="vision-board-stage__header">
					<div>
						<h1 class="title is-5">
							{{ activeBoard.title }}
						</h1>
					</div>

					<div class="vision-board-stage__actions">
						<BaseButton
							v-for="kind in nodeKinds"
							:key="kind"
							@click="() => addNode(kind)"
						>
							{{ $t(`project.vision_board.nodes.${kind}`) }}
						</BaseButton>
					</div>
				</header>

				<div
					ref="viewportRef"
					class="vision-board-stage__viewport"
					:style="{
						height: `${viewportHeight}px`,
						minHeight: `${viewportHeight}px`,
						maxHeight: `${viewportHeight}px`,
					}"
				>
					<div
						ref="canvasRef"
						class="vision-board-stage__canvas"
						:class="{'vision-board-stage__canvas--connecting': pendingEdge !== null}"
						@dblclick="onCanvasDoubleClick"
						@mousedown="handleCanvasMouseDown"
						@mousemove="handleCanvasPresenceMove"
						@mouseleave="handleCanvasPresenceLeave"
					>
						<div
							class="vision-board-stage__surface-frame"
							:style="{						
								inlineSize: `${canvasSize.width}px`,
								blockSize: `${canvasSize.height}px`,
							}"
						>
							<div
								class="vision-board-stage__surface"
								:class="{'is-dragging': draggingNodeId !== null || resizingNodeId !== null}"
								:style="{
									inlineSize: `${canvasSize.width}px`,
									blockSize: `${canvasSize.height}px`,
									transform: `scale(${currentZoom})`,
								}"
							>
								<svg
									class="vision-board-stage__edges"
									:viewBox="`0 0 ${canvasSize.width} ${canvasSize.height}`"
									:width="canvasSize.width"
									:height="canvasSize.height"
									preserveAspectRatio="none"
								>
									<path
										v-for="edge in selectedBoardEdges"
										:key="edge.id"
										:d="getEdgePath(edge)"
										:stroke="edge.color || edgeStrokeColor"
										class="vision-board-stage__edge-path"
										:class="{'vision-board-stage__edge-path--selected': selectedEdgeId === edge.id}"
										marker-end="url(#vision-board-arrow)"
										@click.stop="selectEdge(edge.id)"
									/>
									<path
										v-if="pendingEdge !== null"
										:d="getPendingEdgePath()"
										:stroke="edgeStrokeColor"
										class="vision-board-stage__edge-path vision-board-stage__edge-path--preview"
										marker-end="url(#vision-board-arrow)"
									/>
									<defs>
										<marker
											id="vision-board-arrow"
											viewBox="0 0 10 10"
											refX="8"
											refY="5"
											markerWidth="7"
											markerHeight="7"
											orient="auto"
										>
											<path
												d="M 0 0 L 10 5 L 0 10 z"
												fill="context-stroke"
											/>
										</marker>
									</defs>
								</svg>

								<div
									v-if="selectedEdge !== null"
									class="vision-edge-toolbar"
									:style="getEdgeToolbarStyle(selectedEdge)"
									@mousedown.stop
								>
									<button
										class="vision-edge-toolbar__button"
										type="button"
										@mousedown.stop
										@click="startEdgeLabelEdit"
									>
										<Icon icon="pen" />
									</button>
									<div
										v-if="showEdgeColorPicker"
										class="vision-edge-toolbar__colors"
									>
										<button
											v-for="color in edgeColors"
											:key="`${selectedEdge.id}-${color}`"
											class="vision-edge-toolbar__color"
											:class="{'is-active': selectedEdge.color === color || (selectedEdge.color === '' && color === edgeStrokeColor)}"
											:style="{backgroundColor: color}"
											type="button"
											@mousedown.stop
											@click="() => updateSelectedEdgeColor(color)"
										/>
									</div>
									<button
										class="vision-edge-toolbar__button"
										type="button"
										@mousedown.stop
										@click="showEdgeColorPicker = !showEdgeColorPicker"
									>
										<Icon icon="fill-drip" />
									</button>
									<button
										class="vision-edge-toolbar__button"
										type="button"
										@mousedown.stop
										@click="focusSelectedEdge"
									>
										<Icon icon="bullseye" />
									</button>
									<button
										class="vision-edge-toolbar__button"
										type="button"
										@mousedown.stop
										@click="() => removeEdge(selectedEdge.id)"
									>
										<Icon icon="trash-alt" />
									</button>
									<input
										v-if="isEditingEdgeLabel"
										ref="edgeLabelInputRef"
										v-model="edgeLabelDraft"
										class="vision-edge-toolbar__input"
										type="text"
										:placeholder="$t('project.vision_board.edgeLabelPlaceholder')"
										@mousedown.stop
										@keydown.enter.prevent="saveSelectedEdgeLabel"
										@keydown.esc.prevent="cancelEdgeLabelEdit"
										@blur="saveSelectedEdgeLabel"
									>
								</div>

								<button
									v-for="edge in labeledEdges"
									:key="`edge-label-${edge.id}`"
									class="vision-edge-label"
									:class="{'vision-edge-label--selected': selectedEdgeId === edge.id}"
									:style="getEdgeLabelStyle(edge)"
									type="button"
									@click.stop="selectEdge(edge.id)"
								>
									{{ edge.label }}
								</button>

								<div
									v-for="presence in activeCollaborators"
									:key="`${presence.sender.id}-${presence.sessionId}`"
									class="vision-board-cursor"
									:style="getCollaboratorCursorStyle(presence)"
								>
									<Icon
										class="vision-board-cursor__pointer"
										icon="location-arrow"
										:style="{color: getCollaboratorColor(presence)}"
									/>
									<div
										class="vision-board-cursor__label"
										:style="{backgroundColor: getCollaboratorColor(presence)}"
									>
										{{ getCollaboratorDisplayName(presence.sender) }} {{ getCollaboratorActivityLabel(presence.activity) }}
									</div>
								</div>

								<div
									v-if="selectedNode !== null && selectedEdge === null"
									class="vision-node-toolbar"
									:style="getNodeToolbarStyle(selectedNode)"
									@mousedown.stop
								>
									<div
										v-if="showNodeColorPicker"
										class="vision-node-toolbar__colors"
									>
										<button
											v-for="color in nodeColors"
											:key="`${selectedNode.id}-${color}`"
											class="vision-node-toolbar__color"
											:class="{'is-active': selectedNode.color === color}"
											:style="{backgroundColor: color}"
											type="button"
											@mousedown.stop
											@click="() => updateSelectedNodeColor(color)"
										/>
									</div>
									<button
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="showNodeColorPicker = !showNodeColorPicker"
									>
										<Icon icon="fill-drip" />
									</button>
									<button
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="focusSelectedNode"
									>
										<Icon icon="bullseye" />
									</button>
									<button
										v-if="selectedNode.kind === 'video'"
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="openVideoUrlEditor"
									>
										<Icon icon="pen" />
									</button>
									<button
										v-if="selectedNode.kind === 'card'"
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="toggleCardTaskEditor"
									>
										<Icon icon="pen" />
									</button>
									<button
										v-if="selectedNode.kind === 'card'"
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="() => linkBoardTask(selectedNode)"
									>
										<Icon icon="link" />
									</button>
									<button
										v-if="selectedNode.kind === 'image' || selectedNode.kind === 'video'"
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="openNodeTitleEditor"
									>
										<Icon icon="heading" />
									</button>
									<button
										v-if="selectedNode.kind === 'image'"
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="showExistingImagePicker = !showExistingImagePicker"
									>
										<Icon icon="file-image" />
									</button>
									<label
										v-if="selectedNode.kind === 'image'"
										class="vision-node-toolbar__button vision-node-toolbar__upload"
										@mousedown.stop
									>
										<input
											type="file"
											accept="image/*"
											@change="(event) => uploadNodeImage(selectedNode, event)"
										>
										<Icon icon="arrow-up-from-bracket" />
									</label>
									<button
										class="vision-node-toolbar__button"
										type="button"
										@mousedown.stop
										@click="() => removeNode(selectedNode.id)"
									>
										<Icon icon="trash-alt" />
									</button>
									<input
										v-if="selectedNode.kind === 'video' && showVideoUrlEditor"
										ref="videoUrlInputRef"
										v-model="videoUrlDraft"
										class="vision-node-toolbar__input"
										type="url"
										placeholder="https://youtube.com/..."
										@mousedown.stop
										@keydown.enter.prevent="saveSelectedVideoUrl"
										@keydown.esc.prevent="cancelVideoUrlEdit"
										@blur="saveSelectedVideoUrl"
									>
									<input
										v-if="editingNodeTitleId === selectedNode?.id"
										v-model="nodeTitleDraft"
										class="vision-node-toolbar__input vision-node-toolbar__title-input"
										type="text"
										placeholder="Enter title..."
										@mousedown.stop
										@keydown.enter.prevent="saveNodeTitle"
										@keydown.esc.prevent="cancelNodeTitleEdit"
										@blur="saveNodeTitle"
									>
									<input
										v-if="editingCardTaskNodeId === selectedNode?.id"
										ref="cardTaskInputRef"
										v-model="cardQueries[selectedNode.id]"
										class="vision-node-toolbar__input"
										type="text"
										:placeholder="$t('project.vision_board.cardSearchPlaceholder')"
										@mousedown.stop
										@input="(event) => searchCardTasks(selectedNode.id, (event.target as HTMLInputElement).value)"
									>
									<div
										v-if="editingCardTaskNodeId === selectedNode?.id && (cardSearchResults[selectedNode.id] ?? []).length > 0"
										class="vision-node-toolbar__search-results"
									>
										<button
											v-for="task in cardSearchResults[selectedNode.id]"
											:key="task.id"
											class="vision-node__search-result vision-node__interactive"
											type="button"
											@mousedown.stop
											@click="() => selectCardTask(selectedNode, task)"
										>
											{{ task.title }}
										</button>
									</div>
									<BaseButton
										v-if="editingCardTaskNodeId === selectedNode?.id && (cardQueries[selectedNode.id] ?? '').trim() !== ''"
										class="vision-node-toolbar__action"
										@mousedown.stop
										@click="() => createAndLinkCardTask(selectedNode)"
									>
										{{ $t('project.vision_board.createLinkedTask') }}
									</BaseButton>
									<div
										v-if="selectedNode.kind === 'image' && showExistingImagePicker"
										class="vision-node-toolbar__attachments"
									>
										<button
											v-for="attachment in existingImageAttachments"
											:key="attachment.id"
											class="vision-node-toolbar__attachment"
											type="button"
											@mousedown.stop
											@click="() => selectExistingImageAttachment(selectedNode, attachment)"
										>
											<img
												:src="existingImageAttachmentUrls[attachment.id]"
												:alt="attachment.file.name"
												class="vision-node-toolbar__attachment-image"
											>
										</button>
									</div>
								</div>

								<div
									v-for="node in selectedBoardNodes"
									:key="node.id"
									:ref="(element) => setNodeRef(node.id, element)"
									class="vision-node"
									:class="[
										`vision-node--${node.kind}`,
										{'vision-node--editing': editingTextNodeId === node.id},
										{'vision-node--selected': isNodeSelected(node.id)},
									]"
									:style="getNodeStyle(node)"
									@mousedown.stop="(event) => startDrag(node, event)"
								>
									<button
										v-for="handle in connectionHandles"
										:key="`${node.id}-${handle}`"
										class="vision-node__handle vision-node__interactive"
										:class="`vision-node__handle--${handle}`"
										type="button"
										@mousedown.stop="(event) => startEdgeConnection(node.id, handle, event)"
										@mouseup.stop="() => finishEdgeConnection(node.id, handle)"
									/>

									<div
										v-if="node.kind === 'text'"
										class="vision-node__text"
									>
										<textarea
											v-if="editingTextNodeId === node.id"
											:ref="(element) => setTextEditorRef(node.id, element)"
											v-model="textDrafts[node.id]"
											class="vision-node__content vision-node__content--editor vision-node__interactive"
											@keydown.esc.prevent="() => finishTextEdit(node)"
										/>
										<div
											v-else
											class="vision-node__text-display"
											@dblclick.stop="() => startTextEdit(node)"
										>
											{{ node.content || node.title || ' ' }}
										</div>
									</div>

									<template v-else>
										<textarea
											v-if="node.kind === 'container'"
											v-model="node.content"
											class="vision-node__content vision-node__interactive"
											@change="() => updateNode(node)"
										/>

										<div
											v-else-if="node.kind === 'image'"
											class="vision-node__media"
										>
											<div
												v-if="node.title"
												class="vision-node__title-label"
												:style="{backgroundColor: node.color || undefined}"
											>
												{{ node.title }}
											</div>
											<img
												v-if="getNodeImageSrc(node)"
												:src="getNodeImageSrc(node)"
												:alt="node.title || $t('project.vision_board.nodes.image')"
												class="vision-node__image"
												@dragstart.prevent
											>
										</div>

										<div
											v-else-if="node.kind === 'video'"
											class="vision-node__media"
										>
											<div
												v-if="node.title"
												class="vision-node__title-label"
												:style="{backgroundColor: node.color || undefined}"
											>
												{{ node.title }}
											</div>
											<iframe
												v-if="getVideoEmbedUrl(node.url)"
												:src="getVideoEmbedUrl(node.url) ?? undefined"
												class="vision-node__video"
												allowfullscreen
												@dragstart.prevent
											/>
										</div>

										<div
											v-else-if="node.kind === 'card'"
											class="vision-node__card"
										>
											<KanbanCard
												v-if="getCardTask(node) !== null"
												class="vision-node__interactive vision-node__card-preview"
												:task="getCardTask(node)!"
												:project-id="props.projectId"
												open-behavior="route"
												@mousedown.stop
											/>
											<template v-else>
												<p class="vision-node__content-preview">
													{{ node.taskId ? `${$t('task.task')} #${node.taskId}` : $t('project.vision_board.cardUnlinked') }}
												</p>
												<BaseButton
													class="vision-node__interactive"
													@mousedown.stop
													@click="() => linkBoardTask(node)"
												>
													{{ $t('project.vision_board.linkBoardTask') }}
												</BaseButton>
												<input
													v-model="cardQueries[node.id]"
													class="vision-node__title vision-node__interactive"
													type="text"
													:placeholder="$t('project.vision_board.cardSearchPlaceholder')"
													@input="(event) => searchCardTasks(node.id, (event.target as HTMLInputElement).value)"
												>
												<div
													v-if="(cardSearchResults[node.id] ?? []).length > 0"
													class="vision-node__search-results"
												>
													<button
														v-for="task in cardSearchResults[node.id]"
														:key="task.id"
														class="vision-node__search-result vision-node__interactive"
														type="button"
														@click="() => selectCardTask(node, task)"
													>
														{{ task.title }}
													</button>
												</div>
												<BaseButton
													v-if="(cardQueries[node.id] ?? '').trim() !== ''"
													class="vision-node__interactive"
													@mousedown.stop
													@click="() => createAndLinkCardTask(node)"
												>
													{{ $t('project.vision_board.createLinkedTask') }}
												</BaseButton>
											</template>
										</div>
									</template>

									<div
										class="vision-node__resize-handle vision-node__interactive"
										@mousedown.stop="(event) => startResize(node, event)"
									/>
								</div>

								<p
									v-if="selectedBoardNodes.length === 0"
									class="vision-board-stage__empty"
								>
									{{ $t('project.vision_board.editorPlaceholder') }}
								</p>
							</div>
						</div>
					</div>
					<div class="vision-board-stage__viewport-controls">
						<button
							class="vision-board-stage__viewport-button"
							type="button"
							:disabled="currentZoom <= minZoom"
							@click="() => changeZoom(-zoomStep)"
						>
							-
						</button>
						<div class="vision-board-stage__zoom-readout">
							{{ zoomPercentage }}%
						</div>
						<button
							class="vision-board-stage__viewport-button"
							type="button"
							:disabled="currentZoom >= maxZoom"
							@click="() => changeZoom(zoomStep)"
						>
							+
						</button>
						<button
							class="vision-board-stage__viewport-button"
							type="button"
							@click="toggleFullscreen"
						>
							<Icon :icon="isFullscreen ? 'times' : 'arrow-up-right-from-square'" />
						</button>
					</div>
				</div>
			</section>
		</template>
	</ProjectWrapper>
</template>

<script setup lang="ts">
import {computed, nextTick, onMounted, onUnmounted, ref, shallowReactive, watch} from 'vue'
import {useRouteQuery} from '@vueuse/router'
import {useDebounceFn, useResizeObserver, useThrottleFn} from '@vueuse/core'

import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ITask} from '@/modelTypes/ITask'
import type {IVisionBoard} from '@/modelTypes/IVisionBoard'
import type {IVisionBoardEdge} from '@/modelTypes/IVisionBoardEdge'
import type {IVisionBoardNode, VisionBoardNodeKind} from '@/modelTypes/IVisionBoardNode'

import {canPreviewImage} from '@/models/attachment'
import VisionBoardModel from '@/models/visionBoard'
import AttachmentModel from '@/models/attachment'
import VisionBoardEdgeModel from '@/models/visionBoardEdge'
import VisionBoardNodeModel from '@/models/visionBoardNode'
import TaskModel from '@/models/task'
import AttachmentService, {PREVIEW_SIZE} from '@/services/attachment'
import VisionBoardService from '@/services/visionBoard'
import VisionBoardEdgeService from '@/services/visionBoardEdge'
import VisionBoardNodeService from '@/services/visionBoardNode'
import TaskCollectionService, {getDefaultTaskFilterParams, type TaskFilterParams} from '@/services/taskCollection'
import TaskService from '@/services/task'
import {uploadFile} from '@/helpers/attachments'
import {useWebSocket} from '@/composables/useWebSocket'

import BaseButton from '@/components/base/BaseButton.vue'
import FormField from '@/components/input/FormField.vue'
import Message from '@/components/misc/Message.vue'
import Multiselect from '@/components/input/Multiselect.vue'
import Icon from '@/components/misc/Icon'
import ProjectWrapper from '@/components/project/ProjectWrapper.vue'
import KanbanCard from '@/components/tasks/partials/KanbanCard.vue'

type ConnectionHandle = Exclude<IVisionBoardEdge['sourceHandle'], ''>
type CollaboratorActivity = 'idle' | 'dragging' | 'resizing' | 'connecting'

interface RealtimeSender {
	id: number
	name: string
	username: string
}

interface RealtimeEnvelope<T> {
	sender?: RealtimeSender
	data?: T
}

interface BoardChangedPayload {
	sessionId: string
	reason?: string
}

interface BoardPresencePayload {
	sessionId: string
	x?: number
	y?: number
	left?: boolean
	activity: CollaboratorActivity
}

interface CollaboratorPresence {
	sessionId: string
	x: number
	y: number
	activity: CollaboratorActivity
	sender: RealtimeSender
	updatedAt: number
}

const props = defineProps<{
	projectId: number,
	viewId: number,
	isLoadingProject: boolean,
}>()

const visionBoardService = shallowReactive(new VisionBoardService())
const attachmentService = shallowReactive(new AttachmentService())
const visionBoardNodeService = shallowReactive(new VisionBoardNodeService())
const visionBoardEdgeService = shallowReactive(new VisionBoardEdgeService())
const taskCollectionService = shallowReactive(new TaskCollectionService())
const taskService = shallowReactive(new TaskService())
const {connected: wsConnected, publish, subscribe} = useWebSocket()

const boards = ref<IVisionBoard[]>([])
const activeBoard = ref<IVisionBoard | null>(null)
const foundTasks = ref<ITask[]>([])
const selectedTask = ref<ITask | null>(null)
const showCreateForm = ref(false)
const nodeKinds: VisionBoardNodeKind[] = ['text', 'card', 'container', 'image', 'video']
const nodeColors = ['#38425c', '#4e4f8d', '#2f5d73', '#6f4554', '#5f4b8b', '#6b5a3a', '#2f6758']
const connectionHandles: ConnectionHandle[] = ['top', 'left', 'right', 'bottom']
const edgeStrokeColor = '#8c5fd3'
const edgeColors = ['#8c5fd3', '#62e7c7', '#f59ac2', '#74b8ff', '#ffd166', '#ff8c69', '#c4a7ff']
const cardQueries = ref<Record<number, string>>({})
const cardSearchResults = ref<Record<number, ITask[]>>({})
const cardTasks = ref<Record<number, ITask>>({})
const textDrafts = ref<Record<number, string>>({})
const editingTextNodeId = ref<number | null>(null)
const editingCardTaskNodeId = ref<number | null>(null)
const selectedNodeId = ref<number | null>(null)
const selectedEdgeId = ref<number | null>(null)
const isEditingEdgeLabel = ref(false)
const showEdgeColorPicker = ref(false)
const edgeLabelDraft = ref('')
const isFullscreen = ref(false)
const showNodeColorPicker = ref(false)
const showExistingImagePicker = ref(false)
const showVideoUrlEditor = ref(false)
const videoUrlDraft = ref('')
const editingNodeTitleId = ref<number | null>(null)
const nodeTitleDraft = ref('')
const viewportRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLElement | null>(null)
const edgeLabelInputRef = ref<HTMLInputElement | null>(null)
const videoUrlInputRef = ref<HTMLInputElement | null>(null)
const cardTaskInputRef = ref<HTMLInputElement | null>(null)
const canvasSize = ref({width: 2400, height: 1600})
const nodeImageBlobUrls = ref<Record<number, string>>({})
const existingImageAttachments = ref<IAttachment[]>([])
const existingImageAttachmentUrls = ref<Record<number, string>>({})
const viewportHeight = ref(640)
const collaboratorPresence = ref<Record<string, CollaboratorPresence>>({})
const pendingRealtimeReload = ref(false)
const lastPresencePoint = ref<{x: number, y: number} | null>(null)
const minZoom = .5
const maxZoom = 2
const zoomStep = .1
const collaboratorColors = ['#5cc8ff', '#ff8c69', '#62e7c7', '#ffd166', '#f59ac2', '#b39cff']
const collaboratorPresenceTimeout = 10000
const localRealtimeSessionId = typeof crypto !== 'undefined' && 'randomUUID' in crypto
	? crypto.randomUUID()
	: `${Date.now()}-${Math.random().toString(36).slice(2)}`
const draggingNodeId = ref<number | null>(null)
const dragOffset = ref({x: 0, y: 0})
const resizingNodeId = ref<number | null>(null)
const resizeOrigin = ref({x: 0, y: 0, width: 0, height: 0})

const selectedBoardIdQuery = useRouteQuery('visionBoardId')

const nodeRefs = new Map<number, HTMLElement>()
const textEditorRefs = new Map<number, HTMLTextAreaElement>()
let unsubscribeBoardChangedWs: (() => void) | null = null
let unsubscribeBoardPresenceWs: (() => void) | null = null
let collaboratorPresenceCleanupTimer: ReturnType<typeof window.setInterval> | null = null

const pendingEdge = ref<null | {
	sourceNodeId: number,
	sourceHandle: ConnectionHandle,
	currentPoint: {x: number, y: number},
}>(null)

const selectedBoardId = computed<number | null>(() => {
	const value = selectedBoardIdQuery.value
	if (value === undefined || value === null || value === '') {
		return null
	}

	const parsed = Number(value)
	return Number.isNaN(parsed) ? null : parsed
})

const selectedBoardIdValue = computed(() => selectedBoardId.value?.toString() ?? '')
const selectedBoardNodes = computed(() => activeBoard.value?.nodes ?? [])
const selectedBoardEdges = computed(() => activeBoard.value?.edges ?? [])
const currentZoom = computed(() => activeBoard.value?.viewportZoom || 1)
const zoomPercentage = computed(() => Math.round(currentZoom.value * 100))
const labeledEdges = computed(() => selectedBoardEdges.value.filter(edge => edge.label.trim() !== ''))
const boardRealtimeEventBase = computed(() => activeBoard.value === null ? '' : `project.${props.projectId}.board.${activeBoard.value.id}`)
const boardChangedEvent = computed(() => boardRealtimeEventBase.value === '' ? '' : `${boardRealtimeEventBase.value}.changed`)
const boardPresenceEvent = computed(() => boardRealtimeEventBase.value === '' ? '' : `${boardRealtimeEventBase.value}.presence`)
const activeCollaborators = computed(() => Object.values(collaboratorPresence.value))
const selectedNode = computed(() => {
	if (selectedNodeId.value === null) {
		return null
	}

	return selectedBoardNodes.value.find(node => node.id === selectedNodeId.value) ?? null
})
const selectedEdge = computed(() => {
	if (selectedEdgeId.value === null) {
		return null
	}

	return selectedBoardEdges.value.find(edge => edge.id === selectedEdgeId.value) ?? null
})

function setNodeRef(nodeId: number, element: Element | null) {
	if (!(element instanceof HTMLElement)) {
		nodeRefs.delete(nodeId)
		return
	}

	nodeRefs.set(nodeId, element)
}

function setTextEditorRef(nodeId: number, element: Element | null) {
	if (!(element instanceof HTMLTextAreaElement)) {
		textEditorRefs.delete(nodeId)
		return
	}

	textEditorRefs.set(nodeId, element)
}

function refreshCanvasSize() {
	const width = canvasRef.value?.clientWidth ?? 1600
	const height = canvasRef.value?.clientHeight ?? 1000
	canvasSize.value = {
		width: Math.max(width, getBoardContentBounds().width),
		height: Math.max(height, getBoardContentBounds().height),
	}
}

function refreshViewportHeight() {
	const top = viewportRef.value?.getBoundingClientRect().top
	if (top === undefined) {
		return
	}

	viewportHeight.value = Math.max(420, Math.floor(window.innerHeight - top - 24))
}

function getBoardContentBounds() {
	let width = 1600
	let height = 1000

	selectedBoardNodes.value.forEach(node => {
		width = Math.max(width, node.x + node.width + 240)
		height = Math.max(height, node.y + node.height + 240)
	})

	return {width, height}
}

async function loadBoard(boardId: number) {
	activeBoard.value = await visionBoardService.get({
		projectId: props.projectId,
		id: boardId,
	})
	collaboratorPresence.value = {}
	editingTextNodeId.value = null
	selectedNodeId.value = null
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null
	editingCardTaskNodeId.value = null
	edgeLabelDraft.value = ''
	nodeTitleDraft.value = ''
	pendingEdge.value = null
	await hydrateCardTasks()
	await hydrateImageNodeUrls()
	await nextTick()
	refreshViewportHeight()
	refreshCanvasSize()
}

async function loadBoards() {
	boards.value = await visionBoardService.getAll({projectId: props.projectId})

	if (selectedBoardId.value !== null && !boards.value.some(board => board.id === selectedBoardId.value)) {
		selectedBoardIdQuery.value = undefined
	}

	if (selectedBoardId.value !== null) {
		await loadBoard(selectedBoardId.value)
	} else {
		activeBoard.value = null
	}

	if (boards.value.length === 0) {
		showCreateForm.value = true
	}
}

watch(
	() => props.projectId,
	loadBoards,
	{immediate: true},
)

watch(selectedBoardId, async (boardId) => {
	if (boardId === null) {
		activeBoard.value = null
		return
	}

	await loadBoard(boardId)
})

watch([selectedBoardNodes, selectedBoardEdges], refreshCanvasSize, {deep: true})
watch(selectedNode, async (node) => {
	showExistingImagePicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null
	if (node?.kind !== 'card') {
		editingCardTaskNodeId.value = null
	}
	if (node?.kind !== 'image') {
		existingImageAttachments.value = []
		existingImageAttachmentUrls.value = {}
	}

	if (node?.kind !== 'video') {
		videoUrlDraft.value = ''
	}

	if (node?.kind === 'image') {
		await loadExistingImageAttachments()
	}
})

watch(boardChangedEvent, (eventName) => {
	unsubscribeBoardChangedWs?.()
	unsubscribeBoardChangedWs = null

	if (eventName === '') {
		return
	}

	unsubscribeBoardChangedWs = subscribe(eventName, (msg) => {
		if (msg.event !== eventName) {
			return
		}

		const payload = msg.data as RealtimeEnvelope<BoardChangedPayload> | undefined
		if (payload?.data?.sessionId === localRealtimeSessionId) {
			return
		}

		reloadBoardFromRealtime()
	})
}, {immediate: true})

watch(boardPresenceEvent, (eventName, previousEventName) => {
	if (previousEventName) {
		publishPresenceLeave(previousEventName)
	}

	unsubscribeBoardPresenceWs?.()
	unsubscribeBoardPresenceWs = null
	collaboratorPresence.value = {}

	if (eventName === '') {
		return
	}

	unsubscribeBoardPresenceWs = subscribe(eventName, (msg) => {
		if (msg.event !== eventName) {
			return
		}

		const payload = msg.data as RealtimeEnvelope<BoardPresencePayload> | undefined
		if (!payload?.sender || !payload.data?.sessionId || payload.data.sessionId === localRealtimeSessionId) {
			return
		}

		const key = getCollaboratorPresenceKey(payload.sender, payload.data.sessionId)
		if (payload.data.left) {
			const nextPresence = {...collaboratorPresence.value}
			delete nextPresence[key]
			collaboratorPresence.value = nextPresence
			return
		}

		if (typeof payload.data.x !== 'number' || typeof payload.data.y !== 'number') {
			return
		}

		collaboratorPresence.value = {
			...collaboratorPresence.value,
			[key]: {
				sessionId: payload.data.sessionId,
				x: payload.data.x,
				y: payload.data.y,
				activity: payload.data.activity,
				sender: payload.sender,
				updatedAt: Date.now(),
			},
		}
	})
}, {immediate: true})

watch(wsConnected, (isConnected, wasConnected) => {
	if (wasConnected && !isConnected) {
		reloadBoardFromRealtime()
	}
})

onMounted(() => {
	refreshViewportHeight()
	refreshCanvasSize()
	collaboratorPresenceCleanupTimer = window.setInterval(pruneCollaboratorPresence, 2000)
	window.addEventListener('mousedown', handleGlobalMouseDown)
	window.addEventListener('mousemove', handlePointerMove)
	window.addEventListener('mouseup', stopPointerTracking)
	window.addEventListener('keydown', handleWindowKeydown)
	window.addEventListener('resize', refreshViewportHeight)
})

useResizeObserver(canvasRef, refreshCanvasSize)
useResizeObserver(viewportRef, refreshViewportHeight)

onUnmounted(() => {
	publishPresenceLeave()
	unsubscribeBoardChangedWs?.()
	unsubscribeBoardPresenceWs?.()
	if (collaboratorPresenceCleanupTimer !== null) {
		window.clearInterval(collaboratorPresenceCleanupTimer)
	}
	window.removeEventListener('mousedown', handleGlobalMouseDown)
	window.removeEventListener('mousemove', handlePointerMove)
	window.removeEventListener('mouseup', stopPointerTracking)
	window.removeEventListener('keydown', handleWindowKeydown)
	window.removeEventListener('resize', refreshViewportHeight)
})

const persistViewport = useDebounceFn(async () => {
	if (activeBoard.value === null) {
		return
	}

	await visionBoardService.update(new VisionBoardModel({
		id: activeBoard.value.id,
		projectId: props.projectId,
		taskId: activeBoard.value.taskId,
		title: activeBoard.value.title,
		viewportX: activeBoard.value.viewportX,
		viewportY: activeBoard.value.viewportY,
		viewportZoom: activeBoard.value.viewportZoom,
	}))
}, 150)

const debouncedBoardReload = useDebounceFn(async (boardId: number) => {
	await loadBoard(boardId)
}, 250)

function reloadBoardFromRealtime() {
	if (activeBoard.value === null) {
		return
	}

	if (draggingNodeId.value !== null || resizingNodeId.value !== null || editingTextNodeId.value !== null || pendingEdge.value !== null) {
		pendingRealtimeReload.value = true
		return
	}

	pendingRealtimeReload.value = false
	void debouncedBoardReload(activeBoard.value.id)
}

function notifyBoardChanged(reason: string) {
	if (boardChangedEvent.value === '') {
		return
	}

	publish(boardChangedEvent.value, {
		sessionId: localRealtimeSessionId,
		reason,
	} satisfies BoardChangedPayload)
}

function getCollaboratorDisplayName(sender: RealtimeSender) {
	return sender.name || sender.username
}

function getCollaboratorActivityLabel(activity: CollaboratorActivity) {
	switch (activity) {
		case 'dragging':
			return 'moving'
		case 'resizing':
			return 'resizing'
		case 'connecting':
			return 'connecting'
		default:
			return ''
	}
}

function getCollaboratorPresenceKey(sender: RealtimeSender, sessionId: string) {
	return `${sender.id}:${sessionId}`
}

function getCollaboratorColor(presence: CollaboratorPresence) {
	const seed = `${presence.sender.id}:${presence.sessionId}`
	let total = 0
	for (const char of seed) {
		total += char.charCodeAt(0)
	}

	return collaboratorColors[total % collaboratorColors.length]
}

function getCollaboratorCursorStyle(presence: CollaboratorPresence) {
	return {
		left: `${presence.x}px`,
		top: `${presence.y}px`,
	}
}

function getCurrentCollaboratorActivity(): CollaboratorActivity {
	if (draggingNodeId.value !== null) {
		return 'dragging'
	}
	if (resizingNodeId.value !== null) {
		return 'resizing'
	}
	if (pendingEdge.value !== null) {
		return 'connecting'
	}

	return 'idle'
}

function publishPresenceLeave(eventName = boardPresenceEvent.value) {
	if (eventName === '') {
		return
	}

	publish(eventName, {
		sessionId: localRealtimeSessionId,
		left: true,
		activity: 'idle',
	} satisfies BoardPresencePayload)
}

const publishBoardPresence = useThrottleFn((payload: BoardPresencePayload) => {
	if (boardPresenceEvent.value === '') {
		return
	}

	publish(boardPresenceEvent.value, payload)
}, 60)

function publishIdlePresence() {
	if (lastPresencePoint.value === null) {
		return
	}

	publishBoardPresence({
		sessionId: localRealtimeSessionId,
		x: lastPresencePoint.value.x,
		y: lastPresencePoint.value.y,
		activity: 'idle',
	})
}

function handleCanvasPresenceMove(event: MouseEvent) {
	if (activeBoard.value === null) {
		return
	}

	const point = toCanvasPoint(event)
	lastPresencePoint.value = point
	publishBoardPresence({
		sessionId: localRealtimeSessionId,
		x: point.x,
		y: point.y,
		activity: getCurrentCollaboratorActivity(),
	})
}

function handleCanvasPresenceLeave() {
	lastPresencePoint.value = null
	publishPresenceLeave()
}

function pruneCollaboratorPresence() {
	const threshold = Date.now() - collaboratorPresenceTimeout
	collaboratorPresence.value = Object.fromEntries(
		Object.entries(collaboratorPresence.value)
			.filter(([, presence]) => presence.updatedAt >= threshold),
	)
}

function getTaskSearchParams(query: string): TaskFilterParams {
	return {
		...getDefaultTaskFilterParams(),
		sort_by: ['id'],
		order_by: ['desc'],
		s: query,
	}
}

async function findTasks(query: string) {
	foundTasks.value = await taskCollectionService.getAll(
		{projectId: props.projectId},
		getTaskSearchParams(query),
	)
}

function onBoardSelect(event: Event) {
	const target = event.target as HTMLSelectElement
	selectedBoardIdQuery.value = target.value === '' ? undefined : target.value
}

async function createBoard() {
	if (selectedTask.value === null) {
		return
	}

	const existing = boards.value.find(board => board.taskId === selectedTask.value?.id)
	if (existing) {
		selectedBoardIdQuery.value = String(existing.id)
		showCreateForm.value = false
		return
	}

	const board = await visionBoardService.create(new VisionBoardModel({
		projectId: props.projectId,
		taskId: selectedTask.value.id,
		title: selectedTask.value.title,
	}))

	await loadBoards()
	selectedBoardIdQuery.value = String(board.id)
	showCreateForm.value = false
	selectedTask.value = null
	foundTasks.value = []
}

function getDefaultNodePosition() {
	const offset = selectedBoardNodes.value.length * 24
	return {x: 48 + offset, y: 48 + offset}
}

async function addNode(kind: VisionBoardNodeKind, position = getDefaultNodePosition()) {
	if (activeBoard.value === null) {
		return
	}

	const isTextNode = kind === 'text'
	const node = await visionBoardNodeService.create(new VisionBoardNodeModel({
		projectId: props.projectId,
		boardId: activeBoard.value.id,
		kind,
		title: isTextNode ? '' : `${kind[0].toUpperCase()}${kind.slice(1)} node`,
		content: '',
		x: position.x,
		y: position.y,
		width: isTextNode ? 300 : 280,
		height: isTextNode ? 160 : 220,
		color: '',
	}))

	activeBoard.value.nodes = [...selectedBoardNodes.value, node]
	notifyBoardChanged('node.created')
	refreshCanvasSize()

	if (isTextNode) {
		await startTextEdit(node)
	} else {
		selectedNodeId.value = node.id
	}
}

async function hydrateImageNodeUrls() {
	if (activeBoard.value === null) {
		return
	}

	const imageNodes = selectedBoardNodes.value.filter(node => node.kind === 'image' && node.attachmentId > 0)
	const entries = await Promise.all(imageNodes.map(async node => {
		try {
			const blobUrl = await attachmentService.getBlobUrl(new AttachmentModel({
				id: node.attachmentId,
				taskId: activeBoard.value?.taskId ?? 0,
			}), PREVIEW_SIZE.LG) as string
			return [node.id, blobUrl] as const
		} catch {
			return [node.id, ''] as const
		}
	}))

	nodeImageBlobUrls.value = entries.reduce<Record<number, string>>((acc, [nodeId, blobUrl]) => {
		if (blobUrl !== '') {
			acc[nodeId] = blobUrl
		}
		return acc
	}, {})
}

async function hydrateCardTasks() {
	if (activeBoard.value === null) {
		cardTasks.value = {}
		return
	}

	const taskIds = [...new Set(
		selectedBoardNodes.value
			.filter(node => node.kind === 'card' && node.taskId > 0)
			.map(node => node.taskId),
	)]

	const entries = await Promise.all(taskIds.map(async taskId => {
		try {
			const task = await taskService.get({id: taskId})
			return [taskId, task] as const
		} catch {
			return [taskId, null] as const
		}
	}))

	cardTasks.value = entries.reduce<Record<number, ITask>>((acc, [taskId, task]) => {
		if (task !== null) {
			acc[taskId] = task
		}
		return acc
	}, {})
}

async function loadExistingImageAttachments() {
	if (activeBoard.value === null) {
		return
	}

	const attachments = await attachmentService.getAll({taskId: activeBoard.value.taskId})
	existingImageAttachments.value = attachments.filter(canPreviewImage)
	const previews = await Promise.all(existingImageAttachments.value.map(async attachment => {
		try {
			const blobUrl = await attachmentService.getBlobUrl(attachment, PREVIEW_SIZE.MD) as string
			return [attachment.id, blobUrl] as const
		} catch {
			return [attachment.id, ''] as const
		}
	}))

	existingImageAttachmentUrls.value = previews.reduce<Record<number, string>>((acc, [attachmentId, blobUrl]) => {
		if (blobUrl !== '') {
			acc[attachmentId] = blobUrl
		}
		return acc
	}, {})
}

async function removeNode(nodeId: number) {
	if (activeBoard.value === null) {
		return
	}

	const edgesToDelete = selectedBoardEdges.value.filter(edge => edge.sourceNodeId === nodeId || edge.targetNodeId === nodeId)
	for (const edge of edgesToDelete) {
		await removeEdge(edge.id)
	}

	await visionBoardNodeService.delete(new VisionBoardNodeModel({
		projectId: props.projectId,
		boardId: activeBoard.value.id,
		id: nodeId,
	}))

	activeBoard.value.nodes = selectedBoardNodes.value.filter(node => node.id !== nodeId)
	if (editingTextNodeId.value === nodeId) {
		editingTextNodeId.value = null
	}
	if (selectedNodeId.value === nodeId) {
		selectedNodeId.value = null
	}
	notifyBoardChanged('node.deleted')
	refreshCanvasSize()
}

async function updateNode(node: IVisionBoardNode) {
	await visionBoardNodeService.update(new VisionBoardNodeModel({
		...node,
		projectId: props.projectId,
		boardId: activeBoard.value?.id ?? node.boardId,
	}))
	notifyBoardChanged('node.updated')
}

function getNodeStyle(node: IVisionBoardNode) {
	return {
		left: `${node.x}px`,
		top: `${node.y}px`,
		width: `${node.width}px`,
		height: `${node.height}px`,
		backgroundColor: node.color || undefined,
	}
}

function getNodeImageSrc(node: IVisionBoardNode) {
	return nodeImageBlobUrls.value[node.id] || ''
}

async function selectExistingImageAttachment(node: IVisionBoardNode, attachment: IAttachment) {
	node.attachmentId = attachment.id
	nodeImageBlobUrls.value[node.id] = existingImageAttachmentUrls.value[attachment.id] || await attachmentService.getBlobUrl(attachment, PREVIEW_SIZE.LG) as string
	showExistingImagePicker.value = false
	await updateNode(node)
}

function getNodeToolbarStyle(node: IVisionBoardNode) {
	return {
		left: `${node.x + node.width / 2}px`,
		top: `${node.y}px`,
		transform: 'translate(-50%, calc(-100% - 1rem))',
	}
}

function getNodeById(nodeId: number) {
	return selectedBoardNodes.value.find(node => node.id === nodeId)
}

function getCardTask(node: IVisionBoardNode) {
	return cardTasks.value[node.taskId] ?? null
}

function isNodeSelected(nodeId: number) {
	return selectedNodeId.value === nodeId || pendingEdge.value?.sourceNodeId === nodeId
}

function getNodeAnchor(nodeId: number, handle: IVisionBoardEdge['sourceHandle']) {
	const node = getNodeById(nodeId)
	if (!node) {
		return {x: 0, y: 0}
	}

	const inset = 6

	switch (handle) {
		case 'top':
			return {x: node.x + node.width / 2, y: node.y + inset}
		case 'left':
			return {x: node.x + inset, y: node.y + node.height / 2}
		case 'right':
			return {x: node.x + node.width - inset, y: node.y + node.height / 2}
		case 'bottom':
			return {x: node.x + node.width / 2, y: node.y + node.height - inset}
		default:
			return {x: node.x + node.width / 2, y: node.y + node.height / 2}
	}
}

function inferHandleBetweenNodes(nodeId: number, otherNodeId: number, fallback: ConnectionHandle = 'right'): ConnectionHandle {
	const node = getNodeById(nodeId)
	const otherNode = getNodeById(otherNodeId)
	if (!node || !otherNode) {
		return fallback
	}

	const dx = (otherNode.x + otherNode.width / 2) - (node.x + node.width / 2)
	const dy = (otherNode.y + otherNode.height / 2) - (node.y + node.height / 2)

	if (Math.abs(dx) >= Math.abs(dy)) {
		return dx >= 0 ? 'right' : 'left'
	}

	return dy >= 0 ? 'bottom' : 'top'
}

function getResolvedSourceHandle(edge: IVisionBoardEdge): ConnectionHandle {
	return edge.sourceHandle || inferHandleBetweenNodes(edge.sourceNodeId, edge.targetNodeId, 'right')
}

function getResolvedTargetHandle(edge: IVisionBoardEdge): ConnectionHandle {
	return edge.targetHandle || inferHandleBetweenNodes(edge.targetNodeId, edge.sourceNodeId, 'left')
}

function getBezierPath(
	source: {x: number, y: number},
	target: {x: number, y: number},
	sourceHandle: IVisionBoardEdge['sourceHandle'],
	targetHandle: IVisionBoardEdge['targetHandle'],
) {
	const dx = target.x - source.x
	const dy = target.y - source.y
	const horizontal = Math.max(Math.abs(dx) * .45, 80)
	const vertical = Math.max(Math.abs(dy) * .45, 80)

	const sourceControl = {
		x: source.x + (
			sourceHandle === 'right'
				? horizontal
				: sourceHandle === 'left'
					? -horizontal
					: 0
		),
		y: source.y + (
			sourceHandle === 'bottom'
				? vertical
				: sourceHandle === 'top'
					? -vertical
					: 0
		),
	}

	const targetControl = {
		x: target.x + (
			targetHandle === 'right'
				? horizontal
				: targetHandle === 'left'
					? -horizontal
					: 0
		),
		y: target.y + (
			targetHandle === 'bottom'
				? vertical
				: targetHandle === 'top'
					? -vertical
					: 0
		),
	}

	return `M ${source.x} ${source.y} C ${sourceControl.x} ${sourceControl.y}, ${targetControl.x} ${targetControl.y}, ${target.x} ${target.y}`
}

function getEdgePath(edge: IVisionBoardEdge) {
	const sourceHandle = getResolvedSourceHandle(edge)
	const targetHandle = getResolvedTargetHandle(edge)
	const source = getNodeAnchor(edge.sourceNodeId, sourceHandle)
	const target = getNodeAnchor(edge.targetNodeId, targetHandle)
	return getBezierPath(source, target, sourceHandle, targetHandle)
}

function getPendingEdgePath() {
	if (pendingEdge.value === null) {
		return ''
	}

	const source = getNodeAnchor(pendingEdge.value.sourceNodeId, pendingEdge.value.sourceHandle)
	return getBezierPath(source, pendingEdge.value.currentPoint, pendingEdge.value.sourceHandle, 'left')
}

function getEdgeLabelStyle(edge: IVisionBoardEdge) {
	const source = getNodeAnchor(edge.sourceNodeId, getResolvedSourceHandle(edge))
	const target = getNodeAnchor(edge.targetNodeId, getResolvedTargetHandle(edge))

	return {
		left: `${(source.x + target.x) / 2}px`,
		top: `${(source.y + target.y) / 2}px`,
	}
}

function toCanvasPoint(event: MouseEvent) {
	const rect = canvasRef.value?.getBoundingClientRect()
	if (!rect || !canvasRef.value) {
		return {x: 0, y: 0}
	}

	return {
		x: (event.clientX - rect.left + canvasRef.value.scrollLeft) / currentZoom.value,
		y: (event.clientY - rect.top + canvasRef.value.scrollTop) / currentZoom.value,
	}
}

function changeZoom(delta: number) {
	if (activeBoard.value === null || canvasRef.value === null) {
		return
	}

	const viewport = canvasRef.value
	const centerX = (viewport.scrollLeft + viewport.clientWidth / 2) / currentZoom.value
	const centerY = (viewport.scrollTop + viewport.clientHeight / 2) / currentZoom.value
	const nextZoom = Math.min(maxZoom, Math.max(minZoom, Number((currentZoom.value + delta).toFixed(2))))
	activeBoard.value.viewportZoom = nextZoom

	requestAnimationFrame(() => {
		if (!canvasRef.value) {
			return
		}

		canvasRef.value.scrollLeft = Math.max(0, centerX * nextZoom - canvasRef.value.clientWidth / 2)
		canvasRef.value.scrollTop = Math.max(0, centerY * nextZoom - canvasRef.value.clientHeight / 2)
	})

	persistViewport()
}

function toggleFullscreen() {
	isFullscreen.value = !isFullscreen.value
	requestAnimationFrame(() => {
		refreshViewportHeight()
		refreshCanvasSize()
	})
}

function handleWindowKeydown(event: KeyboardEvent) {
	const target = event.target as HTMLElement | null
	const isTypingTarget = target instanceof HTMLElement && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)

	if (event.key === 'Escape' && isFullscreen.value && editingTextNodeId.value === null) {
		isFullscreen.value = false
		requestAnimationFrame(() => {
			refreshViewportHeight()
			refreshCanvasSize()
		})
	}

	if (event.key === 'Escape' && isEditingEdgeLabel.value) {
		cancelEdgeLabelEdit()
	}

	if ((event.key === 'Delete' || event.key === 'Backspace') && selectedEdgeId.value !== null && !isTypingTarget) {
		event.preventDefault()
		void removeEdge(selectedEdgeId.value)
	}
}

function startEdgeConnection(nodeId: number, handle: ConnectionHandle, event: MouseEvent) {
	selectedNodeId.value = nodeId
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null
	const point = toCanvasPoint(event)
	pendingEdge.value = {
		sourceNodeId: nodeId,
		sourceHandle: handle,
		currentPoint: point,
	}
}

async function finishEdgeConnection(targetNodeId: number, targetHandle: ConnectionHandle) {
	if (pendingEdge.value === null || activeBoard.value === null) {
		return
	}

	if (pendingEdge.value.sourceNodeId === targetNodeId) {
		pendingEdge.value = null
		return
	}

	const edge = await visionBoardEdgeService.create(new VisionBoardEdgeModel({
		projectId: props.projectId,
		boardId: activeBoard.value.id,
		sourceNodeId: pendingEdge.value.sourceNodeId,
		targetNodeId,
		sourceHandle: pendingEdge.value.sourceHandle,
		targetHandle,
		label: '',
		color: '',
	}))

	activeBoard.value.edges = [...selectedBoardEdges.value, edge]
	selectedEdgeId.value = edge.id
	selectedNodeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	pendingEdge.value = null
	notifyBoardChanged('edge.created')
}

async function updateEdge(edge: IVisionBoardEdge) {
	await visionBoardEdgeService.update(new VisionBoardEdgeModel({
		...edge,
		projectId: props.projectId,
		boardId: activeBoard.value?.id ?? edge.boardId,
	}))
	notifyBoardChanged('edge.updated')
}

async function removeEdge(edgeId: number) {
	if (activeBoard.value === null) {
		return
	}

	await visionBoardEdgeService.delete(new VisionBoardEdgeModel({
		projectId: props.projectId,
		boardId: activeBoard.value.id,
		id: edgeId,
	}))

	activeBoard.value.edges = selectedBoardEdges.value.filter(edge => edge.id !== edgeId)
	if (selectedEdgeId.value === edgeId) {
		selectedEdgeId.value = null
	}
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	edgeLabelDraft.value = ''
	notifyBoardChanged('edge.deleted')
}

function getEdgeToolbarStyle(edge: IVisionBoardEdge) {
	const base = getEdgeLabelStyle(edge)
	return {
		...base,
		transform: 'translate(-50%, calc(-100% - 1rem))',
	}
}

async function updateSelectedNodeColor(color: string) {
	if (selectedNode.value === null) {
		return
	}

	selectedNode.value.color = color
	await updateNode(selectedNode.value)
}

async function openVideoUrlEditor() {
	if (selectedNode.value?.kind !== 'video') {
		return
	}

	showExistingImagePicker.value = false
	showVideoUrlEditor.value = true
	videoUrlDraft.value = selectedNode.value.url || ''
	await nextTick()
	videoUrlInputRef.value?.focus()
	videoUrlInputRef.value?.select()
}

function cancelVideoUrlEdit() {
	showVideoUrlEditor.value = false
	videoUrlDraft.value = selectedNode.value?.kind === 'video' ? selectedNode.value.url : ''
}

async function saveSelectedVideoUrl() {
	if (selectedNode.value?.kind !== 'video') {
		return
	}

	selectedNode.value.url = videoUrlDraft.value.trim()
	showVideoUrlEditor.value = false
	await updateNode(selectedNode.value)
}

async function openNodeTitleEditor() {
	if (selectedNode.value?.kind !== 'image' && selectedNode.value?.kind !== 'video') {
		return
	}

	showExistingImagePicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = selectedNode.value.id
	nodeTitleDraft.value = selectedNode.value.title || ''
	await nextTick()
	const inputRefs = document.querySelectorAll('.vision-node-toolbar__title-input')
	const input = inputRefs[inputRefs.length - 1] as HTMLInputElement | undefined
	input?.focus()
	input?.select()
}

function cancelNodeTitleEdit() {
	editingNodeTitleId.value = null
	nodeTitleDraft.value = selectedNode.value?.title ?? ''
}

async function saveNodeTitle() {
	if (selectedNode.value?.kind !== 'image' && selectedNode.value?.kind !== 'video') {
		return
	}

	selectedNode.value.title = nodeTitleDraft.value.trim()
	editingNodeTitleId.value = null
	await updateNode(selectedNode.value)
}

function focusSelectedNode() {
	if (selectedNode.value === null || activeBoard.value === null || canvasRef.value === null) {
		return
	}

	const midpoint = {
		x: selectedNode.value.x + selectedNode.value.width / 2,
		y: selectedNode.value.y + selectedNode.value.height / 2,
	}

	const nextZoom = Math.max(currentZoom.value, 1.15)
	activeBoard.value.viewportZoom = nextZoom
	persistViewport()

	requestAnimationFrame(() => {
		if (!canvasRef.value) {
			return
		}

		canvasRef.value.scrollTo({
			left: midpoint.x * nextZoom - canvasRef.value.clientWidth / 2,
			top: midpoint.y * nextZoom - canvasRef.value.clientHeight / 2,
			behavior: 'smooth',
		})
	})
}

async function startEdgeLabelEdit() {
	if (selectedEdge.value === null) {
		return
	}

	edgeLabelDraft.value = selectedEdge.value.label
	isEditingEdgeLabel.value = true
	showEdgeColorPicker.value = false
	await nextTick()
	edgeLabelInputRef.value?.focus()
	edgeLabelInputRef.value?.select()
}

function cancelEdgeLabelEdit() {
	isEditingEdgeLabel.value = false
	edgeLabelDraft.value = selectedEdge.value?.label ?? ''
}

async function saveSelectedEdgeLabel() {
	if (selectedEdge.value === null) {
		return
	}

	selectedEdge.value.label = edgeLabelDraft.value.trim()
	isEditingEdgeLabel.value = false
	await updateEdge(selectedEdge.value)
}

async function updateSelectedEdgeColor(color: string) {
	if (selectedEdge.value === null) {
		return
	}

	selectedEdge.value.color = color
	await updateEdge(selectedEdge.value)
}

function focusSelectedEdge() {
	if (selectedEdge.value === null || activeBoard.value === null || canvasRef.value === null) {
		return
	}

	const source = getNodeAnchor(selectedEdge.value.sourceNodeId, getResolvedSourceHandle(selectedEdge.value))
	const target = getNodeAnchor(selectedEdge.value.targetNodeId, getResolvedTargetHandle(selectedEdge.value))
	const midpoint = {
		x: (source.x + target.x) / 2,
		y: (source.y + target.y) / 2,
	}

	const nextZoom = Math.max(currentZoom.value, 1.25)
	activeBoard.value.viewportZoom = nextZoom
	persistViewport()

	requestAnimationFrame(() => {
		if (!canvasRef.value) {
			return
		}

		canvasRef.value.scrollTo({
			left: midpoint.x * nextZoom - canvasRef.value.clientWidth / 2,
			top: midpoint.y * nextZoom - canvasRef.value.clientHeight / 2,
			behavior: 'smooth',
		})
	})
}

async function uploadNodeImage(node: IVisionBoardNode, event: Event) {
	if (activeBoard.value === null) {
		return
	}

	const input = event.target as HTMLInputElement
	const file = input.files?.[0]
	if (!file) {
		return
	}

	const [attachment] = await uploadFile(activeBoard.value.taskId, file)
	node.attachmentId = attachment.id
	nodeImageBlobUrls.value[node.id] = await attachmentService.getBlobUrl(new AttachmentModel({
		id: attachment.id,
		taskId: activeBoard.value.taskId,
	}), PREVIEW_SIZE.LG) as string
	await updateNode(node)
	input.value = ''
}

function getVideoEmbedUrl(url: string) {
	if (!url) {
		return null
	}

	try {
		const parsed = new URL(url)
		if (parsed.hostname.includes('youtu.be')) {
			const id = parsed.pathname.replace('/', '')
			return id ? `https://www.youtube.com/embed/${id}` : null
		}

		if (parsed.hostname.includes('youtube.com')) {
			const id = parsed.searchParams.get('v')
			return id ? `https://www.youtube.com/embed/${id}` : null
		}
	} catch {
		return null
	}

	return null
}

async function linkBoardTask(node: IVisionBoardNode) {
	if (activeBoard.value === null) {
		return
	}

	node.taskId = activeBoard.value.taskId
	node.title = activeBoard.value.taskTitle
	editingCardTaskNodeId.value = null
	cardSearchResults.value[node.id] = []
	cardQueries.value[node.id] = activeBoard.value.taskTitle
	await updateNode(node)
	await loadCardTask(node.taskId)
}

const searchCardTasks = useDebounceFn(async (nodeId: number, query: string) => {
	if (query.trim() === '') {
		cardSearchResults.value[nodeId] = []
		return
	}

	cardSearchResults.value[nodeId] = await taskCollectionService.getAll(
		{projectId: props.projectId},
		getTaskSearchParams(query),
	)
}, 200)

async function selectCardTask(node: IVisionBoardNode, task: ITask) {
	node.taskId = task.id
	node.title = task.title
	cardQueries.value[node.id] = task.title
	cardSearchResults.value[node.id] = []
	editingCardTaskNodeId.value = null
	cardTasks.value = {
		...cardTasks.value,
		[task.id]: task,
	}
	await updateNode(node)
}

async function createAndLinkCardTask(node: IVisionBoardNode) {
	const title = (cardQueries.value[node.id] ?? '').trim()
	if (title === '') {
		return
	}

	const created = await taskService.create(new TaskModel({
		projectId: props.projectId,
		title,
	}))

	await selectCardTask(node, created)
}

async function loadCardTask(taskId: number) {
	if (taskId <= 0 || cardTasks.value[taskId]) {
		return
	}

	cardTasks.value = {
		...cardTasks.value,
		[taskId]: await taskService.get({id: taskId}),
	}
}

async function toggleCardTaskEditor() {
	if (selectedNode.value?.kind !== 'card') {
		return
	}

	if (editingCardTaskNodeId.value === selectedNode.value.id) {
		editingCardTaskNodeId.value = null
		cardSearchResults.value[selectedNode.value.id] = []
		return
	}

	editingCardTaskNodeId.value = selectedNode.value.id
	cardQueries.value[selectedNode.value.id] = getCardTask(selectedNode.value)?.title || selectedNode.value.title
	await nextTick()
	cardTaskInputRef.value?.focus()
	cardTaskInputRef.value?.select()
}

async function startTextEdit(node: IVisionBoardNode) {
	if (node.kind !== 'text') {
		return
	}

	selectedNodeId.value = node.id
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	editingTextNodeId.value = node.id
	textDrafts.value[node.id] = node.content || node.title || ''

	await nextTick()
	const editor = textEditorRefs.get(node.id)
	editor?.focus()
	editor?.select()
}

async function finishTextEdit(node: IVisionBoardNode) {
	if (editingTextNodeId.value !== node.id) {
		return
	}

	node.content = textDrafts.value[node.id] ?? ''
	node.title = node.content
	editingTextNodeId.value = null
	selectedNodeId.value = node.id
	await updateNode(node)
}

function handleCanvasMouseDown(event: MouseEvent) {
	if (event.target instanceof HTMLElement && event.target.closest('.vision-node, .vision-node-toolbar, .vision-edge-toolbar, .vision-edge-label, .vision-board-stage__viewport-controls')) {
		return
	}

	selectedNodeId.value = null
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	editingNodeTitleId.value = null
	pendingEdge.value = null

	if (editingTextNodeId.value === null) {
		return
	}

	const node = getNodeById(editingTextNodeId.value)
	if (node) {
		void finishTextEdit(node)
	}
}

function handleGlobalMouseDown(event: MouseEvent) {
	if (editingTextNodeId.value === null) {
		return
	}

	const nodeElement = nodeRefs.get(editingTextNodeId.value)
	if (nodeElement && event.target instanceof Node && !nodeElement.contains(event.target)) {
		const node = getNodeById(editingTextNodeId.value)
		if (node) {
			void finishTextEdit(node)
		}
	}
}

function onCanvasDoubleClick(event: MouseEvent) {
	if (event.target instanceof HTMLElement && event.target.closest('.vision-node, .vision-node-toolbar, .vision-board-stage__viewport-controls')) {
		return
	}

	void addNode('text', toCanvasPoint(event))
}

function selectEdge(edgeId: number) {
	selectedEdgeId.value = edgeId
	selectedNodeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null
	edgeLabelDraft.value = selectedBoardEdges.value.find(edge => edge.id === edgeId)?.label ?? ''
}

const persistDraggedNode = useDebounceFn(async (node: IVisionBoardNode) => {
	await updateNode(node)
}, 150)

watch(
	() => ({
		draggingNodeId: draggingNodeId.value,
		resizingNodeId: resizingNodeId.value,
		editingTextNodeId: editingTextNodeId.value,
		hasPendingEdge: pendingEdge.value !== null,
	}),
	({draggingNodeId, resizingNodeId, editingTextNodeId, hasPendingEdge}) => {
		if (draggingNodeId !== null || resizingNodeId !== null || editingTextNodeId !== null || hasPendingEdge || !pendingRealtimeReload.value) {
			return
		}

		reloadBoardFromRealtime()
	},
)

function isInteractiveTarget(target: EventTarget | null) {
	return target instanceof HTMLElement && target.closest('.vision-node__interactive, .vision-node-toolbar, .vision-edge-toolbar, .vision-edge-label, .vision-board-stage__viewport-controls')
}

function getNodeMinSize(node: IVisionBoardNode) {
	if (node.kind === 'text') {
		return {width: 72, height: 44}
	}

	return {width: 140, height: 80}
}

function startDrag(node: IVisionBoardNode, event: MouseEvent) {
	if (resizingNodeId.value !== null || pendingEdge.value !== null || isInteractiveTarget(event.target)) {
		return
	}

	selectedNodeId.value = node.id
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null

	if (node.kind === 'text' && editingTextNodeId.value === node.id) {
		return
	}

	const point = toCanvasPoint(event)
	draggingNodeId.value = node.id
	document.body.style.userSelect = 'none'
	dragOffset.value = {
		x: point.x - node.x,
		y: point.y - node.y,
	}
}

function startResize(node: IVisionBoardNode, event: MouseEvent) {
	selectedNodeId.value = node.id
	selectedEdgeId.value = null
	isEditingEdgeLabel.value = false
	showEdgeColorPicker.value = false
	showNodeColorPicker.value = false
	showVideoUrlEditor.value = false
	editingNodeTitleId.value = null
	const point = toCanvasPoint(event)
	resizingNodeId.value = node.id
	document.body.style.userSelect = 'none'
	resizeOrigin.value = {
		x: point.x,
		y: point.y,
		width: node.width,
		height: node.height,
	}
}

function handlePointerMove(event: MouseEvent) {
	if (activeBoard.value === null) {
		return
	}

	if (pendingEdge.value !== null) {
		pendingEdge.value.currentPoint = toCanvasPoint(event)
		return
	}

	if (draggingNodeId.value !== null) {
		const node = getNodeById(draggingNodeId.value)
		if (!node) {
			return
		}

		const point = toCanvasPoint(event)
		node.x = Math.max(0, point.x - dragOffset.value.x)
		node.y = Math.max(0, point.y - dragOffset.value.y)
		persistDraggedNode(node)
		refreshCanvasSize()
		return
	}

	if (resizingNodeId.value !== null) {
		const node = getNodeById(resizingNodeId.value)
		if (!node) {
			return
		}

		const point = toCanvasPoint(event)
		const minSize = getNodeMinSize(node)
		node.width = Math.max(minSize.width, resizeOrigin.value.width + (point.x - resizeOrigin.value.x))
		node.height = Math.max(minSize.height, resizeOrigin.value.height + (point.y - resizeOrigin.value.y))
		persistDraggedNode(node)
		refreshCanvasSize()
	}
}

function stopPointerTracking() {
	draggingNodeId.value = null
	resizingNodeId.value = null
	pendingEdge.value = null
	document.body.style.userSelect = ''
	publishIdlePresence()
}
</script>

<style lang="scss" scoped>
.project-vision-board {
	display: flex;
	flex-direction: column;
	gap: 1rem;
	min-block-size: 0;
}

.vision-board-toolbar {
	display: flex;
	flex-wrap: wrap;
	align-items: end;
	justify-content: space-between;
	gap: 1rem;
}

.vision-board-toolbar__group {
	min-inline-size: min(100%, 18rem);
}

.vision-board-toolbar__actions {
	display: flex;
	align-items: center;
	gap: .75rem;
}

.vision-board-create,
.vision-board-stage {
	background: var(--white);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	padding: 1rem;
}

.vision-board-stage {
	display: flex;
	flex-direction: column;
	min-block-size: 0;
	overflow: hidden;
}

.vision-board-stage--fullscreen {
	position: fixed;
	inset: 1rem;
	z-index: 40;
	border-radius: 14px;
}

.vision-board-create__actions {
	display: flex;
	justify-content: flex-end;
}

.vision-board-stage__header {
	display: flex;
	flex-wrap: wrap;
	align-items: start;
	justify-content: space-between;
	gap: 1rem;
	margin-block-end: 1rem;
}

.vision-board-stage__actions {
	display: flex;
	flex-wrap: wrap;
	gap: .5rem;
}

.vision-board-stage__canvas {
	position: relative;
	block-size: 100%;
	max-block-size: 100%;
	min-block-size: 0;
	border: 0;
	border-radius: 10px;
	background: #16293a;
	overflow: auto;
	user-select: none;
}

.vision-board-stage__viewport {
	position: relative;
	display: block;
	flex: 0 0 auto;
	min-block-size: 0;
	overflow: hidden;
}

.vision-board-stage__surface-frame {
	position: relative;
	min-inline-size: 100%;
	min-block-size: 100%;
}

.vision-board-stage__surface {
	position: relative;
	transform-origin: top left;
}

.vision-board-cursor {
	position: absolute;
	z-index: 4;
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	pointer-events: none;
	transform: translate(.1rem, -.1rem);
}

.vision-board-cursor__pointer {
	filter: drop-shadow(0 2px 6px rgba(5, 7, 18, .45));
	font-size: 1rem;
	transform: rotate(45deg);
}

.vision-board-cursor__label {
	border-radius: 999px;
	box-shadow: 0 10px 24px rgba(5, 7, 18, .24);
	color: #101826;
	font-size: .72rem;
	font-weight: 700;
	line-height: 1;
	padding: .28rem .5rem;
	white-space: nowrap;
}

.vision-board-stage__edges {
	position: absolute;
	inset-block-start: 0;
	inset-inline-start: 0;
	pointer-events: none;
}

.vision-board-stage__edge-path {
	fill: none;
	stroke-linecap: round;
	stroke-width: 3;
	pointer-events: stroke;

	&--preview {
		opacity: .75;
		stroke-dasharray: 8 8;
		pointer-events: none;
	}

	&--selected {
		stroke-width: 4;
	}
}

.vision-board-stage__empty {
	position: absolute;
	inset: 0;
	display: grid;
	place-items: center;
	margin: 0;
	text-align: center;
	color: rgba(255, 255, 255, .72);
}

.vision-board-stage__viewport-controls {
	position: absolute;
	inset-inline-end: 1.25rem;
	inset-block-end: 1.25rem;
	z-index: 3;
	display: inline-flex;
	align-items: center;
	gap: .5rem;
	padding: .55rem;
	border: 1px solid rgba(166, 133, 227, .6);
	border-radius: 14px;
	background: rgba(33, 37, 58, .95);
	box-shadow: 0 12px 30px rgba(5, 7, 18, .3);
}

.vision-board-stage__viewport-button {
	display: grid;
	place-items: center;
	inline-size: 2.75rem;
	block-size: 2.75rem;
	border: 1px solid rgba(166, 133, 227, .75);
	border-radius: 10px;
	background: rgba(47, 33, 63, .88);
	color: var(--gray);
	cursor: pointer;
	font: inherit;
	font-size: 1.25rem;

	&:disabled {
		opacity: .45;
		cursor: default;
	}
}

.vision-board-stage__zoom-readout {
	min-inline-size: 4rem;
	text-align: center;
	color: var(--gray);
	font-size: 1.1rem;
	font-weight: 600;
}

.vision-edge-toolbar {
	position: absolute;
	display: inline-flex;
	align-items: center;
	gap: .45rem;
	padding: .55rem;
	border: 1px solid rgba(173, 193, 214, .24);
	border-radius: 10px;
	background: rgba(48, 56, 77, .97);
	box-shadow: 0 18px 30px rgba(5, 7, 18, .3);
	z-index: 2;
}

.vision-edge-toolbar__button {
	display: grid;
	place-items: center;
	inline-size: 2.25rem;
	block-size: 2.25rem;
	border: 1px solid rgba(166, 133, 227, .45);
	border-radius: 8px;
	background: rgba(47, 33, 63, .78);
	color: rgba(255, 255, 255, .9);
	cursor: pointer;
}

.vision-edge-toolbar__colors {
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	padding-inline-end: .25rem;
}

.vision-edge-toolbar__color {
	inline-size: 1rem;
	block-size: 1rem;
	border: 1px solid rgba(255, 255, 255, .28);
	border-radius: 999px;
	cursor: pointer;
	padding: 0;

	&.is-active {
		box-shadow: 0 0 0 2px var(--primary);
	}
}

.vision-edge-toolbar__input {
	inline-size: 10rem;
	border: 1px solid rgba(173, 193, 214, .2);
	border-radius: 8px;
	background: rgba(17, 22, 36, .35);
	color: var(--white);
	font: inherit;
	padding: .5rem .625rem;
}

.vision-node-toolbar {
	position: absolute;
	display: inline-flex;
	align-items: center;
	flex-wrap: wrap;
	gap: .45rem;
	padding: .55rem;
	border: 1px solid rgba(173, 193, 214, .24);
	border-radius: 10px;
	background: rgba(48, 56, 77, .97);
	box-shadow: 0 18px 30px rgba(5, 7, 18, .3);
	z-index: 2;
	pointer-events: auto;
}

.vision-node-toolbar__button {
	display: grid;
	place-items: center;
	inline-size: 2.25rem;
	block-size: 2.25rem;
	border: 1px solid rgba(166, 133, 227, .45);
	border-radius: 8px;
	background: rgba(47, 33, 63, .78);
	color: rgba(255, 255, 255, .9);
	cursor: pointer;
}

.vision-node-toolbar__upload {
	display: grid;
	place-items: center;

	input {
		display: none;
	}
}

.vision-node-toolbar__colors {
	display: inline-flex;
	align-items: center;
	gap: .35rem;
	padding-inline-end: .25rem;
}

.vision-node-toolbar__color {
	inline-size: 1rem;
	block-size: 1rem;
	border: 1px solid rgba(255, 255, 255, .28);
	border-radius: 999px;
	cursor: pointer;
	padding: 0;

	&.is-active {
		box-shadow: 0 0 0 2px var(--primary);
	}
}

.vision-node-toolbar__attachments {
	display: flex;
	flex-wrap: wrap;
	gap: .4rem;
	inline-size: 100%;
	padding-block-start: .2rem;
}

.vision-node-toolbar__search-results {
	display: flex;
	flex-direction: column;
	gap: .25rem;
	inline-size: min(18rem, 100%);
	max-block-size: 12rem;
	overflow: auto;
}

.vision-node-toolbar__action {
	justify-content: center;
}

.vision-node-toolbar__attachment {
	inline-size: 3rem;
	block-size: 3rem;
	border: 1px solid rgba(173, 193, 214, .24);
	border-radius: 8px;
	background: rgba(17, 22, 36, .35);
	cursor: pointer;
	overflow: hidden;
	padding: 0;
}

.vision-node-toolbar__attachment-image {
	inline-size: 100%;
	block-size: 100%;
	object-fit: cover;
}

.vision-edge-label {
	position: absolute;
	transform: translate(-50%, -50%);
	border: 0;
	background: transparent;
	color: rgba(255, 255, 255, .95);
	cursor: pointer;
	font: inherit;
	font-size: 1rem;
	font-weight: 700;
	padding: .15rem .3rem;
	text-shadow: -1px 2px 10px rgb(0,0,0),0 4px 8px rgb(0, 0, 0);
	z-index: 1;
}

.vision-edge-label--selected {
	color: var(--gray);
	text-shadow: -1px 2px 10px rgb(0,0,0),0 4px 8px rgb(0, 0, 0);
}

.vision-node {
	position: absolute;
	display: flex;
	flex-direction: column;
	gap: .5rem;
	border: 1px solid rgba(173, 193, 214, .28);
	border-radius: 8px;
	background: #30384d;
	box-shadow: 0 18px 40px rgba(5, 7, 18, .28);
	color: rgba(255, 255, 255, .92);
	padding: 1rem;
	overflow: visible;
	user-select: none;
}

.vision-node--image {
	padding: 0;
	overflow: visible;
	background: rgba(17, 22, 36, .65);
}

.vision-node--video {
	padding: 0;
	overflow: visible;
	background: rgba(17, 22, 36, .65);
}

.vision-node--editing {
	box-shadow: 0 0 0 1px rgba(166, 133, 227, .55), 0 18px 40px rgba(5, 7, 18, .28);
}

.vision-node--selected {
	box-shadow: 0 0 0 1px rgba(166, 133, 227, .4), 0 18px 40px rgba(5, 7, 18, .28);
}

.vision-node__title,
.vision-node__content {
	inline-size: 100%;
	border: 1px solid rgba(173, 193, 214, .2);
	border-radius: 6px;
	background: rgba(17, 22, 36, .35);
	color: var(--gray);
	font: inherit;
	padding: .625rem .75rem;
	user-select: text;
}

.vision-node__content {
	flex: 1;
	min-block-size: 5rem;
	resize: none;
}

.vision-node__content--editor {
	min-block-size: 100%;
	color: #fff;
}

.vision-node__text {
	display: flex;
	flex: 1;
	min-block-size: 0;
}

.vision-node__text-display {
	inline-size: 100%;
	block-size: 100%;
	overflow: auto;
	padding-block-start: .25rem;
	white-space: pre-wrap;
	word-break: break-word;
	font-size: 1.1rem;
	line-height: 1.4;
	user-select: none;
}

.vision-node__media,
.vision-node__card {
	display: flex;
	flex-direction: column;
	block-size: 100%;
	gap: .5rem;
	min-block-size: 0;
	position: relative;
	overflow: visible;
}

.vision-node--image .vision-node__media {
	inline-size: 100%;
	block-size: 100%;
	gap: 0;
	padding: 0;
}

.vision-node--video .vision-node__media {
	inline-size: 100%;
	block-size: 100%;
	gap: 0;
	padding: 0;
}

.vision-node__title-label {
	position: absolute;
	top: 0;
	left: 0;
	transform: translateY(-125%);
	margin-bottom: .5rem;
	padding: .2rem .75rem;
	border-radius: 6px;
	color: var(--gray);
	font-size: .95rem;
	font-weight: 500;
	white-space: nowrap;
	text-overflow: ellipsis;
	z-index: 10;
}

.vision-node__image,
.vision-node__video {
	inline-size: 100%;
	block-size: 100%;
	min-block-size: 8rem;
	border: 1px solid rgba(173, 193, 214, .2);
	border-radius: 6px;
	object-fit: cover;
	background: rgba(17, 22, 36, .35);
}

.vision-node__content-preview {
	margin: 0;
	color: rgba(255, 255, 255, .72);
}

.vision-node__card-preview {
	block-size: 100%;
	min-block-size: 0;
	overflow: auto;
}

.vision-node__search-results {
	display: flex;
	flex-direction: column;
	gap: .25rem;
	max-block-size: 8rem;
	overflow: auto;
}

.vision-node__search-result {
	border: 1px solid rgba(173, 193, 214, .2);
	border-radius: 6px;
	background: rgba(17, 22, 36, .3);
	color: var(--gray);
	cursor: pointer;
	font: inherit;
	padding: .375rem .5rem;
	text-align: start;
}

.vision-node__resize-handle {
	position: absolute;
	inset-inline-end: .3rem;
	inset-block-end: .3rem;
	inline-size: .9rem;
	block-size: .9rem;
	background: linear-gradient(135deg, transparent 45%, rgba(255, 255, 255, .72) 45%);
	cursor: nwse-resize;
	opacity: 0;
	pointer-events: none;
}

.vision-node__handle {
	position: absolute;
	inline-size: .85rem;
	block-size: .85rem;
	border: 2px solid #a685e3;
	border-radius: 999px;
	background: #2b0d3b;
	cursor: crosshair;
	padding: 0;
	transform: translate(-50%, -50%);
	opacity: 0;
	pointer-events: auto;
}

.vision-node__handle--top {
	inset-inline-start: 50%;
	inset-block-start: 0;
}

.vision-node__handle--left {
	inset-inline-start: 0;
	inset-block-start: 50%;
}

.vision-node__handle--right {
	inset-inline-start: 100%;
	inset-block-start: 50%;
}

.vision-node__handle--bottom {
	inset-inline-start: 50%;
	inset-block-start: 100%;
}

.vision-node--selected .vision-node__handle,
.vision-board-stage__canvas--connecting .vision-node__handle,
.vision-node--editing .vision-node__handle {
	opacity: 1;
	pointer-events: auto;
}

.vision-node--selected .vision-node__resize-handle,
.vision-node--editing .vision-node__resize-handle {
	opacity: 1;
	pointer-events: auto;
}

.vision-board-empty-state {
	margin: 0;
}
</style>
