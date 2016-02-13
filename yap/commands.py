import os
import json
import errno
from datetime import datetime, date, time, timedelta

import click
from sqlalchemy import create_engine, case
from sqlalchemy.exc import SQLAlchemyError
from tabulate import tabulate

import yap
import yap.db
import yap.json_util
from yap.models import Task, Session
from yap.types import delete_option, due_date, wait_date, on_date, duration


@click.group('yap', invoke_without_command=True,
             context_settings={'help_option_names': ['-h', '--help']})
@click.version_option(version=yap.__version__)
@click.option('--home', default="~", type=click.Path(file_okay=False),
              help="Override directory for database file, default is user home")
@click.pass_context
def cli(ctx, home):
    home = os.path.expanduser(home)
    ctx.obj = {'home': home}
    cli.setup_db(home)
    if ctx.invoked_subcommand is None:
        ctx.invoke(next_)


def setup_db(home):
    db_path = os.path.join(home, '.yap.sqlite')
    sql_echo = bool(os.environ.get('YAP_SQL_ECHO'))
    engine = create_engine('sqlite:///%s' % db_path, echo=sql_echo)
    Session.configure(bind=engine)
    yap.db.setup()
cli.setup_db = setup_db


@cli.command('list', short_help="List tasks")
@click.option('-c', '--context', help="only items in context")
@click.option('-d', '--done', is_flag=True, help="only done tasks")
@click.option('-w', '--waiting', is_flag=True, help="only waiting tasks")
@click.option('-a', '--archived', is_flag=True, help="only archived tasks")
def list_(context, done, waiting, archived, only_next=False):
    session = Session()
    query = session.query(Task)

    headers = ['ID']
    attrs = ['id']

    ctx = context or _get_context()
    if ctx:
        query = query.filter(Task.context == ctx)

    if done:
        headers.append('Done at')
        attrs.append('str_done_at')
        query = query\
            .filter(Task.done == True)\
            .filter(Task.archived != True)\
            .order_by(Task.done_at.desc())\
            .limit(yap.LIST_DONE_MAX)
    elif waiting:
        headers.append('Wait date')
        attrs.append('str_wait_date')
        query = query\
            .filter(Task.waiting == True)\
            .filter(Task.archived != True)\
            .order_by(Task.wait_date)
    elif archived:
        query = query.filter(Task.archived == True)\
            .order_by(Task.created_at.desc())
    else:
        query = query\
            .filter(Task.done != True, Task.waiting != True)\
            .filter(Task.archived != True)\
            .order_by(  # Show tasks with order date first
                case([(Task.due_date == None, 0)], 1),
                Task.due_date,
                case([(Task.order == None, 1)], 0),
                Task.order.asc())

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


@cli.command('next', short_help="List next tasks to do")
@click.pass_context
def next_(ctx):
    ctx.invoke(list_, only_next=True)


@cli.command(short_help="Show task detail")
@click.argument('id', type=click.INT)
def show(id):
    session = Session()
    task = session.query(Task).get(id)
    if not task:
        raise click.ClickException("task not found")
    for k, v in sorted(vars(task).items()):
        if not k.startswith('_'):
            print "%s: %s" % (k, v)
    session.close()


@cli.command(short_help="Add new task")
@click.argument('title', nargs=-1)
@click.option('-d', '--due', type=due_date,
              help="do the task before due date")
@click.option('-w', '--wait', type=wait_date,
              help="do not show the task before wait date")
@click.option('-o', '--on', type=on_date,
              help="set due date and wait date to same day")
@click.option('-r', '--recur', type=duration,
              help="repeat the task periodically")
@click.option('-s', '--shift', is_flag=True,
              help="when the task is done forward due date "
                   "exactly one recurrence period after done date")
@click.option('-c', '--context',
              help="set a context other than current context")
def add(title, due, wait, on, recur, shift, context):
    session = Session()
    task = Task()
    task.id = yap.db.get_smallest_empty_id(session, Task)
    task.title = ' '.join(title)
    if on:
        task.due_date = datetime.combine(on, time.max)
        task.wait_date = datetime.combine(on, time.min)
    else:
        task.due_date = due
        task.wait_date = wait
    task.context = context or _get_context()
    task.recur = recur
    if task.recur:
        task.shift = shift
    task.order = Task.find_next_order(session)
    session.add(task)
    session.commit()
    print "id: %d" % task.id


