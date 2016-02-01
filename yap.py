#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO move done tasks to another table for smaller ids, keep done tasks for a week
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
import sys
import argparse
import subprocess
from datetime import datetime, timedelta, time

from tabulate import tabulate
from sqlalchemy import create_engine, MetaData, Table, and_, case
from sqlalchemy import Column, Integer, String, DateTime
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.schema import CreateTable
from sqlalchemy.ext.hybrid import hybrid_property

__version__ = "0.0.0"

DATE_FORMAT = '%Y-%m-%d'
DATETIME_FORMAT = DATE_FORMAT + ' %H:%M'
DB_PATH = os.path.expanduser('~/.yap.sqlite')

_sql_echo = bool(os.environ.get('YAP_SQL_ECHO'))
engine = create_engine('sqlite:///%s' % DB_PATH, echo=_sql_echo)
Session = sessionmaker(bind=engine)
Base = declarative_base()


class TodoNotFoundError(ValueError):
    def __init__(self, tid):
        super(TodoNotFoundError, self).__init__("todo id not found: %s" % tid)


class Todo(Base):
    __tablename__ = 'todo'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    due_date = Column(DateTime)
    wait_date = Column(DateTime)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    done_at = Column(DateTime)

    @hybrid_property
    def done(self):
        return self.done_at != None

    @property
    def overdue(self):
        if self.due_date is None:
            return False
        return self.due_date < datetime.utcnow()

    @property
    def remaining(self):
        if self.due_date is None:
            return None
        return self.due_date - datetime.utcnow()

    @hybrid_property
    def waiting(self):
        if self.wait_date is None:
            return False
        return self.wait_date > datetime.utcnow()

    @waiting.expression
    def waiting(cls):
        return and_(cls.wait_date != None, cls.wait_date > datetime.utcnow())

    @property
    def str_due_date(self):
        if self.due_date:
            due_date = human_datetime(self.due_date)
            if self.overdue:
                return red(due_date)
            if self.remaining < timedelta(days=1):
                return yellow(due_date)
            return due_date

    @property
    def str_done_at(self):
        return human_datetime(self.done_at)

    @property
    def str_wait_date(self):
        return human_datetime(self.wait_date)


def human_datetime(d):
    if d.time() == time.min:
        fmt = DATE_FORMAT
    else:
        fmt = DATETIME_FORMAT
    return d.strftime(fmt)


# Done items are moved to a separate table
# in order to reuse ids in original table.
class DoneTodo(Todo):
    __tablename__ = 'done_todo'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    due_date = Column(DateTime)
    wait_date = Column(DateTime)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)
    done_at = Column(DateTime)

    __mapper_args__ = {'concrete': True}


def setup_db():
    """Runs each schema operation and version upgrade in single transaction."""
    metadata = MetaData()
    todo = Table(
            Todo.__tablename__, metadata,
            Todo.id.copy(),
            Todo.title.copy(),
    )
    done_todo = Table(
            DoneTodo.__tablename__, metadata,
            DoneTodo.id.copy(),
            DoneTodo.title.copy(),
            DoneTodo.due_date.copy(),
            DoneTodo.wait_date.copy(),
            DoneTodo.created_at.copy(),
    )
    session = Session()
    operations = [
        (create_table, session, todo),
        (add_column, session, todo, Todo.due_date),
        (add_column, session, todo, Todo.wait_date),
        (add_column, session, todo, Todo.created_at),
        (create_table, session, done_todo),
        (add_column, session, todo, Todo.done_at),
        (add_column, session, done_todo, DoneTodo.done_at),
    ]
    current_version = session.execute("pragma user_version").fetchone()[0]
    for operation in operations[current_version:]:
        operation[0](*operation[1:])
        current_version += 1
        session.execute("pragma user_version = %d" % current_version)
        session.commit()


def create_table(session, table):
    # Table.create() does commit() implicitly, we do not want this.
    sql = str(CreateTable(table).compile(engine))
    session.execute(sql)


def add_column(session, table, column):
    # sqlalchemy has no construct for altering tables :(
    table_name = table.description
    column = column.copy()
    column_name = column.compile(dialect=session.bind.dialect)
    column_type = column.type.compile(session.bind.dialect)
    session.execute('ALTER TABLE %s ADD COLUMN %s %s' % (
        table_name, column_name, column_type))


def do_maintenance():
    pass


def get_smallest_empty_id(session, model):
    i = 1
    all_items = session.query(model).order_by(model.id.asc()).all()
    all_ids = set([x.id for x in all_items])
    while i in all_ids:
        i += 1
    return i


def red(s):
    return "\033[91m {}\033[00m".format(s)


def yellow(s):
    return "\033[93m {}\033[00m".format(s)


def cmd_version(_):
    print __version__


def cmd_add(args):
    session = Session()
    todo = Todo()
    todo.id = get_smallest_empty_id(session, Todo)
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


def _due_date_first(a, b):
    """Compare function for sorting Todo items.
    Puts items with due date above items with no due date.
    """
    if a.due_date is None and b.due_date is not None:
        return 1
    if a.due_date is not None and b.due_date is None:
        return -1
    return cmp((a.due_date, a.created_at), (b.due_date, b.created_at))


def cmd_list(args):
    session = Session()
    query = session.query(Todo)
    if args.done:
        query = query.filter(Todo.done == True).union(session.query(DoneTodo))
        query = query.order_by(Todo.done_at.desc())
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


def run_script(script):  # TODO
    return subprocess.check_output(
            ['osascript', '-'], stdin=subprocess.PIPE, stderr=subprocess.PIPE)


def display_notification(message, title):  # TODO
    pass


def strdate(s):
    return datetime.strptime(s, DATE_FORMAT)


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


if __name__ == '__main__':
    setup_db()
    do_maintenance()
    parse_args()
