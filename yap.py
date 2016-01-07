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


def setup_db():
    current_version = conn.execute("pragma user_version").fetchone()[0]
    statements = [
        '''create table todo (
        id integer primary key,
        title text not null,
        done boolean not null default 0)''',
        '''alter table todo add column due_date text''',
        '''alter table todo add column start_date text''',
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


def check_date_format(s):
    return datetime.strptime(s, DATE_FORMAT).strftime(DATE_FORMAT)


def cmd_add(args):
    main, flags = parse_args(args)
    title = ' '.join(main)
    if title == '':
        raise ValueError('title required')
    due_date = None
    start_date = None
    for flag, value in flags.items():
        if flag in ('-d', '--due'):
            # check format
            due_date = check_date_format(value)
        elif flag in ('-s', '--start'):
            start_date = check_date_format(value)
        else:
            raise ValueError('unknown flag')
    print "id: %d" % conn.execute(
            "insert into todo(title, due_date, start_date) values(?, ?, ?)",
            (title, due_date, start_date)).lastrowid


def cmd_list(_):
    table = []
    for row in conn.execute("select id, title, done, start_date, due_date from todo "
                            "order by due_date desc"):
        check = u'✓' if row['done'] else ''
        table.append([row['id'], check, row['start_date'], row['due_date'], row['title']])
    print tabulate(table, headers=['ID', 'Done', 'Start Date', u'Due date ▾', 'Title'])


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
