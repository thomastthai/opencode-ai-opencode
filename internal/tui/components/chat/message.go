package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/opencode-ai/opencode/internal/config"
	"github.com/opencode-ai/opencode/internal/diff"
	"github.com/opencode-ai/opencode/internal/llm/agent"
	"github.com/opencode-ai/opencode/internal/llm/models"
	"github.com/opencode-ai/opencode/internal/llm/tools"
	"github.com/opencode-ai/opencode/internal/message"
	"github.com/opencode-ai/opencode/internal/tui/styles"
	"github.com/opencode-ai/opencode/internal/tui/theme"
)

type uiMessageType int

const (
	userMessageType uiMessageType = iota
	assistantMessageType
	toolMessageType

	maxResultHeight = 10
)

type uiMessage struct {
	ID          string
	messageType uiMessageType
	position    int
	height      int
	content     string
}

// isValidHexColor validates if a string is a valid hex color format
func isValidHexColor(hex string) bool {
	// Match #RRGGBB or #RGB format
	hexRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	return hexRegex.MatchString(hex)
}

// getEffectiveColors returns the effective colors to use for messages
// Either from hex config overrides or from the theme
func getEffectiveColors(cfg *config.Config, t theme.Theme, isUser bool) (textColor lipgloss.TerminalColor, bgColor lipgloss.TerminalColor) {
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging {
		if isUser {
			// User message colors
			if cfg.TUI.MessageLayoutConfig.UserTextColor != "" && isValidHexColor(cfg.TUI.MessageLayoutConfig.UserTextColor) {
				textColor = lipgloss.Color(cfg.TUI.MessageLayoutConfig.UserTextColor)
			} else {
				textColor = t.Text()
			}
			
			if cfg.TUI.MessageLayoutConfig.UserBackgroundColor != "" && isValidHexColor(cfg.TUI.MessageLayoutConfig.UserBackgroundColor) {
				bgColor = lipgloss.Color(cfg.TUI.MessageLayoutConfig.UserBackgroundColor)
			} else if cfg.TUI.MessageLayoutConfig.UseBackgrounds {
				bgColor = t.UserMessageBackground()
			} else {
				bgColor = t.Background()
			}
		} else {
			// Assistant message colors
			if cfg.TUI.MessageLayoutConfig.AssistantTextColor != "" && isValidHexColor(cfg.TUI.MessageLayoutConfig.AssistantTextColor) {
				textColor = lipgloss.Color(cfg.TUI.MessageLayoutConfig.AssistantTextColor)
			} else {
				textColor = t.TextMuted()
			}
			
			if cfg.TUI.MessageLayoutConfig.AssistantBackgroundColor != "" && isValidHexColor(cfg.TUI.MessageLayoutConfig.AssistantBackgroundColor) {
				bgColor = lipgloss.Color(cfg.TUI.MessageLayoutConfig.AssistantBackgroundColor)
			} else if cfg.TUI.MessageLayoutConfig.UseBackgrounds {
				bgColor = t.AssistantMessageBackground()
			} else {
				bgColor = t.Background()
			}
		}
	} else {
		// Classic layout - use theme colors
		if isUser {
			textColor = t.Text()
		} else {
			textColor = t.TextMuted()
		}
		bgColor = t.Background()
	}
	
	return textColor, bgColor
}

