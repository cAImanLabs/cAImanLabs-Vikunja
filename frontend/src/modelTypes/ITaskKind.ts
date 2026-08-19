// Mirrors the backend's models.TaskKind string enum (pkg/models/task_kind.go).
export const TASK_KINDS = {
	TASK: 'task',
	EPIC: 'epic',
	STORY: 'story',
	BUG: 'bug',
	SUBTASK: 'subtask',
	FEATURE: 'feature',
} as const
export type TaskKind = typeof TASK_KINDS[keyof typeof TASK_KINDS]

// Display order for kind pickers - keeps Task (the default/most common) first.
export const TASK_KIND_ORDER: TaskKind[] = [
	TASK_KINDS.TASK,
	TASK_KINDS.EPIC,
	TASK_KINDS.STORY,
	TASK_KINDS.BUG,
	TASK_KINDS.SUBTASK,
	TASK_KINDS.FEATURE,
]
