import AbstractService from '@/services/abstractService'
import AdminTaskModel from '@/models/adminTask'
import type {IAdminTask} from '@/modelTypes/IAdminTask'

export default class AdminTaskService extends AbstractService<IAdminTask> {
	constructor() {
		super({
			getAll: '/admin/tasks',
		})
	}

	modelFactory(data: Partial<IAdminTask>) {
		return new AdminTaskModel(data)
	}
}