// getBorderChar returns the effective border character for a specific type
// Takes into account which sides are enabled for intelligent corner selection
func getBorderChar(borderConfig *config.BorderConfig, charType string, sides *config.BorderSides) string {
	switch charType {
	case "vertical":
		if borderConfig.Character != "" {
			return borderConfig.Character
		}
		return "│"
	case "horizontal":
		if borderConfig.HorizontalChar != "" {
			return borderConfig.HorizontalChar
		}
		return "─"
	case "topLeft":
		if borderConfig.TopLeftChar != "" {
			return borderConfig.TopLeftChar
		}
		// Use intelligent default based on enabled sides
		if sides.Top && sides.Left {
			return "┌"
		} else if sides.Top && !sides.Left {
			return "─"
		} else if !sides.Top && sides.Left {
			return "│"
		}
		return "┌"
	case "topRight":
		if borderConfig.TopRightChar != "" {
			return borderConfig.TopRightChar
		}
		// Use intelligent default based on enabled sides
		if sides.Top && sides.Right {
			return "┐"
		} else if sides.Top && !sides.Right {
			return "─"
		} else if !sides.Top && sides.Right {
			return "│"
		}
		return "┐"
	case "bottomLeft":
		if borderConfig.BottomLeftChar != "" {
			return borderConfig.BottomLeftChar
		}
		// Use intelligent default based on enabled sides
		if sides.Bottom && sides.Left {
			return "└"
		} else if sides.Bottom && !sides.Left {
			return "─"
		} else if !sides.Bottom && sides.Left {
			return "│"
		}
		return "└"
	case "bottomRight":
		if borderConfig.BottomRightChar != "" {
			return borderConfig.BottomRightChar
		}
		// Use intelligent default based on enabled sides
		if sides.Bottom && sides.Right {
			return "┘"
		} else if sides.Bottom && !sides.Right {
			return "─"
		} else if !sides.Bottom && sides.Right {
			return "│"
		}
		return "┘"
	default:
		return "│"
	}
}

// getBorderColors returns the effective colors for a border
func getBorderColors(borderConfig *config.BorderConfig, t theme.Theme, isUser bool, bgColor lipgloss.TerminalColor) (fg lipgloss.TerminalColor, bg lipgloss.TerminalColor) {
	// Foreground color
	if borderConfig.ForegroundColor != "" && isValidHexColor(borderConfig.ForegroundColor) {
		fg = lipgloss.Color(borderConfig.ForegroundColor)
	} else {
		if isUser {
			fg = t.Secondary()
		} else {
			fg = t.Primary()
		}
	}
	
	// Background color
	if borderConfig.BackgroundColor != "" && isValidHexColor(borderConfig.BackgroundColor) {
		bg = lipgloss.Color(borderConfig.BackgroundColor)
	} else {
		bg = bgColor
	}
	
	return fg, bg
}

// createFullBorder creates a complete border around the content
func createFullBorder(content []string, borderConfig *config.BorderConfig, t theme.Theme, isUser bool, bgColor lipgloss.TerminalColor, contentWidth int) []string {
	if borderConfig == nil {
		return content
	}
	
	sides := borderConfig.Sides
	if !sides.Top && !sides.Right && !sides.Bottom && !sides.Left {
		return content
	}
	
	fg, bg := getBorderColors(borderConfig, t, isUser, bgColor)
	borderStyle := lipgloss.NewStyle().Foreground(fg).Background(bg)
	spacingStyle := lipgloss.NewStyle().Background(bgColor)
	
	var result []string
	
	// Calculate the actual content width (max width of content lines)
	maxContentWidth := 0
	for _, line := range content {
		if w := lipgloss.Width(line); w > maxContentWidth {
			maxContentWidth = w
		}
	}
	
	// The content width should be exactly what we have, no extra padding
	contentWidth = maxContentWidth
	
	// Top border - spans full width including corners
	if sides.Top {
		var topLine strings.Builder
		
		if sides.Left {
			topLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "topLeft", &sides)))
		} else {
			// Add spacing if no left border but we have top border
			topLine.WriteString(spacingStyle.Render(" "))
		}
		
		// Horizontal line spans the content width
		horizontalChar := getBorderChar(borderConfig, "horizontal", &sides)
		topLine.WriteString(borderStyle.Render(strings.Repeat(horizontalChar, contentWidth)))
		
		if sides.Right {
			topLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "topRight", &sides)))
		} else {
			// Add spacing if no right border but we have top border
			topLine.WriteString(spacingStyle.Render(" "))
		}
		
		result = append(result, topLine.String())
	}
	
	// Content lines with left/right borders
	for _, line := range content {
		var contentLine strings.Builder
		
		if sides.Left {
			contentLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "vertical", &sides)))
		} else {
			// Add consistent spacing if no left border (to match top/bottom border spacing)
			contentLine.WriteString(spacingStyle.Render(" "))
		}
		
		// Add content without extra padding since spacing is handled in renderMessage
		contentLine.WriteString(line)
		
		if sides.Right {
			contentLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "vertical", &sides)))
		}
		
		result = append(result, contentLine.String())
	}
	
	// Bottom border - spans full width including corners
	if sides.Bottom {
		var bottomLine strings.Builder
		
		if sides.Left {
			bottomLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "bottomLeft", &sides)))
		} else {
			// Add spacing if no left border but we have bottom border
			bottomLine.WriteString(spacingStyle.Render(" "))
		}
		
		// Horizontal line spans the content width
		horizontalChar := getBorderChar(borderConfig, "horizontal", &sides)
		bottomLine.WriteString(borderStyle.Render(strings.Repeat(horizontalChar, contentWidth)))
		
		if sides.Right {
			bottomLine.WriteString(borderStyle.Render(getBorderChar(borderConfig, "bottomRight", &sides)))
		} else {
			// Add spacing if no right border but we have bottom border
			bottomLine.WriteString(spacingStyle.Render(" "))
		}
		
		result = append(result, bottomLine.String())
	}
	
	return result
}

