#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add water flowers
    yap list
    yap done 1

"""

# TODO rename start_date to waiting date
# TODO move up 3 subcommand (order column)
# TODO filters (https://taskwarrior.org/docs/syntax.html) (https://taskwarrior.org/docs/filter.html)
# TODO recurring tasks (https://taskwarrior.org/docs/recurrence.html) (https://taskwarrior.org/docs/durations.html)
# TODO search
# TODO context (https://taskwarrior.org/docs/context.html)
# TODO projects
# TODO human dates (https://taskwarrior.org/docs/dates.html) (https://taskwarrior.org/docs/named_dates.html)
# TODO move done tasks to another table
# TODO color overdue red
# TODO remove start date column from list output
# TODO auto-completing
# TODO daemon
# TODO notifications

import os
import sys
import argparse
import subprocess
from datetime import datetime

from tabulate import tabulate
from sqlalchemy import create_engine, MetaData, Table, or_
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy import Column, Integer, String, Boolean, Date, DateTime
from sqlalchemy.schema import CreateTable

__version__ = "0.0.0"

DATE_FORMAT = '%Y-%m-%d'
DB_PATH = os.path.expanduser('~/.yap.sqlite')

engine = create_engine('sqlite:///%s' % DB_PATH, echo=False)
Session = sessionmaker(bind=engine)
Base = declarative_base()


class Todo(Base):
    __tablename__ = 'todo'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    done = Column(Boolean, nullable=False, default=False)
    due_date = Column(Date)
    start_date = Column(Date)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)


def setup_db():
    """Runs each schema operation and version upgrade in single transaction."""
    todo = Table(
            'todo', MetaData(),
            Todo.id.copy(),
            Todo.title.copy(),
            Todo.done.copy(),
    )
    session = Session()
    operations = [
        (create_table, session, todo),
        (add_column, session, todo, Todo.due_date),
        (add_column, session, todo, Todo.start_date),
        (add_column, session, todo, Todo.created_at),
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


def cmd_version(_):
    print __version__


def cmd_add(args):
    todo = Todo()
    todo.title = ' '.join(args.title)
    todo.due_date = args.due
    todo.start_date = args.start
    session = Session()
    session.add(todo)
    session.commit()
    print "id: %d" % todo.id


def cmd_edit(args):
    session = Session()
    todo = session.query(Todo).get(args.id)
    if not todo:
        raise ValueError("id not found")
    if args.title:
        todo.title = ' '.join(args.title)
    if args.due:
        todo.due_date = args.due
    if args.start:
        todo.start_date = args.start
    session.commit()


def cmd_list(args):
    def due_date_first(a, b):
        """Compare function for sorting Todo items.
        Puts items with due date above items with no due date.
        """
        if a.due_date is None and b.due_date is not None:
            return 1
        if a.due_date is not None and b.due_date is None:
            return -1
        return cmp((a.due_date, a.created_at), (b.due_date, b.created_at))
    session = Session()
    query = session.query(Todo).filter(Todo.done == args.done)
    if args.start:
        query = query.filter(Todo.start_date > datetime.today())
    else:
        query = query.filter(or_(Todo.start_date == None,
                                 Todo.start_date <= datetime.today()))
    items = query.all()
    # Cannot get items with due date first with simple ORDER BY.
    items.sort(cmp=due_date_first)
    table = [[t.id, t.start_date, t.due_date, t.title] for t in items]
    print tabulate(table, headers=['ID', 'Start Date', u'Due date ▾', 'Title'])
    session.close()


def cmd_done(args):
    session = Session()
    session.query(Todo).filter(Todo.id.in_(args.id))\
        .update({Todo.done: True}, synchronize_session=False)
    session.commit()


def cmd_undone(args):
    session = Session()
    session.query(Todo).filter(Todo.id.in_(args.id))\
        .update({Todo.done: False}, synchronize_session=False)
    session.commit()


def cmd_remove(args):
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
    parser_add.add_argument('-d', '--due', type=strdate)
    parser_add.add_argument('-s', '--start', type=strdate)

    parser_edit = subparsers.add_parser('edit')
    parser_edit.set_defaults(func=cmd_edit)
    parser_edit.add_argument('id', type=int)
    parser_edit.add_argument('-t', '--title', nargs='+')
    parser_edit.add_argument('-d', '--due', type=strdate)
    parser_edit.add_argument('-s', '--start', type=strdate)

    parser_done = subparsers.add_parser('done')
    parser_done.set_defaults(func=cmd_done)
    parser_done.add_argument('id', type=int, nargs='+')

    parser_undone = subparsers.add_parser('undone')
    parser_undone.set_defaults(func=cmd_undone)
    parser_undone.add_argument('id', type=int, nargs='+')

    parser_remove = subparsers.add_parser('remove')
    parser_remove.set_defaults(func=cmd_remove)
    parser_remove.add_argument('id', type=int, nargs='+')

    parser_list = subparsers.add_parser('list')
    parser_list.set_defaults(func=cmd_list)
    parser_list.add_argument('-d', '--done', action='store_true',
                             help="show done items")
    parser_list.add_argument('-s', '--start', action='store_true',
                             help="show items with forward start date")

    parser_daemon = subparsers.add_parser('daemon')
    parser_daemon.set_defaults(func=cmd_daemon)

    # If invoked with no subcommand, run list subcommand
    if len(sys.argv) == 1:
        sys.argv.append('list')

    args = parser.parse_args()
    args.func(args)


if __name__ == '__main__':
    setup_db()
    parse_args()
