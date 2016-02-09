#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO add --date flag to add and edit commands
# TODO add --context flag to list command
# TODO implement postpone command
# TODO do not allow --recur when no due date
# TODO move up 3 subcommand (order column)
# TODO daemon subcommand
# TODO notifications
# TODO projects
# TODO search
# TODO filters (https://taskwarrior.org/docs/syntax.html) (https://taskwarrior.org/docs/filter.html)
# TODO auto-complete commands
# TODO note command with $EDITOR

import os

__version__ = "0.0.0"

DB_PATH = os.path.expanduser('~/.yap.sqlite')
CONTEXT_PATH = os.path.expanduser('~/.yap.context')
LIST_DONE_MAX = 20
