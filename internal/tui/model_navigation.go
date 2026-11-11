package tui

// openModule shows a module on the inner tab it had last.
func (model *Model) openModule(module Module) {
	model.activeModule = module
	model.showSelectedInnerTab()
}

// openInnerTab shows an inner tab of the active module; positions the module
// does not have are ignored.
func (model *Model) openInnerTab(inner InnerTab) {
	if inner < 0 || int(inner) >= len(model.activeModule.innerTabs()) {
		return
	}
	model.innerTabs[model.activeModule] = inner
	model.showSelectedInnerTab()
}

// cycleInnerTab moves to the next (step 1) or previous (step -1) inner tab of
// the active module, wrapping around.
func (model *Model) cycleInnerTab(step int) {
	count := len(model.activeModule.innerTabs())
	next := (int(model.innerTabs[model.activeModule]) + step + count) % count
	model.openInnerTab(InnerTab(next))
}

// showSelectedInnerTab points the content area at the remembered inner tab.
func (model *Model) showSelectedInnerTab() {
	spec := model.activeModule.innerTabs()[model.innerTabs[model.activeModule]]
	model.switchTab(spec.view)
}

// toggleSidebar hides or shows the sidebar on top of the responsive mode.
func (model *Model) toggleSidebar() {
	model.sidebarHidden = !model.sidebarHidden
	model.ensureCursorVisible()
}
