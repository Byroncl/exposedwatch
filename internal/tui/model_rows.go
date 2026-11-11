package tui

// switchTab shows another tab with the cursor back on its first row.
func (model *Model) switchTab(tab TabID) {
	if tab == model.activeTab {
		return
	}
	model.activeTab = tab
	model.visibleRows = nil
	model.cursor = 0
	model.scrollOffset = scrollOffsetFloor
	model.rebuildVisibleRows()
}

// rebuildVisibleRows applies filters and sort to the active tab, keeping the
// selected row when it still exists.
func (model *Model) rebuildVisibleRows() {
	var selected *rowKey
	if model.cursor >= 0 && model.cursor < len(model.visibleRows) {
		current := model.visibleRows[model.cursor].key
		selected = &current
	}
	rows := model.buildRows(model.activeTab)
	if model.activeTab.supportsFiltering() {
		filtered := make([]displayRow, 0, len(rows))
		for _, row := range rows {
			if row.meetsMinimum(model.minSeverity) && row.matchesText(model.filterText) {
				filtered = append(filtered, row)
			}
		}
		sortRows(filtered, model.sortMode)
		rows = filtered
	}
	model.visibleRows = rows
	model.cursor = 0
	if selected != nil {
		for index, row := range rows {
			if row.key == *selected {
				model.cursor = index
				break
			}
		}
	}
	model.ensureCursorVisible()
}

func (model *Model) moveCursor(delta int) {
	model.setCursor(model.cursor + delta)
}

func (model *Model) setCursor(position int) {
	last := len(model.visibleRows) - 1
	if position > last {
		position = last
	}
	if position < 0 {
		position = 0
	}
	model.cursor = position
	model.ensureCursorVisible()
}

// ensureCursorVisible scrolls so the cursor row is inside the table window.
func (model *Model) ensureCursorVisible() {
	capacity := model.computeLayout().tableRows
	if capacity < 1 {
		model.scrollOffset = scrollOffsetFloor
		return
	}
	if model.cursor < model.scrollOffset {
		model.scrollOffset = model.cursor
	}
	if model.cursor >= model.scrollOffset+capacity {
		model.scrollOffset = model.cursor - capacity + 1
	}
	if model.scrollOffset < scrollOffsetFloor {
		model.scrollOffset = scrollOffsetFloor
	}
}
