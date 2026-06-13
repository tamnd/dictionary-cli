---
title: "Quick start"
description: "Run your first dict command."
weight: 30
---

Once `dict` is on your `PATH`:

```bash
dict --help       # see the command tree
dict version      # build info
```

This is a fresh scaffold, so the command tree is just `version` for now. Add
your first real command in `cli/`, build on the `dictionary` library package,
and document it here.

A good first command usually fetches one thing and prints it as JSON, so the
output pipes straight into `jq` and the rest of your tools.