@cli.command(short_help="Edit task")
@click.argument('id', type=click.INT)
@click.option('-t', '--title', help="replace title")
@click.option('-a', '--append', help="append text to title")
@click.option('-p', '--prepend', help="prepend text to title")
@click.option('-d', '--due', type=due_date)
@click.option('-w', '--wait', type=wait_date)
@click.option('-o', '--on', type=on_date)
@click.option('-r', '--recur', type=duration)
@click.option('-s', '--shift', is_flag=True)
@click.option('-c', '--context')
def edit(id, title, append, prepend, due, wait, on, recur, shift, context):
    session = Session()
    task = session.query(Task).get(id)
    if not task:
        raise click.ClickException("task not found")

    if title:
        task.title = None if title == delete_option else ' '.join(title)
    if append:
        task.title = "%s %s" % (task.title, ' '.join(append))
    if prepend:
        task.title = "%s %s" % (' '.join(prepend), task.title)
    if on:
        if on == delete_option:
            task.due_date = None
            task.wait_date = None
        else:
            task.due_date = datetime.combine(on, time.max)
            task.wait_date = datetime.combine(on, time.min)
    if due:
        task.due_date = None if due == delete_option else due
    if wait:
        task.wait_date = None if wait == delete_option else wait
    if context:
        task.context = None if context == delete_option else context
    if recur:
        task.recur = None if recur == delete_option else recur
    if shift is not None:
        task.shift = shift

    session.commit()


@cli.command('done', short_help="Mark task as done")
@click.argument('id', type=click.INT, nargs=-1)
def done_(id):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(id)).all()
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


@cli.command(short_help="Mark task as undone")
@click.argument('id', type=click.INT, nargs=-1)
def undone(id):
    session = Session()
    tasks = session.query(Task).filter(Task.id.in_(id)).all()
    for task in tasks:
        task.id = yap.db.get_smallest_empty_id(session, Task)
        task.done_at = None
    session.commit()


@cli.command(short_help="Delete task")
@click.argument('id', type=click.INT, nargs=-1)
def delete(id):
    session = Session()
    session.query(Task).filter(Task.id.in_(id))\
        .delete(synchronize_session=False)
    session.commit()


@cli.command(short_help="Archive task")
@click.argument('id', type=click.INT, nargs=-1)
def archive(id):
    session = Session()
    session.query(Task).filter(Task.id.in_(id))\
        .update({Task.id: yap.db.get_next_negative_id(session, Task)},
                synchronize_session=False)
    session.commit()


@cli.command('wait', short_help="Hide task until date")
@click.argument('wait_date', type=wait_date)
@click.argument('id', type=click.INT, nargs=-1)
def wait_(wait_date, id):
    session = Session()
    session.query(Task).filter(Task.id.in_(id))\
        .update({Task.wait_date: wait_date}, synchronize_session=False)
    session.commit()


@cli.command(short_help="Postpone due date")
@click.argument('due_date', type=due_date)
@click.argument('id', type=click.INT, nargs=-1)
def postpone(due_date, id):
    session = Session()
    session.query(Task).filter(Task.id.in_(id))\
        .update({Task.due_date: due_date}, synchronize_session=False)
    session.commit()


@cli.command('context', short_help="Get or set the context")
@click.argument('name', required=False)
@click.option('-c', '--clear', is_flag=True)
def context_(name, clear):
    if name:
        with open(_get_context_path(), 'w') as f:
            f.write(name)
    elif clear:
        os.unlink(_get_context_path())
    else:
        print _get_context()


def _get_context():
    try:
        with open(_get_context_path(), 'r') as f:
            return f.read()
    except IOError as e:
        if e.errno != errno.ENOENT:
            raise


def _get_context_path():
    ctx = click.get_current_context()
    home = ctx.obj['home']
    return os.path.join(home, '.yap.context')


@cli.command(short_help="Export database as JSON")
@click.argument('output', type=click.File('w'))
def export(output):
    session = Session()
    d = {'task': [task.to_dict() for task in session.query(Task).all()]}
    session.close()
    out = json.dumps(d, default=yap.json_util.datetime_encoder, indent=2)
    output.write(out)
    output.write('\n')
    output.close()


@cli.command('import', short_help="Import tasks from JSON")
@click.argument('input', type=click.File('r'))
def import_(input):
    has_error = False
    d = json.load(input, object_hook=yap.json_util.datetime_decoder)
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
        raise click.ClickException("import completed with errors")
