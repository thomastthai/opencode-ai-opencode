package dialog

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// MockProvider for testing
type MockProvider struct {
	items []CompletionItemI
}

func (m *MockProvider) GetId() string {
	return "mock"
}

func (m *MockProvider) GetEntry() CompletionItemI {
	return &CompletionItem{Title: "Mock", Value: "mock"}
}

func (m *MockProvider) GetChildEntries(query string) ([]CompletionItemI, error) {
	return m.items, nil
}

func TestCompletionDialog_VimNavigation(t *testing.T) {
	// Create test items
	items := []CompletionItemI{
		&CompletionItem{Title: "Item 1", Value: "item1"},
		&CompletionItem{Title: "Item 2", Value: "item2"},
		&CompletionItem{Title: "Item 3", Value: "item3"},
	}
	
	provider := &MockProvider{items: items}
	dialog := NewCompletionDialogCmp(provider)
	
	// Initialize with slash command to activate dialog
	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	
	// Get initial selection
	initialItem, initialIdx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 0, initialIdx, "Should start at first item")
	assert.Equal(t, "item1", initialItem.GetValue())
	
	// Test j key (down)
	jKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = dialog.Update(jKey)
	
	afterJItem, afterJIdx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 1, afterJIdx, "Should move down with j key")
	assert.Equal(t, "item2", afterJItem.GetValue())
	
	// Test k key (up)
	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = dialog.Update(kKey)
	
	afterKItem, afterKIdx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 0, afterKIdx, "Should move up with k key")
	assert.Equal(t, "item1", afterKItem.GetValue())
	
	// Test arrow keys still work
	downKey := tea.KeyMsg{Type: tea.KeyDown}
	_, _ = dialog.Update(downKey)
	
	afterDownItem, afterDownIdx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 1, afterDownIdx, "Should move down with arrow key")
	assert.Equal(t, "item2", afterDownItem.GetValue())
	
	upKey := tea.KeyMsg{Type: tea.KeyUp}
	_, _ = dialog.Update(upKey)
	
	afterUpItem, afterUpIdx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 0, afterUpIdx, "Should move up with arrow key")
	assert.Equal(t, "item1", afterUpItem.GetValue())
}

func TestCompletionDialog_VimNavigationBounds(t *testing.T) {
	// Create test items
	items := []CompletionItemI{
		&CompletionItem{Title: "Item 1", Value: "item1"},
		&CompletionItem{Title: "Item 2", Value: "item2"},
	}
	
	provider := &MockProvider{items: items}
	dialog := NewCompletionDialogCmp(provider)
	
	// Initialize with slash command
	_, _ = dialog.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	
	// Try to go up from first item
	kKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = dialog.Update(kKey)
	
	item, idx := dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 0, idx, "Should stay at first item when trying to go up")
	assert.Equal(t, "item1", item.GetValue())
	
	// Go to last item
	jKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = dialog.Update(jKey)
	
	// Try to go down from last item
	_, _ = dialog.Update(jKey)
	
	item, idx = dialog.(*completionDialogCmp).listView.GetSelectedItem()
	assert.Equal(t, 1, idx, "Should stay at last item when trying to go down")
	assert.Equal(t, "item2", item.GetValue())
}

func TestCompletionDialog_KeyBindings(t *testing.T) {
	provider := &MockProvider{items: []CompletionItemI{}}
	dialog := NewCompletionDialogCmp(provider)
	
	// Get binding keys from the dialog
	bindings := dialog.BindingKeys()
	
	// Check that we have some key bindings
	assert.Greater(t, len(bindings), 0, "Should have key bindings")
	
	// Check for specific keys we expect
	hasTab := false
	hasEsc := false
	
	for _, binding := range bindings {
		for _, k := range binding.Keys() {
			if k == "tab" {
				hasTab = true
			}
			if k == "esc" {
				hasEsc = true
			}
		}
	}
	
	assert.True(t, hasTab, "Should have tab key binding")
	assert.True(t, hasEsc, "Should have escape key binding")
}