func toMarkdown(content string, focused bool, width int) string {
	r := styles.GetMarkdownRenderer(width)
	rendered, _ := r.Render(content)
	return rendered
}

func renderMessage(msg string, isUser bool, isFocused bool, width int, info ...string) string {
	t := theme.CurrentTheme()
	cfg := config.Get()
	
	var style lipgloss.Style
	var messageWidth int
	var leftPadding int
	var rightPadding int
	
	// Get effective colors (either from hex config or theme)
	textColor, bgColor := getEffectiveColors(cfg, t, isUser)
	
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging {
		// Modern messaging app layout
		if isUser {
			// User messages: right-aligned, narrower, with background
			messageWidth = int(float64(width) * cfg.TUI.MessageLayoutConfig.UserMessageWidth)
			rightPadding = cfg.TUI.MessageLayoutConfig.UserRightMargin
			leftPadding = width - messageWidth - rightPadding
			if leftPadding < 0 {
				leftPadding = 0
				messageWidth = width - rightPadding
			}
			
			// Create the message style without margin using effective colors
			style = styles.BaseStyle().
				Foreground(textColor).
				Background(bgColor)
				// Don't set Width() or Padding() on style as it causes gray background fill
				
			// Remove lipgloss borders - they cause gray background issues
			// We'll add manual border characters instead
		} else {
			// Assistant messages: left-aligned, wider
			messageWidth = int(float64(width) * cfg.TUI.MessageLayoutConfig.AssistantMessageWidth)
			leftPadding = cfg.TUI.MessageLayoutConfig.AssistantLeftMargin
			
			// Create the message style without margin using effective colors
			style = styles.BaseStyle().
				Foreground(textColor).
				Background(bgColor)
				// Temporarily remove left border to test gray background
				// BorderLeft(true).
				// BorderForeground(t.Primary()).
				// BorderStyle(lipgloss.ThickBorder())
				// Don't set Width() or Padding() on style as it causes gray background fill
		}
	} else {
		// Classic layout (current behavior)
		style = styles.BaseStyle().
			Width(width - 1).
			BorderLeft(true).
			Foreground(textColor).
			BorderForeground(t.Primary()).
			BorderStyle(lipgloss.ThickBorder()).
			Background(bgColor)

		if isUser {
			style = style.BorderForeground(t.Secondary())
		}
		messageWidth = width - 1
		leftPadding = 0
	}
	
	// Add manual padding by adjusting the content and width
	contentWidth := messageWidth
	paddedContent := toMarkdown(msg, isFocused, contentWidth)
	
	// Add borders around the content for messaging layout
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging {
		lines := strings.Split(paddedContent, "\n")
		var borderConfig *config.BorderConfig
		
		if isUser {
			borderConfig = &cfg.TUI.MessageLayoutConfig.UserBorder
		} else {
			borderConfig = &cfg.TUI.MessageLayoutConfig.AssistantBorder
		}
		
		// Remove trailing empty lines first
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
		
		// Add spacing to content lines
		spacingStyle := lipgloss.NewStyle().Background(bgColor)
		for i, line := range lines {
			if strings.TrimSpace(line) != "" {
				// Add padding space for better appearance
				lines[i] = spacingStyle.Render(" ") + line + spacingStyle.Render(" ")
			} else {
				// Handle empty lines properly - use same width as content lines
				lines[i] = spacingStyle.Render("  ")
			}
		}
		
		// Apply full borders around the content
		// createFullBorder will use the actual content width automatically
		borderedLines := createFullBorder(lines, borderConfig, t, isUser, bgColor, 0)
		paddedContent = strings.Join(borderedLines, "\n")
	}
	
	parts := []string{
		paddedContent,
	}

	// Remove newline at the end
	parts[0] = strings.TrimSuffix(parts[0], "\n")
	if len(info) > 0 {
		parts = append(parts, info...)
	}

	rendered := style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			parts...,
		),
	)


	// For messaging layout, wrap in a full-width container with proper background
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging && (leftPadding > 0 || rightPadding > 0) {
		// Create a full-width container
		containerStyle := lipgloss.NewStyle().
			Width(width).
			Background(t.Background())
		
		if isUser {
			// For user messages, padding goes on both left and right
			parts := []string{}
			
			// Add left padding if needed
			if leftPadding > 0 {
				leftPaddingStyle := lipgloss.NewStyle().
					Width(leftPadding).
					Height(lipgloss.Height(rendered)).
					Background(t.Background())
				parts = append(parts, leftPaddingStyle.Render(strings.Repeat(" ", leftPadding)))
			}
			
			// Add the message
			parts = append(parts, rendered)
			
			// Add right padding if needed
			if rightPadding > 0 {
				rightPaddingStyle := lipgloss.NewStyle().
					Width(rightPadding).
					Height(lipgloss.Height(rendered)).
					Background(t.Background())
				parts = append(parts, rightPaddingStyle.Render(strings.Repeat(" ", rightPadding)))
			}
			
			rendered = containerStyle.Render(
				lipgloss.JoinHorizontal(lipgloss.Top, parts...),
			)
		} else {
			// For assistant messages, only left padding
			if leftPadding > 0 {
				leftPaddingStyle := lipgloss.NewStyle().
					Width(leftPadding).
					Height(lipgloss.Height(rendered)).
					Background(t.Background())
				
				rendered = containerStyle.Render(
					lipgloss.JoinHorizontal(
						lipgloss.Top,
						leftPaddingStyle.Render(strings.Repeat(" ", leftPadding)),
						rendered,
					),
				)
			}
		}
	}

	return rendered
}

