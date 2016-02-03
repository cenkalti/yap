import argparse
import subprocess
import sys
from datetime import datetime

from sqlalchemy import case
from tabulate import tabulate

import yap
import yap.db
from yap.models import Todo, DoneTodo, Session


def cmd_version(_):
    print yap.__version__


def cmd_list(args):
    session = Session()
    query = session.query(Todo)
    if args.done:
        query = query.filter(Todo.done == True).union(session.query(DoneTodo))
        query = query.order_by(Todo.done_at.desc())
        query = query.limit(yap.LIST_DONE_MAX)
    elif args.waiting:
        query = query.filter(Todo.waiting == True)
        query = query.order_by(Todo.wait_date)
    else:
        query = query.filter(Todo.done != True, Todo.waiting != True)
        # Show items with due date first
        query = query.order_by(case([(Todo.due_date == None, 0)], 1),
                               Todo.due_date)

    items = query.all()

    if args.done:
        headers = ('ID', 'Done at', 'Due date', 'Title')
        attrs = ('id', 'str_done_at', 'str_due_date', 'title')
    elif args.waiting:
        headers = ('ID', 'Wait date', 'Due date', 'Title')
        attrs = ('id', 'str_wait_date', 'str_due_date', 'title')
    else:
        headers = ('ID', 'Due date', 'Title')
        attrs = ('id', 'str_due_date', 'title')

    table = [[getattr(item, attr) for attr in attrs] for item in items]
    print tabulate(table, headers=headers)
    session.close()


def cmd_show(args):
    session = Session()
    todo = session.query(Todo).get(args.id)
    if not todo:
        raise TodoNotFoundError(args.id)
    for k, v in sorted(vars(todo).items()):
        if not k.startswith('_'):
            print "%s: %s" % (k, v)
    session.close()


def cmd_add(args):
    session = Session()
    todo = Todo()
    todo.id = yap.db.get_smallest_empty_id(session, Todo)
    todo.title = ' '.join(args.title)
    todo.due_date = args.due
    todo.wait_date = args.wait
    session.add(todo)
    session.commit()
    print "id: %d" % todo.id


def cmd_edit(args):
    session = Session()
    todo = session.query(Todo).get(args.id)
    if not todo:
        raise TodoNotFoundError(args.id)
    if args.title:
        todo.title = ' '.join(args.title)
    if args.due:
        todo.due_date = args.due
    if args.wait:
        todo.wait_date = args.wait
    session.commit()


def cmd_done(args):
    session = Session()
    session.query(Todo).filter(Todo.id.in_(args.id)).update({
        Todo.done_at: datetime.utcnow(),
    }, synchronize_session=False)
    session.commit()


def cmd_undone(args):
    session = Session()
    session.query(Todo).filter(Todo.id.in_(args.id)).update({
        Todo.done_at: None,
    }, synchronize_session=False)
    session.commit()


def cmd_delete(args):
    session = Session()
    session.query(Todo).filter(Todo.id.in_(args.id))\
        .delete(synchronize_session=False)
    session.commit()


def cmd_daemon(args):  # TODO
    pass


def strdate(s):
    return datetime.strptime(s, yap.DATE_FORMAT)


def parse_args():
    parser = argparse.ArgumentParser()
    subparsers = parser.add_subparsers()

    parser_version = subparsers.add_parser('version')
    parser_version.set_defaults(func=cmd_version)

    parser_add = subparsers.add_parser('add')
    parser_add.set_defaults(func=cmd_add)
    parser_add.add_argument('title', nargs='+')
    parser_add.add_argument('-d', '--due', type=strdate,
                            help="due date")
    parser_add.add_argument('-w', '--wait', type=strdate,
                            help="do not show before wait date")

    parser_list = subparsers.add_parser('list')
    parser_list.set_defaults(func=cmd_list)
    parser_list_group = parser_list.add_mutually_exclusive_group()
    parser_list_group.add_argument('-d', '--done', action='store_true',
                                   help="show done items")
    parser_list_group.add_argument('-w', '--waiting', action='store_true',
                                   help="show waiting items")

    parser_edit = subparsers.add_parser('edit')
    parser_edit.set_defaults(func=cmd_edit)
    parser_edit.add_argument('id', type=int)
    parser_edit.add_argument('-t', '--title', nargs='+')
    parser_edit.add_argument('-d', '--due', type=strdate)
    parser_edit.add_argument('-w', '--wait', type=strdate)

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

    parser_daemon = subparsers.add_parser('daemon')
    parser_daemon.set_defaults(func=cmd_daemon)

    # If invoked with no subcommand, run list subcommand
    if len(sys.argv) == 1:
        sys.argv.append('list')

    args = parser.parse_args()
    args.func(args)


class TodoNotFoundError(ValueError):
    def __init__(self, tid):
        super(TodoNotFoundError, self).__init__("todo id not found: %s" % tid)


def run_script(script):  # TODO
    return subprocess.check_output(
            ['osascript', '-'], stdin=subprocess.PIPE, stderr=subprocess.PIPE)


def display_notification(message, title):  # TODO
    pass
