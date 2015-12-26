#!/usr/bin/env python
# coding=utf-8
"""
Command line todo app.
Usage:
    yap add "deneme bir iki"
    yap list
    yap done 1

"""
import sys
import os.path
import sqlite3

conn = None


def setup_db():
    current_version = conn.execute("pragma user_version").fetchone()[0]
    statements = [
        '''create table if not exists todo (
        id integer primary key,
        title text not null,
        done boolean not null default 0)''',
    ]
    for statement in statements[current_version:]:
        conn.execute(statement)
        current_version += 1
        conn.execute("pragma user_version = %d" % current_version)


def cmd_add(args):
    title = " ".join(args)
    conn.execute("insert into todo(title) values(?)", (title, ))


def cmd_list(_):
    for row in conn.execute("select id, title, done from todo"):
        check = "✓" if row[2] == 1 else " "
        print check, row[0], row[1]


def cmd_done(args):
    todo_id = int(args[0])
    conn.execute("update todo set done=1 where id=?", (todo_id, ))


def cmd_undone(args):
    todo_id = int(args[0])
    conn.execute("update todo set done=0 where id=?", (todo_id,))


if __name__ == "__main__":
    conn = sqlite3.connect(os.path.expanduser("~/.yap.db"))
    setup_db()
    cmd, options = sys.argv[1], sys.argv[2:]
    {
        "add": cmd_add,
        "list": cmd_list,
        "done": cmd_done,
        "undone": cmd_undone,
    }[cmd](options)
    conn.commit()
    conn.close()