func renderUserMessage(msg message.Message, isFocused bool, width int, position int) uiMessage {
	var styledAttachments []string
	t := theme.CurrentTheme()
	attachmentStyles := styles.BaseStyle().
		MarginLeft(1).
		Background(t.TextMuted()).
		Foreground(t.Text())
	for _, attachment := range msg.BinaryContent() {
		file := filepath.Base(attachment.Path)
		var filename string
		if len(file) > 10 {
			filename = fmt.Sprintf(" %s %s...", styles.DocumentIcon, file[0:7])
		} else {
			filename = fmt.Sprintf(" %s %s", styles.DocumentIcon, file)
		}
		styledAttachments = append(styledAttachments, attachmentStyles.Render(filename))
	}
	content := ""
	if len(styledAttachments) > 0 {
		attachmentContent := styles.BaseStyle().Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Left, styledAttachments...))
		content = renderMessage(msg.Content().String(), true, isFocused, width, attachmentContent)
	} else {
		content = renderMessage(msg.Content().String(), true, isFocused, width)
	}
	userMsg := uiMessage{
		ID:          msg.ID,
		messageType: userMessageType,
		position:    position,
		height:      lipgloss.Height(content),
		content:     content,
	}
	return userMsg
}

// Returns multiple uiMessages because of the tool calls
func renderAssistantMessage(
	msg message.Message,
	msgIndex int,
	allMessages []message.Message, // we need this to get tool results and the user message
	messagesService message.Service, // We need this to get the task tool messages
	focusedUIMessageId string,
	isSummary bool,
	width int,
	position int,
) []uiMessage {
	messages := []uiMessage{}
	content := msg.Content().String()
	thinking := msg.IsThinking()
	thinkingContent := msg.ReasoningContent().Thinking
	finished := msg.IsFinished()
	finishData := msg.FinishPart()
	info := []string{}

	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Add finish info if available
	if finished && config.Get().TUI.ShowModelInfo {
		switch finishData.Reason {
		case message.FinishReasonEndTurn:
			took := formatTimestampDiff(msg.CreatedAt, finishData.Time)
			info = append(info, baseStyle.
				Width(width-1).
				Foreground(t.TextMuted()).
				Render(fmt.Sprintf(" %s (%s)", models.SupportedModels[msg.Model].Name, took)),
			)
		case message.FinishReasonCanceled:
			info = append(info, baseStyle.
				Width(width-1).
				Foreground(t.TextMuted()).
				Render(fmt.Sprintf(" %s (%s)", models.SupportedModels[msg.Model].Name, "canceled")),
			)
		case message.FinishReasonError:
			info = append(info, baseStyle.
				Width(width-1).
				Foreground(t.TextMuted()).
				Render(fmt.Sprintf(" %s (%s)", models.SupportedModels[msg.Model].Name, "error")),
			)
		case message.FinishReasonPermissionDenied:
			info = append(info, baseStyle.
				Width(width-1).
				Foreground(t.TextMuted()).
				Render(fmt.Sprintf(" %s (%s)", models.SupportedModels[msg.Model].Name, "permission denied")),
			)
		}
	}
	if content != "" || (finished && finishData.Reason == message.FinishReasonEndTurn) {
		if content == "" {
			content = "*Finished without output*"
		}
		if isSummary {
			info = append(info, baseStyle.Width(width-1).Foreground(t.TextMuted()).Render(" (summary)"))
		}

		content = renderMessage(content, false, true, width, info...)
		messages = append(messages, uiMessage{
			ID:          msg.ID,
			messageType: assistantMessageType,
			position:    position,
			height:      lipgloss.Height(content),
			content:     content,
		})
		position += messages[0].height
		position++ // for the space
	} else if thinking && thinkingContent != "" {
		// Render the thinking content
		content = renderMessage(thinkingContent, false, msg.ID == focusedUIMessageId, width)
	}

	for i, toolCall := range msg.ToolCalls() {
		toolCallContent := renderToolMessage(
			toolCall,
			allMessages,
			messagesService,
			focusedUIMessageId,
			false,
			width,
			i+1,
		)
		messages = append(messages, toolCallContent)
		position += toolCallContent.height
		position++ // for the space
	}
	return messages
}

