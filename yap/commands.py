import os
import sys
import json
import errno
import argparse
import subprocess
from datetime import datetime, date, time, timedelta
from functools import wraps

import isodate
from sqlalchemy import case
from sqlalchemy.exc import SQLAlchemyError
from tabulate import tabulate

import yap
import yap.db
import yap.json_util
import yap.exceptions
from yap.models import Task, Session


def cmd_version(_):
    print yap.__version__


def cmd_next(args):
    args.done = None
    args.waiting = None
    cmd_list(args, limit=args.n)


def cmd_list(args, limit=None):
    session = Session()
    query = session.query(Task)

    headers = ['ID']
    attrs = ['id']

    context = get_context()
    if context:
        query = query.filter(Task.context == context)

    if args.done:
        headers.append('Done at')
        attrs.append('str_done_at')
        query = query\
            .filter(Task.done == True)\
            .order_by(Task.done_at.desc())\
            .limit(yap.LIST_DONE_MAX)
    elif args.waiting:
        headers.append('Wait date')
        attrs.append('str_wait_date')
        query = query\
            .filter(Task.waiting == True)\
            .order_by(Task.wait_date)
    else:
        query = query\
            .filter(Task.done != True, Task.waiting != True)\
            .order_by(  # Show tasks with order date first
                case([(Task.due_date == None, 0)], 1),
                Task.due_date)

    headers.extend(['Due date', 'Title'])
    attrs.extend(['str_due_date', 'title'])

    if limit:
        query = query.limit(limit)

    tasks = query.all()
    table = [[getattr(task, attr) for attr in attrs] for task in tasks]
    session.close()
    print tabulate(table, headers=headers)


