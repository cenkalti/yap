#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO export subcommand
# TODO import subcommand
# TODO move up 3 subcommand (order column)
# TODO recurring tasks (https://taskwarrior.org/docs/recurrence.html) (https://taskwarrior.org/docs/durations.html)
# TODO daemon subcommand
# TODO notifications
# TODO context (https://taskwarrior.org/docs/context.html)
# TODO projects
# TODO human dates (https://taskwarrior.org/docs/dates.html) (https://taskwarrior.org/docs/named_dates.html)
# TODO search
# TODO filters (https://taskwarrior.org/docs/syntax.html) (https://taskwarrior.org/docs/filter.html)
# TODO auto-completing

import os

__version__ = "0.0.0"

DATE_FORMAT = '%Y-%m-%d'
DATETIME_FORMAT = DATE_FORMAT + ' %H:%M'
DB_PATH = os.path.expanduser('~/.yap.sqlite')
LIST_DONE_MAX = 20