func findToolResponse(toolCallID string, futureMessages []message.Message) *message.ToolResult {
	for _, msg := range futureMessages {
		for _, result := range msg.ToolResults() {
			if result.ToolCallID == toolCallID {
				return &result
			}
		}
	}
	return nil
}

func toolName(name string) string {
	switch name {
	case agent.AgentToolName:
		return "Task"
	case tools.BashToolName:
		return "Bash"
	case tools.EditToolName:
		return "Edit"
	case tools.FetchToolName:
		return "Fetch"
	case tools.GlobToolName:
		return "Glob"
	case tools.GrepToolName:
		return "Grep"
	case tools.LSToolName:
		return "List"
	case tools.SourcegraphToolName:
		return "Sourcegraph"
	case tools.ViewToolName:
		return "View"
	case tools.WriteToolName:
		return "Write"
	case tools.PatchToolName:
		return "Patch"
	}
	return name
}

func getToolAction(name string) string {
	switch name {
	case agent.AgentToolName:
		return "Preparing prompt..."
	case tools.BashToolName:
		return "Building command..."
	case tools.EditToolName:
		return "Preparing edit..."
	case tools.FetchToolName:
		return "Writing fetch..."
	case tools.GlobToolName:
		return "Finding files..."
	case tools.GrepToolName:
		return "Searching content..."
	case tools.LSToolName:
		return "Listing directory..."
	case tools.SourcegraphToolName:
		return "Searching code..."
	case tools.ViewToolName:
		return "Reading file..."
	case tools.WriteToolName:
		return "Preparing write..."
	case tools.PatchToolName:
		return "Preparing patch..."
	}
	return "Working..."
}

