import TaskModel from '@/models/task'
import type {IAdminTask} from '@/modelTypes/IAdminTask'

export default class AdminTaskModel extends TaskModel implements IAdminTask {
	declare projectTitle: string

	constructor(data: Partial<IAdminTask> = {}) {
		super(data)
	}
}
