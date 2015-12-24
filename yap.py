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
c = None


def create_db():
    c.execute('''create table if not exists todo (
        id integer primary key,
        title text not null,
        done boolean not null default 0)''')


def cmd_add(title):
    c.execute("insert into todo(title) values(?)", (title,))


def cmd_list():
    for row in c.execute("select id, title, done from todo"):
        check = "✓" if row[2] == 1 else " "
        print check, row[0], row[1]


def cmd_done(todo_id):
    c.execute("update todo set done=1 where id=?", (todo_id,))


def main():
    global conn, c
    conn = sqlite3.connect(os.path.expanduser("~/.yap.db"))
    c = conn.cursor()
    create_db()

    cmd = sys.argv[1]
    if cmd == "add":
        title = sys.argv[2]
        cmd_add(title)
    elif cmd == "list":
        cmd_list()
    elif cmd == "done":
        todo_id = int(sys.argv[2])
        cmd_done(todo_id)
    else:
        raise ValueError

    conn.commit()
    conn.close()


if __name__ == "__main__":
    main()
