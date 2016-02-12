import sys
import operator
import argparse
from functools import wraps
from datetime import datetime, date, timedelta, time

import isodate

import yap
import yap.commands
import yap.exceptions


def _delete_with_empty_string(f):
    @wraps(f)
    def inner(s):
        if s == '':
            return yap.commands.delete_option
        return f(s)
    return inner


def date_time_or_datetime(s, default_day, default_time):
    d = _parse_day_name(s)
    if d:
        return datetime.combine(d, default_time)
    if s == 'now':
        return datetime.now()
    if s == 'today':
        return datetime.combine(date.today(), default_time)
    if s == 'tomorrow':
        return datetime.combine(date.today() + timedelta(days=1), default_time)
    try:
        if s.startswith('-'):
            op = operator.sub
            s = s[1:]
        else:
            op = operator.add
        if s.startswith('P'):  # ISO 8601 duration
            td = isodate.parse_duration(s)
            try:
                td = td.tdelta
            except AttributeError:
                pass
            if td.seconds == 0:  # resolution is day
                now = datetime.combine(datetime.today(), default_time)
            else:
                now = datetime.now()
            return op(now, isodate.parse_duration(s))
        if 'T' in s:  # ISO 8601 combined date and time
            return isodate.parse_datetime(s)
        if '-' in s:
            return datetime.combine(isodate.parse_date(s), default_time)
        if ':' in s:
            return datetime.combine(default_day, isodate.parse_time(s))
        raise ValueError
    except ValueError:
        msg = "%r is not an ISO 8601 date, time or datetime" % s
        raise argparse.ArgumentTypeError(msg)


def _parse_day_name(s):
    days = ['monday', 'tuesday', 'wednesday', 'thursday',
            'friday', 'saturday', 'sunday']
    try:
        i = days.index(s.lower())
    except ValueError:
        pass
    else:
        delta_days = (i - date.today().weekday()) % 7
        return date.today() + timedelta(days=delta_days)


@_delete_with_empty_string
def due_date(s):
    return date_time_or_datetime(s, date.today(), time.max)


@_delete_with_empty_string
def wait_date(s):
    return date_time_or_datetime(s, date.today(), time.min)


@_delete_with_empty_string
def on_date(s):
    return date_time_or_datetime(s, date.today(), time.min).date()


@_delete_with_empty_string
def duration(s):
    return isodate.parse_duration(s)


def parse_args():
    parser = argparse.ArgumentParser(usage='%(prog)s [-h] [-v] subcommand ...')
    parser.add_argument('-v', '--version', action='version',
                        version='%(prog)s ' + yap.__version__)

    subparsers = parser.add_subparsers(title="subcommands", prog="yap", metavar="")

    parser_add = subparsers.add_parser('add', help="add new task")
    parser_add.set_defaults(func=yap.commands.add)
    parser_add.add_argument('title', nargs='+')
    parser_add.add_argument('-d', '--due', type=due_date,
                            help="due date")
    parser_add.add_argument('-w', '--wait', type=wait_date,
                            help="do not show before wait date")
    parser_add.add_argument('-o', '--on', type=on_date,
                            help="set due date and wait date to same day")
    parser_add.add_argument('-r', '--recur', type=duration)
    parser_add.add_argument('-s', '--shift', action='store_true')
    parser_add.add_argument('-c', '--context')

    parser_list = subparsers.add_parser('list', help="list tasks")
    parser_list.set_defaults(func=yap.commands.list_)
    parser_list.add_argument('-c', '--context', help="only items in context")
    parser_list_group = parser_list.add_mutually_exclusive_group()
    parser_list_group.add_argument('-d', '--done', action='store_true',
                                   help="only done tasks")
    parser_list_group.add_argument('-w', '--waiting', action='store_true',
                                   help="only waiting tasks")
    parser_list_group.add_argument('-a', '--archived', action='store_true',
                                   help="only archived tasks")

    parser_next = subparsers.add_parser('next', help="list next tasks to do")
    parser_next.set_defaults(func=yap.commands.next_)

    parser_edit = subparsers.add_parser('edit', help="edit task")
    parser_edit.set_defaults(func=yap.commands.edit)
    parser_edit.add_argument('id', type=int)
    edit_group = parser_edit.add_mutually_exclusive_group()
    edit_group.add_argument('-t', '--title', nargs='+',
                            help="replace title")
    edit_group.add_argument('-a', '--append', nargs='+',
                            help="append text to title")
    edit_group.add_argument('-p', '--prepend', nargs='+',
                            help="prepend text to title")
    parser_edit.add_argument('-d', '--due', type=due_date)
    parser_edit.add_argument('-w', '--wait', type=wait_date)
    parser_edit.add_argument('-o', '--on', type=on_date)
    parser_edit.add_argument('-r', '--recur', type=duration)
    parser_edit.add_argument('-s', '--shift', type=bool)
    parser_edit.add_argument('-c', '--context')

    parser_show = subparsers.add_parser('show', help="show task detail")
    parser_show.set_defaults(func=yap.commands.show)
    parser_show.add_argument('id', type=int)

    parser_done = subparsers.add_parser('done', help="mark task as done")
    parser_done.set_defaults(func=yap.commands.done)
    parser_done.add_argument('id', type=int, nargs='+')

    parser_undone = subparsers.add_parser('undone', help="mark task as undone")
    parser_undone.set_defaults(func=yap.commands.undone)
    parser_undone.add_argument('id', type=int, nargs='+')

    parser_delete = subparsers.add_parser('delete', help="delete task")
    parser_delete.set_defaults(func=yap.commands.delete)
    parser_delete.add_argument('id', type=int, nargs='+')

    parser_archive = subparsers.add_parser('archive', help="archive task")
    parser_archive.set_defaults(func=yap.commands.archive)
    parser_archive.add_argument('id', type=int, nargs='+')

    parser_wait = subparsers.add_parser('wait', help="hide task until date")
    parser_wait.set_defaults(func=yap.commands.wait)
    parser_wait.add_argument('wait_date', type=wait_date)
    parser_wait.add_argument('id', type=int, nargs='+')

    parser_postpone = subparsers.add_parser('postpone',
                                            help="postpone due date")
    parser_postpone.set_defaults(func=yap.commands.postpone)
    parser_postpone.add_argument('due_date', type=due_date)
    parser_postpone.add_argument('id', type=int, nargs='+')

    parser_context = subparsers.add_parser('context', help="get or set context")
    parser_context.set_defaults(func=yap.commands.context)
    parser_context.add_argument('name', nargs='?')
    parser_context.add_argument('-c', '--clear', action='store_true')

    parser_export = subparsers.add_parser('export',
                                          help="export database as json")
    parser_export.set_defaults(func=yap.commands.export)
    parser_export.add_argument('outfile', nargs='?',
                               type=argparse.FileType('w'), default=sys.stdout)

    parser_import = subparsers.add_parser('import',
                                          help="import tasks from json")
    parser_import.set_defaults(func=yap.commands.import_)
    parser_import.add_argument('infile', nargs='?',
                               type=argparse.FileType('r'), default=sys.stdin)

    parser_daemon = subparsers.add_parser('daemon',
                                          help="run notification daemon")
    parser_daemon.set_defaults(func=yap.commands.daemon)

    # If invoked with no subcommand, run next subcommand
    if len(sys.argv) == 1:
        sys.argv.append('next')

    args = parser.parse_args()
    try:
        args.func(args)
    except yap.exceptions.YapError as e:
        sys.stderr.write(str(e))
        sys.stderr.write('\n')
        sys.exit(1)
