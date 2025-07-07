# Opencode Command Registry & Slash Command Roadmap

## 1. Command Registry & Discovery

- [x] **1.1. Define Command Registry Structure**

  - Design a `Command` struct (fields: name, description, handler, children, metadata).
  - Use a hierarchical data structure (e.g., `map[string]*Command` or tree).
  - Define interfaces for built-in, user, and project commands.
  - Place registry code in a dedicated package (`commands/registry.go`).

- [x] **1.2. Implement Command Scanning**

  - Use `filepath.WalkDir` to scan user/project command directories for `.md` files.
  - Parse YAML/frontmatter metadata using `gopkg.in/yaml.v3` or similar.
  - Store discovered commands in the registry, marking their source type.

- [ ] **1.3. Register Built-in Commands**

  - Register built-in commands at startup via a unified API.
  - Link commands to their Go handler functions.

- [ ] **1.4. Support Dynamic Sub-Commands**
  - Allow registry nodes to have a `DynamicChildren` callback for lazy, runtime sub-command enumeration (e.g., `/session list`).
  - Ensure uniform handling of static and dynamic children in the registry.

---

## 2. Slash Command Parsing & Autocomplete

- [ ] **2.1. Extend Command Parser**

  - Tokenize input (`/command [subcommand] [args]`) with `strings.Fields` or a custom lexer (supporting quoted args).
  - Traverse registry hierarchy matching tokens; return parse errors or suggestions as needed.

- [ ] **2.2. Autocomplete Structure**

  - Implement a trie (prefix tree) for efficient prefix-based autocomplete.
  - For substring/fuzzy matches, maintain a flat index or use a library like `github.com/sahilm/fuzzy`.
  - Trie nodes point to corresponding command registry entries.

- [ ] **2.3. Integrate Filtering**
  - Combine prefix (trie) and substring/fuzzy filtering in the popup for user-friendly matching.

---

## 3. Popup/Picker UI Enhancements

- [ ] **3.1. Generalize Picker Component**

  - Picker accepts any `ItemProvider` interface to supply items (static or runtime).
  - Support both static registry lists and dynamic sources (e.g., sessions).

- [ ] **3.2. Keyboard Navigation**

  - Support arrow keys and vim keys (`h`, `j`, `k`, `l`) for navigation.
  - Implement a simple state machine to manage selection and navigation.

- [ ] **3.3. Show Breadcrumbs/Context**

  - Maintain and display breadcrumbs showing command/sub-command path (e.g., `/session > list > ...`).

- [ ] **3.4. UI Feedback for Invalid Input**
  - Indicate "No match" or highlight invalid/unrecognized input in the picker.

---

## 4. Command Execution Integration

- [ ] **4.1. Unified Command Dispatcher**

  - Dispatcher routes parsed commands to the correct handler (built-in or user/project).
  - Standardize handler signature: `func(ctx Context, args []string) error`.

- [ ] **4.2. Argument Prompting**

  - If a command requires missing arguments, prompt user via UI.
  - Use command metadata to drive prompts.

- [ ] **4.3. Dynamic Selection Integration**
  - For commands needing runtime selection (e.g., choose a session), invoke the picker and supply the result to the handler.

---

## 5. Dynamic Context Updates

- [ ] **5.1. Monitor Directory Changes**
  - Use `fsnotify` or polling to watch for updates in command directories.
  - On changes, re-scan and update the command registry (with debouncing/throttling).

---

## 6. Testing

- [ ] **6.1. Unit Testing**

  - Write table-driven tests for the parser, covering edge cases and errors.
  - Test registry population and lookup, static and dynamic children.
  - Simulate picker UI navigation, including vim keys.

- [ ] **6.2. Integration Testing**
  - Create end-to-end tests simulating user input for commands and UI navigation.
  - Mock dynamic sources for isolation.

---

## 7. Documentation

- [ ] **7.1. Update User & Developer Docs**
  - Document slash command usage, sub-commands, argument prompts, and navigation keys.
  - Add screenshots/gifs of popup UI.
  - Explain how to add project/user commands.

---

## 8. General Go Engineering Practices

- Use interfaces to decouple registry, discovery, and execution.
- Keep command metadata in structs, avoid hardcoding UI logic in handlers.
- Organize code into packages: `commands/`, `ui/`, `registry/`, etc.
- Use `context.Context` in handlers for cancellation/timeouts.
- Add logging and error reporting for registry loading and command execution.

---

**Legend:**

- [x] = Completed
- [ ] = To Do

_Last updated: 2025-07-03 01:15 UTC_