def cmd_show(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    for k, v in sorted(vars(task).items()):
        if not k.startswith('_'):
            print "%s: %s" % (k, v)
    session.close()


def cmd_add(args):
    session = Session()
    task = Task()
    task.id = yap.db.get_smallest_empty_id(session, Task)
    task.title = ' '.join(args.title)
    task.due_date = args.due
    task.wait_date = args.wait
    task.context = args.context or get_context()
    task.recur = args.recur
    session.add(task)
    session.commit()
    print "id: %d" % task.id


delete = object()


def cmd_edit(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)

    if args.title:
        task.title = None if args.title == delete else ' '.join(args.title)
    if args.due:
        task.due_date = None if args.due == delete else args.due
    if args.wait:
        task.wait_date = None if args.wait == delete else args.wait
    if args.context:
        task.context = None if args.context == delete else args.context
    if args.recur:
        task.recur = None if args.recur == delete else args.recur

    session.commit()


def cmd_append(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    task.title = "%s %s" % (task.title, ' '.join(args.title))
    session.commit()


def cmd_prepend(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    task.title = "%s %s" % (' '.join(args.title), task.title)
    session.commit()


def cmd_done(args):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(args.id)).all()
    for task in tasks:
        if task.recurring:
            now = datetime.now()
            if task.due_date.time() == time.max:
                now = datetime.combine(now.date(), time.max)
            new_due_date = now + task.recur
            delta = new_due_date - task.due_date
            task.due_date = new_due_date
            if task.wait_date:
                task.wait_date += delta
        else:
            # Make done task's id negative so new tasks can reuse positive ids.
            task.id = yap.db.get_next_negative_id(session, Task)
            task.done_at = datetime.now()
    session.commit()


def cmd_undone(args):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(args.id)).all()
    for task in tasks:
        task.id = yap.db.get_smallest_empty_id(session, Task)
        task.done_at = None
    session.commit()


def cmd_delete(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .delete(synchronize_session=False)
    session.commit()


def cmd_export(args):
    session = Session()
    d = {'task': [task.to_dict() for task in session.query(Task).all()]}
    session.close()
    out = json.dumps(d, default=yap.json_util.datetime_encoder, indent=2)
    args.outfile.write(out)
    args.outfile.write('\n')
    args.outfile.close()


def cmd_import(args):
    has_error = False
    d = json.load(args.infile, object_hook=yap.json_util.datetime_decoder)
    for task in d['task']:
        session = Session()
        session.add(Task.from_dict(task))
        try:
            session.commit()
        except SQLAlchemyError as e:
            print e
            has_error = True
        finally:
            session.close()
    if has_error:
        raise yap.exceptions.TaskImportError


def cmd_context(args):
    if args.name:
        with open(yap.CONTEXT_PATH, 'w') as f:
            f.write(args.name)
    elif args.clear:
        os.unlink(yap.CONTEXT_PATH)
    else:
        print get_context()


def get_context():
    try:
        with open(yap.CONTEXT_PATH, 'r') as f:
            return f.read()
    except IOError as e:
        if e.errno != errno.ENOENT:
            raise


def cmd_daemon(args):  # TODO
    pass


def delete_with_empty_string(f):
    @wraps(f)
    def inner(s):
        if s == '':
            return delete
        return f(s)
    return inner


def date_time_or_datetime(s, default_day, default_time):
    d = parse_day_month_name(s)
    if d:
        return datetime.combine(d, default_time)
    if s == 'now':
        return datetime.now()
    if s == 'today':
        return datetime.combine(date.today(), default_time)
    if s == 'tomorrow':
        return datetime.combine(date.today() + timedelta(days=1), default_time)
    try:
        if s.startswith('P'):
            return datetime.now() + isodate.parse_duration(s)
        if s.startswith('-P'):
            return datetime.now() - isodate.parse_duration(s)
        if 'T' in s:
            return isodate.parse_datetime(s)
        try:
            return datetime.combine(default_day, isodate.parse_time(s))
        except ValueError:
            return datetime.combine(isodate.parse_date(s), default_time)
    except ValueError:
        msg = "%r is not an ISO 8601 date, time or datetime" % s
        raise argparse.ArgumentTypeError(msg)


def parse_day_month_name(s):
    s = s.lower()
    today = date.today()

    days = ['monday', 'tuesday', 'wednesday', 'thursday',
            'friday', 'saturday', 'sunday']
    try:
        i = days.index(s)
    except ValueError:
        pass
    else:
        delta_days = (i - today.weekday()) % 7
        return date.today() + timedelta(days=delta_days)

    months = ['january', 'february', 'march', 'april', 'may', 'june',
              'july', 'august', 'september', 'october', 'november', 'december']
    try:
        i = months.index(s)
    except ValueError:
        pass
    else:
        delta_months = (i+1 - today.month) % 12
        return date.today() + isodate.Duration(months=delta_months)


@delete_with_empty_string
def _due_date(s):
    return date_time_or_datetime(s, date.today(), time.max)


@delete_with_empty_string
def _wait_date(s):
    return date_time_or_datetime(s, date.today(), time.min)


@delete_with_empty_string
def duration(s):
    return isodate.parse_duration(s)


def parse_args():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers()

    parser_version = subparsers.add_parser('version')
    parser_version.set_defaults(func=cmd_version)

    parser_add = subparsers.add_parser('add')
    parser_add.set_defaults(func=cmd_add)
    parser_add.add_argument('title', nargs='+')
    parser_add.add_argument('-d', '--due', type=_due_date,
                            help="due date")
    parser_add.add_argument('-w', '--wait', type=_wait_date,
                            help="do not show before wait date")
    parser_add.add_argument('-r', '--recur', type=duration)
    parser_add.add_argument('-c', '--context')

    parser_list = subparsers.add_parser('list')
    parser_list.set_defaults(func=cmd_list)
    parser_list_group = parser_list.add_mutually_exclusive_group()
    parser_list_group.add_argument('-d', '--done', action='store_true',
                                   help="show done tasks")
    parser_list_group.add_argument('-w', '--waiting', action='store_true',
                                   help="show waiting tasks")

    parser_next = subparsers.add_parser('next')
    parser_next.set_defaults(func=cmd_next)
    parser_next.add_argument('n', type=int, default=1, nargs='?')

    parser_edit = subparsers.add_parser('edit')
    parser_edit.set_defaults(func=cmd_edit)
    parser_edit.add_argument('id', type=int)
    parser_edit.add_argument('-t', '--title', nargs='+')
    parser_edit.add_argument('-d', '--due', type=_due_date)
    parser_edit.add_argument('-w', '--wait', type=_wait_date)
    parser_edit.add_argument('-r', '--recur', type=duration)
    parser_edit.add_argument('-c', '--context')

    parser_append = subparsers.add_parser('append')
    parser_append.set_defaults(func=cmd_append)
    parser_append.add_argument('id', type=int)
    parser_append.add_argument('title', nargs='+')

    parser_prepend = subparsers.add_parser('prepend')
    parser_prepend.set_defaults(func=cmd_prepend)
    parser_prepend.add_argument('id', type=int)
    parser_prepend.add_argument('title', nargs='+')

    parser_show = subparsers.add_parser('show')
    parser_show.set_defaults(func=cmd_show)
    parser_show.add_argument('id', type=int)

    parser_done = subparsers.add_parser('done')
    parser_done.set_defaults(func=cmd_done)
    parser_done.add_argument('id', type=int, nargs='+')

    parser_undone = subparsers.add_parser('undone')
    parser_undone.set_defaults(func=cmd_undone)
    parser_undone.add_argument('id', type=int, nargs='+')

    parser_delete = subparsers.add_parser('delete')
    parser_delete.set_defaults(func=cmd_delete)
    parser_delete.add_argument('id', type=int, nargs='+')

    parser_export = subparsers.add_parser('export')
    parser_export.set_defaults(func=cmd_export)
    parser_export.add_argument('outfile', nargs='?',
                               type=argparse.FileType('w'), default=sys.stdout)

    parser_import = subparsers.add_parser('import')
    parser_import.set_defaults(func=cmd_import)
    parser_import.add_argument('infile', nargs='?',
                               type=argparse.FileType('r'), default=sys.stdin)

    parser_context = subparsers.add_parser('context')
    parser_context.set_defaults(func=cmd_context)
    parser_context.add_argument('name', nargs='?')
    parser_context.add_argument('-c', '--clear', action='store_true')

    parser_daemon = subparsers.add_parser('daemon')
    parser_daemon.set_defaults(func=cmd_daemon)

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


def run_script(script):  # TODO
    return subprocess.check_output(
            ['osascript', '-'], stdin=subprocess.PIPE, stderr=subprocess.PIPE)


def display_notification(message, title):  # TODO
    pass
