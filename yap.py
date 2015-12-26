#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add deneme bir ki
    yap list
    yap done 1

"""
import sys
import os.path
import sqlite3
from datetime import datetime

from tabulate import tabulate

DATE_FORMAT = '%Y-%m-%d'

conn = None


class Todo(object):
    def __init__(self, row):
        self.id = row['id']
        self.title = row['title']
        self.done = bool(row['done'])
        if row['due_date']:
            self.due_date = datetime.strptime(row['due_date'], DATE_FORMAT)
        else:
            self.due_date = None


def setup_db():
    current_version = conn.execute("pragma user_version").fetchone()[0]
    statements = [
        '''create table todo (
        id integer primary key,
        title text not null,
        done boolean not null default 0)''',
        '''alter table todo add column due_date text''',
    ]
    for statement in statements[current_version:]:
        conn.execute(statement)
        current_version += 1
        conn.execute("pragma user_version = %d" % current_version)


def parse_args(args):
    main, flags = [], {}
    while args:
        head = args[0]
        if head.startswith('-'):
            flags[head] = args[1]
            args = args[2:]
        else:
            main.append(head)
            args = args[1:]
    return main, flags


def cmd_add(args):
    main, flags = parse_args(args)
    title = ' '.join(main)
    due_date = None
    for flag, value in flags.items():
        if flag in ('-d', '--due'):
            due_date = datetime.strptime(value, DATE_FORMAT).strftime(DATE_FORMAT)
    print "id: %d" % conn.execute(
            "insert into todo(title, due_date) values(?, ?)",
            (title, due_date)).lastrowid


def cmd_list(_):
    table = []
    for row in conn.execute("select id, title, done, due_date from todo "
                            "order by due_date desc"):
        todo = Todo(row)
        check = u'✓' if todo.done else ''
        if todo.due_date:
            due_date = todo.due_date.strftime(DATE_FORMAT)
        else:
            due_date = ''
        table.append([todo.id, check, due_date, todo.title])
    print tabulate(table, headers=['ID', 'Done', u'Due date ▾', 'Title'])


def cmd_done(args):
    todo_id = int(args[0])
    conn.execute("update todo set done=1 where id=?", (todo_id, ))


def cmd_undone(args):
    todo_id = int(args[0])
    conn.execute("update todo set done=0 where id=?", (todo_id, ))


def cmd_remove(args):
    todo_id = int(args[0])
    conn.execute("delete from todo where id=?", (todo_id, ))


if __name__ == '__main__':
    conn = sqlite3.connect(os.path.expanduser('~/.yap.db'))
    conn.row_factory = sqlite3.Row
    setup_db()
    cmd, options = sys.argv[1], sys.argv[2:]
    {
        'add': cmd_add,
        'list': cmd_list,
        'done': cmd_done,
        'undone': cmd_undone,
        'remove': cmd_remove,
    }[cmd](options)
    conn.commit()
    conn.close()
