export function setTitle(title : undefined | string) {
	document.title = (typeof title === 'undefined' || title === '')
		? 'cAImanDesk'
		: `${title} | cAImanDesk`
}
