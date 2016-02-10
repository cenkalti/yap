#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO add strict reucurrence
# TODO move up 3 subcommand (order column)
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

import os

__version__ = "0.0.0"

DB_PATH = os.path.expanduser('~/.yap.sqlite')
CONTEXT_PATH = os.path.expanduser('~/.yap.context')
LIST_DONE_MAX = 20
