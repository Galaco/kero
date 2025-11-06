package gui

import "github.com/AllenDang/cimgui-go/imgui"

// BeginTabBar starts a tab bar
// Returns true if the tab bar is open
func BeginTabBar(label string) bool {
	return imgui.BeginTabBar(label)
}

// BeginTabBarWithFlags starts a tab bar with flags
func BeginTabBarWithFlags(label string, flags imgui.TabBarFlags) bool {
	return imgui.BeginTabBarV(label, flags)
}

// EndTabBar ends a tab bar (must be called after BeginTabBar)
func EndTabBar() {
	imgui.EndTabBar()
}

// BeginTabItem starts a tab item (a single tab in a tab bar)
// Returns true if the tab is selected
func BeginTabItem(label string) bool {
	return imgui.BeginTabItem(label)
}

// BeginTabItemWithFlags starts a tab item with flags
func BeginTabItemWithFlags(label string, open *bool, flags imgui.TabItemFlags) bool {
	return imgui.BeginTabItemV(label, open, flags)
}

// EndTabItem ends a tab item (must be called after BeginTabItem)
func EndTabItem() {
	imgui.EndTabItem()
}
