# Project Rules

See @../AGENTS.md for project guidelines.

See @../redhat-compliance-and-responsible-ai.md for Red Hat compliance and responsible AI rules.

## Git Commit Sign-off

**MANDATORY**: Every `git commit` command MUST include the `-s` (`--signoff`) flag to add a DCO sign-off trailer:

```
Signed-off-by: Name <email>
```

- ✅ `git commit -s -m "..."`
- ✅ `git commit -s --amend`
- ❌ `git commit -m "..."` — missing sign-off, DO NOT do this

This applies to every commit in this repository, including amends.
