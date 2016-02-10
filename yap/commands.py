import errno
import json
import os
import subprocess
from datetime import datetime, date, time

from tabulate import tabulate
from sqlalchemy import case
from sqlalchemy.exc import SQLAlchemyError

import yap
import yap.db
import yap.exceptions
import yap.json_util
from yap.models import Task, Session


def cmd_version(_):
    print yap.__version__


def cmd_next(args):
    args.done = None
    args.waiting = None
    args.archived = None
    args.context = None
    cmd_list(args, limit=args.n)


def cmd_list(args, limit=None):
    session = Session()
    query = session.query(Task)

    headers = ['ID']
    attrs = ['id']

    context = args.context or get_context()
    if context:
        query = query.filter(Task.context == context)

    if args.done:
        headers.append('Done at')
        attrs.append('str_done_at')
        query = query\
            .filter(Task.done == True)\
            .filter(Task.archived != True)\
            .order_by(Task.done_at.desc())\
            .limit(yap.LIST_DONE_MAX)
    elif args.waiting:
        headers.append('Wait date')
        attrs.append('str_wait_date')
        query = query\
            .filter(Task.waiting == True)\
            .filter(Task.archived != True)\
            .order_by(Task.wait_date)
    elif args.archived:
        query = query.filter(Task.archived == True)
    else:
        query = query\
            .filter(Task.done != True, Task.waiting != True)\
            .filter(Task.archived != True)\
            .order_by(  # Show tasks with order date first
                case([(Task.due_date == None, 0)], 1),
                Task.due_date,
                Task.created_at)

    headers.append('Due date')
    attrs.append('str_due_date')
    if not context:
        headers.append('Context')
        attrs.append('context')
    headers.append('Title')
    attrs.append('title')

    if limit:
        query = query.limit(limit)

    tasks = query.all()
    table = [[getattr(task, attr) for attr in attrs] for task in tasks]
    session.close()
    print tabulate(table, headers=headers, tablefmt='plain')


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
    if args.on:
        task.due_date = datetime.combine(args.on, time.max)
        task.wait_date = datetime.combine(args.on, time.min)
    else:
        task.due_date = args.due
        task.wait_date = args.wait
    task.context = args.context or get_context()
    task.recur = args.recur
    session.add(task)
    session.commit()
    print "id: %d" % task.id


def cmd_edit(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)

    if args.title:
        task.title = None if args.title == delete else ' '.join(args.title)
    if args.on:
        if args.on == delete:
            task.due_date = None
            task.wait_date = None
        else:
            task.due_date = datetime.combine(args.on, time.max)
            task.wait_date = datetime.combine(args.on, time.min)
    if args.due:
        task.due_date = None if args.due == delete else args.due
    if args.wait:
        task.wait_date = None if args.wait == delete else args.wait
    if args.context:
        task.context = None if args.context == delete else args.context
    if args.recur:
        task.recur = None if args.recur == delete else args.recur

    session.commit()
delete = object()


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
            new_task = Task.from_dict(task.to_dict())
            new_task.id = yap.db.get_smallest_empty_id(session, Task)
            new_task.created_at = datetime.now()
            if new_task.due_date.time() == time.max:
                now = datetime.combine(date.today(), time.max)
            else:
                now = datetime.now()
            new_due_date = now + new_task.recur
            delta = new_due_date - new_task.due_date
            new_task.due_date = new_due_date
            if new_task.wait_date:
                new_task.wait_date += delta
            session.add(new_task)
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


def cmd_archive(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.id: yap.db.get_next_negative_id(session, Task)},
                synchronize_session=False)
    session.commit()


def cmd_wait(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.wait_date: args.wait_date}, synchronize_session=False)
    session.commit()


def cmd_postpone(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.due_date: args.due_date}, synchronize_session=False)
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


def run_script(script):  # TODO
    return subprocess.check_output(
            ['osascript', '-'], stdin=subprocess.PIPE, stderr=subprocess.PIPE)


def display_notification(message, title):  # TODO
    pass
