import errno
import json
import os
import subprocess
from datetime import datetime, date, time, timedelta

from tabulate import tabulate
from sqlalchemy import case
from sqlalchemy.exc import SQLAlchemyError

import yap
import yap.db
import yap.exceptions
import yap.json_util
from yap.models import Task, Session


def version(_):
    print yap.__version__


def next_(args):
    args.done = None
    args.waiting = None
    args.archived = None
    args.context = None
    list_(args, only_next=True)


def list_(args, only_next=False):
    session = Session()
    query = session.query(Task)

    headers = ['ID']
    attrs = ['id']

    ctx = args.context or _get_context()
    if ctx:
        query = query.filter(Task.context == ctx)

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
        query = query.filter(Task.archived == True)\
            .order_by(Task.created_at.desc())
    else:
        query = query\
            .filter(Task.done != True, Task.waiting != True)\
            .filter(Task.archived != True)\
            .order_by(  # Show tasks with order date first
                case([(Task.due_date == None, 0)], 1),
                Task.due_date,
                case([(Task.order == None, 0)], 1),
                Task.order.desc())

    headers.append('Due date')
    attrs.append('str_due_date')
    if not ctx:
        headers.append('Context')
        attrs.append('context')
    headers.append('Title')
    attrs.append('title')

    tasks = query.all()
    if only_next:
        is_next = lambda t: t.remaining or timedelta.max < timedelta(days=1)
        overdue_all = [t for t in tasks if is_next(t)]
        not_overdue = [t for t in tasks if not is_next(t)]
        tasks = overdue_all + not_overdue[:1]

    table = [[getattr(task, attr) for attr in attrs] for task in tasks]
    session.close()
    print tabulate(table, headers=headers, tablefmt='plain')


def show(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    for k, v in sorted(vars(task).items()):
        if not k.startswith('_'):
            print "%s: %s" % (k, v)
    session.close()


def add(args):
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
    task.context = args.context or _get_context()
    task.recur = args.recur
    if task.recur:
        task.shift = args.shift
    task.order = Task.find_next_order(session)
    session.add(task)
    session.commit()
    print "id: %d" % task.id


def edit(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)

    if args.title:
        task.title = None if args.title == delete_option else ' '.join(args.title)
    if args.on:
        if args.on == delete_option:
            task.due_date = None
            task.wait_date = None
        else:
            task.due_date = datetime.combine(args.on, time.max)
            task.wait_date = datetime.combine(args.on, time.min)
    if args.due:
        task.due_date = None if args.due == delete_option else args.due
    if args.wait:
        task.wait_date = None if args.wait == delete_option else args.wait
    if args.context:
        task.context = None if args.context == delete_option else args.context
    if args.recur:
        task.recur = None if args.recur == delete_option else args.recur
    if args.shift is not None:
        task.shift = args.shift

    session.commit()
delete_option = object()


def append(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    task.title = "%s %s" % (task.title, ' '.join(args.title))
    session.commit()


def prepend(args):
    session = Session()
    task = session.query(Task).get(args.id)
    if not task:
        raise yap.exceptions.TaskNotFoundError(args.id)
    task.title = "%s %s" % (' '.join(args.title), task.title)
    session.commit()


def done(args):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(args.id)).all()
    for task in tasks:
        if task.recurring:
            new_task = Task.from_dict(task.to_dict())
            new_task.id = yap.db.get_smallest_empty_id(session, Task)
            new_task.created_at = datetime.now()
            if task.shift:
                if new_task.due_date.time() == time.max:
                    base = datetime.combine(date.today(), time.max)
                else:
                    base = datetime.now()
            else:
                base = task.due_date
            new_due_date = base + new_task.recur
            delta = new_due_date - new_task.due_date
            new_task.due_date = new_due_date
            if new_task.wait_date:
                new_task.wait_date += delta
            session.add(new_task)
        # Make done task's id negative so new tasks can reuse positive ids.
        task.id = yap.db.get_next_negative_id(session, Task)
        task.done_at = datetime.now()
    session.commit()


def undone(args):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(args.id)).all()
    for task in tasks:
        task.id = yap.db.get_smallest_empty_id(session, Task)
        task.done_at = None
    session.commit()


def delete(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .delete(synchronize_session=False)
    session.commit()


def archive(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.id: yap.db.get_next_negative_id(session, Task)},
                synchronize_session=False)
    session.commit()


def wait(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.wait_date: args.wait_date}, synchronize_session=False)
    session.commit()


def postpone(args):
    session = Session()
    session.query(Task).filter(Task.id.in_(args.id))\
        .update({Task.due_date: args.due_date}, synchronize_session=False)
    session.commit()


def export(args):
    session = Session()
    d = {'task': [task.to_dict() for task in session.query(Task).all()]}
    session.close()
    out = json.dumps(d, default=yap.json_util.datetime_encoder, indent=2)
    args.outfile.write(out)
    args.outfile.write('\n')
    args.outfile.close()


def import_(args):
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


def context(args):
    if args.name:
        with open(yap.CONTEXT_PATH, 'w') as f:
            f.write(args.name)
    elif args.clear:
        os.unlink(yap.CONTEXT_PATH)
    else:
        print _get_context()


def _get_context():
    try:
        with open(yap.CONTEXT_PATH, 'r') as f:
            return f.read()
    except IOError as e:
        if e.errno != errno.ENOENT:
            raise


def daemon(args):  # TODO
    pass


def _run_script(script):  # TODO
    return subprocess.check_output(
            ['osascript', '-'], stdin=subprocess.PIPE, stderr=subprocess.PIPE)


def _display_notification(message, title):  # TODO
    pass
