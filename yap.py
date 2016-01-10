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
import sqlite3
import argparse
from datetime import datetime

from tabulate import tabulate

__version__ = "0.0.0"

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


def cmd_version(_):
    print __version__


def cmd_add(args):
    print "id: %d" % conn.execute(
            "insert into todo(title, due_date, start_date) values(?, ?, ?)",
            (' '.join(args.title), args.due, args.start)).lastrowid


def cmd_list(args):
    table = []
    for row in conn.execute(
            "select id, title, start_date, due_date from todo "
            "where done=? order by due_date", (int(args.done), )):
        table.append([
            row['id'], row['start_date'],
            row['due_date'], row['title']])
    print tabulate(table, headers=[
        'ID', 'Start Date', u'Due date ▾', 'Title'])


def cmd_done(args):
    conn.execute("update todo set done=1 where id=?", (args.id, ))


def cmd_undone(args):
    conn.execute("update todo set done=0 where id=?", (args.id, ))


def cmd_remove(args):
    conn.execute("delete from todo where id=?", (args.id, ))


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

    args = parser.parse_args()
    args.func(args)


if __name__ == '__main__':
    conn = sqlite3.connect(os.path.expanduser('~/.yap.db'))
    conn.row_factory = sqlite3.Row
    setup_db()
    parse_args()
    conn.commit()
    conn.close()
