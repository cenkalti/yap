from sqlalchemy import MetaData, Table
from sqlalchemy.sql.ddl import CreateTable

from yap.models import Todo, DoneTodo, engine, Session


def setup():
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
