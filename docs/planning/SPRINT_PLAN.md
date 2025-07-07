# Sprint Plan: 1.3 Register Built-in Commands <!-- ID: 0197ce92-ee78-7a58-8142-ca51b230ac7a -->

## Objective
Implement robust registration of built-in commands at startup via a unified API and ensure each is linked to its Go handler. This foundation enables reliable command discovery and execution.

---

## Tasks

- [x] **1.3.1. Design Built-in Command API**
  - Defined `RegisterBuiltIn(cmd Command)` in `commands/registry.go`.
  - Documented API usage and expectations in code comments.

- [x] **1.3.2. Implement Built-in Command Registration Logic**
  - All built-in commands are registered at application startup in `commands/builtin.go`.
  - `RegisterBuiltInHierarchy` supports nested commands.

- [x] **1.3.3. Link Commands to Go Handler Functions**
  - `BaseCommand` stores and executes the Go handler.
  - Handler signature is `func(ctx context.Context, args map[string]interface{}) error`.
  - `Execute` returns an error for missing handlers.

- [x] **1.3.4. Prevent Duplicate or Invalid Registrations**
  - `Register` returns an error on duplicate command IDs.
  - `RegisterBuiltIn` uses `log.Fatalf` to halt on registration errors.

- [x] **1.3.5. Add/Update Unit Tests**
  - Updated `registry_test.go` for the new `Command` interface.
  - Added tests for registration, duplicates, hierarchies, and handlers.

- [x] **1.3.6. Update Documentation**
  - Added inline code comments and docstrings to all new and modified Go files.

---

**Legend:**  
- [x] = Completed  
- [ ] = To Do  

_Last updated: 2025-07-03 03:51 UTC_