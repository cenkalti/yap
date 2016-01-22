#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add deneme bir ki
    yap list
    yap done 1

"""
import os.path
import argparse
import subprocess
from datetime import datetime
from functools import partial

from tabulate import tabulate
from sqlalchemy import create_engine, MetaData, Table
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy import Column, Integer, String, Boolean, Date, DateTime

__version__ = "0.0.0"

DATE_FORMAT = '%Y-%m-%d'
DB_PATH = os.path.expanduser('~/.yap.sqlite')

Base = declarative_base()
Session = sessionmaker()


class Todo(Base):
    __tablename__ = 'todo'

    id = Column(Integer, primary_key=True)
    title = Column(String, nullable=False)
    done = Column(Boolean, nullable=False, default=False)
    due_date = Column(Date)
    start_date = Column(Date)
    created_at = Column(DateTime, nullable=False, default=datetime.utcnow)


def setup_db():
    metadata = MetaData()
    todo = Table(
            'todo', metadata,
            Todo.id.copy(),
            Todo.title.copy(),
            Todo.done.copy(),
    )
    session = Session()
    operations = [
        partial(todo.create, bind=session.bind),
        partial(add_column, session, todo, Todo.due_date),
        partial(add_column, session, todo, Todo.start_date),
        partial(add_column, session, todo, Todo.created_at),
    ]
    current_version = session.execute("pragma user_version").fetchone()[0]
    for operation in operations[current_version:]:
        operation()
        current_version += 1
        session.execute("pragma user_version = %d" % current_version)
        session.commit()


def add_column(session, table, column):
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
    session.close()


def cmd_list(args):
    session = Session()
    items = session.query(Todo)\
        .filter(Todo.done == args.done)\
        .order_by(Todo.due_date.asc(), Todo.created_at.asc())\
        .all()
    table = [[t.id, t.start_date, t.due_date, t.title] for t in items]
    print tabulate(table, headers=['ID', 'Start Date', u'Due date ▾', 'Title'])
    session.close()


def cmd_done(args):
    session = Session()
    session.query(Todo).filter(Todo.id == args.id).update({Todo.done: True})
    session.commit()
    session.close()


def cmd_undone(args):
    session = Session()
    session.query(Todo).filter(Todo.id == args.id).update({Todo.done: False})
    session.commit()
    session.close()


def cmd_remove(args):
    session = Session()
    session.query(Todo).filter(Todo.id == args.id).delete()
    session.commit()
    session.close()


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

    parser_done = subparsers.add_parser('done')
    parser_done.set_defaults(func=cmd_done)
    parser_done.add_argument('id', type=int)

    parser_undone = subparsers.add_parser('undone')
    parser_undone.set_defaults(func=cmd_undone)
    parser_undone.add_argument('id', type=int)

    parser_remove = subparsers.add_parser('remove')
    parser_remove.set_defaults(func=cmd_remove)
    parser_remove.add_argument('id', type=int)

    parser_list = subparsers.add_parser('list')
    parser_list.set_defaults(func=cmd_list)
    parser_list.add_argument('-d', '--done', action='store_true')

    parser_daemon = subparsers.add_parser('daemon')
    parser_daemon.set_defaults(func=cmd_daemon)

    args = parser.parse_args()
    args.func(args)


def main():
    engine = create_engine('sqlite:///%s' % DB_PATH, echo=False)
    Session.configure(bind=engine)
    setup_db()
    parse_args()


if __name__ == '__main__':
    main()