// renders params, params[0] (params[1]=params[2] ....)
func renderParams(paramsWidth int, params ...string) string {
	if len(params) == 0 {
		return ""
	}
	mainParam := params[0]
	if len(mainParam) > paramsWidth {
		mainParam = mainParam[:paramsWidth-3] + "..."
	}

	if len(params) == 1 {
		return mainParam
	}
	otherParams := params[1:]
	// create pairs of key/value
	// if odd number of params, the last one is a key without value
	if len(otherParams)%2 != 0 {
		otherParams = append(otherParams, "")
	}
	parts := make([]string, 0, len(otherParams)/2)
	for i := 0; i < len(otherParams); i += 2 {
		key := otherParams[i]
		value := otherParams[i+1]
		if value == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}

	partsRendered := strings.Join(parts, ", ")
	remainingWidth := paramsWidth - lipgloss.Width(partsRendered) - 5 // for the space
	if remainingWidth < 30 {
		// No space for the params, just show the main
		return mainParam
	}

	if len(parts) > 0 {
		mainParam = fmt.Sprintf("%s (%s)", mainParam, strings.Join(parts, ", "))
	}

	return ansi.Truncate(mainParam, paramsWidth, "...")
}

func removeWorkingDirPrefix(path string) string {
	wd := config.WorkingDirectory()
	if strings.HasPrefix(path, wd) {
		path = strings.TrimPrefix(path, wd)
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	if strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	if strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	return path
}

func renderToolParams(paramWidth int, toolCall message.ToolCall) string {
	params := ""
	switch toolCall.Name {
	case agent.AgentToolName:
		var params agent.AgentParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		prompt := strings.ReplaceAll(params.Prompt, "\n", " ")
		return renderParams(paramWidth, prompt)
	case tools.BashToolName:
		var params tools.BashParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		command := strings.ReplaceAll(params.Command, "\n", " ")
		return renderParams(paramWidth, command)
	case tools.EditToolName:
		var params tools.EditParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		filePath := removeWorkingDirPrefix(params.FilePath)
		return renderParams(paramWidth, filePath)
	case tools.FetchToolName:
		var params tools.FetchParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		url := params.URL
		toolParams := []string{
			url,
		}
		if params.Format != "" {
			toolParams = append(toolParams, "format", params.Format)
		}
		if params.Timeout != 0 {
			toolParams = append(toolParams, "timeout", (time.Duration(params.Timeout) * time.Second).String())
		}
		return renderParams(paramWidth, toolParams...)
	case tools.GlobToolName:
		var params tools.GlobParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		pattern := params.Pattern
		toolParams := []string{
			pattern,
		}
		if params.Path != "" {
			toolParams = append(toolParams, "path", params.Path)
		}
		return renderParams(paramWidth, toolParams...)
	case tools.GrepToolName:
		var params tools.GrepParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		pattern := params.Pattern
		toolParams := []string{
			pattern,
		}
		if params.Path != "" {
			toolParams = append(toolParams, "path", params.Path)
		}
		if params.Include != "" {
			toolParams = append(toolParams, "include", params.Include)
		}
		if params.LiteralText {
			toolParams = append(toolParams, "literal", "true")
		}
		return renderParams(paramWidth, toolParams...)
	case tools.LSToolName:
		var params tools.LSParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		path := params.Path
		if path == "" {
			path = "."
		}
		return renderParams(paramWidth, path)
	case tools.SourcegraphToolName:
		var params tools.SourcegraphParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		return renderParams(paramWidth, params.Query)
	case tools.ViewToolName:
		var params tools.ViewParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		filePath := removeWorkingDirPrefix(params.FilePath)
		toolParams := []string{
			filePath,
		}
		if params.Limit != 0 {
			toolParams = append(toolParams, "limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Offset != 0 {
			toolParams = append(toolParams, "offset", fmt.Sprintf("%d", params.Offset))
		}
		return renderParams(paramWidth, toolParams...)
	case tools.WriteToolName:
		var params tools.WriteParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		filePath := removeWorkingDirPrefix(params.FilePath)
		return renderParams(paramWidth, filePath)
	default:
		input := strings.ReplaceAll(toolCall.Input, "\n", " ")
		params = renderParams(paramWidth, input)
	}
	return params
}

func truncateHeight(content string, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		return strings.Join(lines[:height], "\n")
	}
	return content
}

func renderToolResponse(toolCall message.ToolCall, response message.ToolResult, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	if response.IsError {
		errContent := fmt.Sprintf("Error: %s", strings.ReplaceAll(response.Content, "\n", " "))
		errContent = ansi.Truncate(errContent, width-1, "...")
		return baseStyle.
			Width(width).
			Foreground(t.Error()).
			Render(errContent)
	}

	resultContent := truncateHeight(response.Content, maxResultHeight)
	switch toolCall.Name {
	case agent.AgentToolName:
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, false, width),
			t.Background(),
		)
	case tools.BashToolName:
		resultContent = fmt.Sprintf("```bash\n%s\n```", resultContent)
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, true, width),
			t.Background(),
		)
	case tools.EditToolName:
		metadata := tools.EditResponseMetadata{}
		json.Unmarshal([]byte(response.Metadata), &metadata)
		truncDiff := truncateHeight(metadata.Diff, maxResultHeight)
		formattedDiff, _ := diff.FormatDiff(truncDiff, diff.WithTotalWidth(width))
		return formattedDiff
	case tools.FetchToolName:
		var params tools.FetchParams
		json.Unmarshal([]byte(toolCall.Input), &params)
		mdFormat := "markdown"
		switch params.Format {
		case "text":
			mdFormat = "text"
		case "html":
			mdFormat = "html"
		}
		resultContent = fmt.Sprintf("```%s\n%s\n```", mdFormat, resultContent)
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, true, width),
			t.Background(),
		)
	case tools.GlobToolName:
		return baseStyle.Width(width).Foreground(t.TextMuted()).Render(resultContent)
	case tools.GrepToolName:
		return baseStyle.Width(width).Foreground(t.TextMuted()).Render(resultContent)
	case tools.LSToolName:
		return baseStyle.Width(width).Foreground(t.TextMuted()).Render(resultContent)
	case tools.SourcegraphToolName:
		return baseStyle.Width(width).Foreground(t.TextMuted()).Render(resultContent)
	case tools.ViewToolName:
		metadata := tools.ViewResponseMetadata{}
		json.Unmarshal([]byte(response.Metadata), &metadata)
		ext := filepath.Ext(metadata.FilePath)
		if ext == "" {
			ext = ""
		} else {
			ext = strings.ToLower(ext[1:])
		}
		resultContent = fmt.Sprintf("```%s\n%s\n```", ext, truncateHeight(metadata.Content, maxResultHeight))
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, true, width),
			t.Background(),
		)
	case tools.WriteToolName:
		params := tools.WriteParams{}
		json.Unmarshal([]byte(toolCall.Input), &params)
		metadata := tools.WriteResponseMetadata{}
		json.Unmarshal([]byte(response.Metadata), &metadata)
		ext := filepath.Ext(params.FilePath)
		if ext == "" {
			ext = ""
		} else {
			ext = strings.ToLower(ext[1:])
		}
		resultContent = fmt.Sprintf("```%s\n%s\n```", ext, truncateHeight(params.Content, maxResultHeight))
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, true, width),
			t.Background(),
		)
	default:
		resultContent = fmt.Sprintf("```text\n%s\n```", resultContent)
		return styles.ForceReplaceBackgroundWithLipgloss(
			toMarkdown(resultContent, true, width),
			t.Background(),
		)
	}
}

