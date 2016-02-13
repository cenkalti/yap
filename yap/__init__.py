#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO setup travis
# TODO coverage
# TODO skip recurred tasks (http://orgmode.org/manual/Repeated-tasks.html)
# TODO add --before --after to add command (mutually exclusive)
# TODO add --before --after to edit command (mutually exclusive)
# TODO add parent to tasks
# TODO daemon subcommand
# TODO notifications
# TODO note command with $EDITOR
# TODO implement undo command
# TODO auto-complete commands
# TODO refactor list command
# TODO do not allow --recur when no due date
# TODO do not allow --on when --due or --wait set
# TODO projects
# TODO search
# TODO filters (https://taskwarrior.org/docs/syntax.html) (https://taskwarrior.org/docs/filter.html)

__version__ = "0.0.0"

LIST_DONE_MAX = 20
