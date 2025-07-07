## Branch-Based Development Workflow

When working on issues or new features, follow this workflow:

1. **Research the issue thoroughly** to understand its scope and impact
2. **Create a descriptive branch** from main with an appropriate prefix:
   - `fix/` for bug fixes (e.g., `fix/completion-dialog-scrolling`)
   - `feat/` for new features (e.g., `feat/oauth-integration`)
   - `docs/` for documentation updates (e.g., `docs/testing-guidelines`)
   - `refactor/` for code refactoring (e.g., `refactor/session-management`)
3. **Work in the feature branch** making atomic commits as you progress
4. **Test thoroughly** before considering the work complete
5. **Push the branch and create a PR** or merge when ready

Example workflow:
```bash
# Start with a bug fix
git checkout main
git pull origin main
git checkout -b fix/completion-dialog-continuous-scroll

# Make changes and commit
git add -A
git commit -m "fix: prevent continuous scrolling in completion dialog"

# Push branch
git push origin fix/completion-dialog-continuous-scroll
```

This approach:
- Keeps main branch stable
- Allows parallel work on multiple issues
- Makes it easy to review changes
- Enables clean reversion if needed

## Git Commit Guidelines

When making a git commit, do not include any reference or advertising from AI assistants (Claude, Claude Code, Gemini, etc.). The commit message should be clear and concise, focusing on the changes made in the codebase.

For multi-line git commit use here doc for the multi-line string.