func renderToolMessage(
	toolCall message.ToolCall,
	allMessages []message.Message,
	messagesService message.Service,
	focusedUIMessageId string,
	nested bool,
	width int,
	position int,
) uiMessage {
	cfg := config.Get()
	originalWidth := width
	
	if nested {
		width = width - 3
	}

	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()
	
	// Calculate tool message width and padding for messaging layout
	var toolWidth int
	var leftPadding int
	
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging && !nested {
		// Tool messages follow assistant message layout
		toolWidth = int(float64(originalWidth) * cfg.TUI.MessageLayoutConfig.AssistantMessageWidth)
		leftPadding = cfg.TUI.MessageLayoutConfig.AssistantLeftMargin
		width = toolWidth
	} else {
		toolWidth = width - 1
		leftPadding = 0
	}

	style := baseStyle.
		Width(toolWidth).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		PaddingLeft(1).
		BorderForeground(t.TextMuted())

	response := findToolResponse(toolCall.ID, allMessages)
	toolNameText := baseStyle.Foreground(t.TextMuted()).
		Render(fmt.Sprintf("%s: ", toolName(toolCall.Name)))

	if !toolCall.Finished {
		// Get a brief description of what the tool is doing
		toolAction := getToolAction(toolCall.Name)

		progressText := baseStyle.
			Width(toolWidth - 2 - lipgloss.Width(toolNameText)).
			Foreground(t.TextMuted()).
			Render(fmt.Sprintf("%s", toolAction))

		content := style.Render(lipgloss.JoinHorizontal(lipgloss.Left, toolNameText, progressText))
		
		// Apply messaging layout wrapper if needed
		if cfg.TUI.MessageLayout == config.MessageLayoutMessaging && leftPadding > 0 && !nested {
			paddingStyle := lipgloss.NewStyle().
				Width(leftPadding).
				Height(lipgloss.Height(content)).
				Background(t.Background())
			
			containerStyle := lipgloss.NewStyle().
				Width(originalWidth).
				Background(t.Background())
			
			content = containerStyle.Render(
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					paddingStyle.Render(""),
					content,
				),
			)
		}
		
		toolMsg := uiMessage{
			messageType: toolMessageType,
			position:    position,
			height:      lipgloss.Height(content),
			content:     content,
		}
		return toolMsg
	}

	params := renderToolParams(toolWidth-2-lipgloss.Width(toolNameText), toolCall)
	responseContent := ""
	if response != nil {
		responseContent = renderToolResponse(toolCall, *response, toolWidth-2)
		responseContent = strings.TrimSuffix(responseContent, "\n")
	} else {
		responseContent = baseStyle.
			Italic(true).
			Width(toolWidth - 2).
			Foreground(t.TextMuted()).
			Render("Waiting for response...")
	}

	parts := []string{}
	if !nested {
		formattedParams := baseStyle.
			Width(toolWidth - 2 - lipgloss.Width(toolNameText)).
			Foreground(t.TextMuted()).
			Render(params)

		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Left, toolNameText, formattedParams))
	} else {
		prefix := baseStyle.
			Foreground(t.TextMuted()).
			Render(" └ ")
		formattedParams := baseStyle.
			Width(toolWidth - 2 - lipgloss.Width(toolNameText)).
			Foreground(t.TextMuted()).
			Render(params)
		parts = append(parts, lipgloss.JoinHorizontal(lipgloss.Left, prefix, toolNameText, formattedParams))
	}

	if toolCall.Name == agent.AgentToolName {
		taskMessages, _ := messagesService.List(context.Background(), toolCall.ID)
		toolCalls := []message.ToolCall{}
		for _, v := range taskMessages {
			toolCalls = append(toolCalls, v.ToolCalls()...)
		}
		for _, call := range toolCalls {
			rendered := renderToolMessage(call, []message.Message{}, messagesService, focusedUIMessageId, true, originalWidth, 0)
			parts = append(parts, rendered.content)
		}
	}
	if responseContent != "" && !nested {
		parts = append(parts, responseContent)
	}

	content := style.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			parts...,
		),
	)
	if nested {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			parts...,
		)
	}
	
	// Apply messaging layout wrapper if needed
	if cfg.TUI.MessageLayout == config.MessageLayoutMessaging && leftPadding > 0 && !nested {
		paddingStyle := lipgloss.NewStyle().
			Width(leftPadding).
			Height(lipgloss.Height(content)).
			Background(t.Background())
		
		containerStyle := lipgloss.NewStyle().
			Width(originalWidth).
			Background(t.Background())
		
		content = containerStyle.Render(
			lipgloss.JoinHorizontal(
				lipgloss.Top,
				paddingStyle.Render(""),
				content,
			),
		)
	}
	
	toolMsg := uiMessage{
		messageType: toolMessageType,
		position:    position,
		height:      lipgloss.Height(content),
		content:     content,
	}
	return toolMsg
}

// Helper function to format the time difference between two Unix timestamps
func formatTimestampDiff(start, end int64) string {
	diffSeconds := float64(end-start) / 1000.0 // Convert to seconds
	if diffSeconds < 1 {
		return fmt.Sprintf("%dms", int(diffSeconds*1000))
	}
	if diffSeconds < 60 {
		return fmt.Sprintf("%.1fs", diffSeconds)
	}
	return fmt.Sprintf("%.1fm", diffSeconds/60)
}
