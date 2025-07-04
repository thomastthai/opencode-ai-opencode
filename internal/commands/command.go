package commands

import (
	"context"
	"fmt"
)

// BaseCommand provides a default implementation of the Command interface.
type BaseCommand struct {
	id          string
	name        string
	description string
	category    string
	commandType CommandType
	example     string
	aliases     []string
	arguments   []ArgumentDefinition
	subCommands []Command
	parent      Command
	metadata    map[string]interface{}
	handler     CommandHandler
}

// NewBaseCommand creates a new BaseCommand.
func NewBaseCommand(id, name, description string) *BaseCommand {
	return &BaseCommand{
		id:          id,
		name:        name,
		description: description,
		subCommands: make([]Command, 0),
		metadata:    make(map[string]interface{}),
	}
}

func (b *BaseCommand) ID() string                                { return b.id }
func (b *BaseCommand) Name() string                              { return b.name }
func (b *BaseCommand) Description() string                       { return b.description }
func (b *BaseCommand) Category() string                          { return b.category }
func (b *BaseCommand) Type() CommandType                         { return b.commandType }
func (b *BaseCommand) Example() string                           { return b.example }
func (b *BaseCommand) GetAliases() []string                      { return b.aliases }
func (b *BaseCommand) GetArguments() []ArgumentDefinition        { return b.arguments }
func (b *BaseCommand) GetSubCommands() []Command                 { return b.subCommands }
func (b *BaseCommand) GetParent() Command                        { return b.parent }
func (b *BaseCommand) GetMetadata() map[string]interface{}       { return b.metadata }

func (b *BaseCommand) GetPath() string {
	if b.parent != nil {
		return b.parent.GetPath() + " " + b.name
	}
	return b.name
}

func (b *BaseCommand) AddSubCommand(sub Command) {
	if bc, ok := sub.(*BaseCommand); ok {
		bc.parent = b
	}
	b.subCommands = append(b.subCommands, sub)
}

func (b *BaseCommand) Execute(ctx context.Context, args map[string]interface{}) error {
	if b.handler != nil {
		return b.handler(ctx, args)
	}
	return fmt.Errorf("no handler defined for command '%s'", b.id)
}

func (b *BaseCommand) ValidateArgs(args map[string]interface{}) error {
	for _, argDef := range b.arguments {
		if argDef.Required {
			if _, exists := args[argDef.Name]; !exists {
				return fmt.Errorf("missing required argument: %s", argDef.Name)
			}
		}
	}
	return nil
}

// CommandBuilder provides a fluent API for creating commands.
type CommandBuilder struct {
	cmd *BaseCommand
}

// NewCommand creates a new command builder.
func NewCommand(id, name, description string) *CommandBuilder {
	return &CommandBuilder{
		cmd: NewBaseCommand(id, name, description),
	}
}

func (b *CommandBuilder) WithCategory(category string) *CommandBuilder {
	b.cmd.category = category
	return b
}

func (b *CommandBuilder) WithType(cmdType CommandType) *CommandBuilder {
	b.cmd.commandType = cmdType
	return b
}

func (b *CommandBuilder) WithExample(example string) *CommandBuilder {
	b.cmd.example = example
	return b
}

func (b *CommandBuilder) WithAliases(aliases []string) *CommandBuilder {
	b.cmd.aliases = aliases
	return b
}

func (b *CommandBuilder) WithArguments(args []ArgumentDefinition) *CommandBuilder {
	b.cmd.arguments = args
	return b
}

func (b *CommandBuilder) WithSubCommands(subs ...Command) *CommandBuilder {
	for _, sub := range subs {
		b.cmd.AddSubCommand(sub)
	}
	return b
}

func (b *CommandBuilder) WithMetadata(meta map[string]interface{}) *CommandBuilder {
	b.cmd.metadata = meta
	return b
}

func (b *CommandBuilder) WithHandler(handler CommandHandler) *CommandBuilder {
	b.cmd.handler = handler
	return b
}

func (b *CommandBuilder) Build() Command {
	return b.cmd
}
