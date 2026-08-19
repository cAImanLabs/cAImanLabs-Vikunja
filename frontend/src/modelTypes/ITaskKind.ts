// Mirrors the backend's models.TaskKind string enum (pkg/models/task_kind.go).
export const TASK_KINDS = {
	TASK: 'task',
	EPIC: 'epic',
	STORY: 'story',
	BUG: 'bug',
	SUBTASK: 'subtask',
	FEATURE: 'feature',
	INITIATIVE: 'initiative',
} as const
export type TaskKind = typeof TASK_KINDS[keyof typeof TASK_KINDS]

// Display order for kind pickers - top of the 5-level hierarchy down to the
// bottom, with the level-1 base issues grouped together in their usual order.
export const TASK_KIND_ORDER: TaskKind[] = [
	TASK_KINDS.INITIATIVE,
	TASK_KINDS.FEATURE,
	TASK_KINDS.EPIC,
	TASK_KINDS.TASK,
	TASK_KINDS.STORY,
	TASK_KINDS.BUG,
	TASK_KINDS.SUBTASK,
]

// Mirrors the backend's TaskKind.TaskKindLevel (pkg/models/task_kind.go): the
// 5-level SAFe-style hierarchy this app enforces relations against.
//   4 Initiative  - strategic, cross-team goals spanning quarters
//   3 Feature     - major deliverables under an initiative, spanning sprints
//   2 Epic        - a targeted set of work under a feature, one team's backlog
//   1 Story/Task/Bug - the day-to-day work done in a single sprint
//   0 Subtask     - granular steps to complete a single base issue
export const TASK_KIND_LEVEL: Record<TaskKind, number> = {
	[TASK_KINDS.SUBTASK]: 0,
	[TASK_KINDS.TASK]: 1,
	[TASK_KINDS.STORY]: 1,
	[TASK_KINDS.BUG]: 1,
	[TASK_KINDS.EPIC]: 2,
	[TASK_KINDS.FEATURE]: 3,
	[TASK_KINDS.INITIATIVE]: 4,
}

// Kinds whose level is a "container" (2 and above) - assigning one of these
// should nudge the user to add at least one child task via Related Tasks.
export const TASK_KINDS_REQUESTING_CHILDREN: TaskKind[] = [
	TASK_KINDS.EPIC,
	TASK_KINDS.FEATURE,
	TASK_KINDS.INITIATIVE,
]
