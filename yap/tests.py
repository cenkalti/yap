import unittest

from click.testing import CliRunner

import yap
from yap.commands import cli
from yap.models import Task, Session


class YapTestCase(unittest.TestCase):

    def setUp(self):
        self.runner = CliRunner()
        self.ifs = self.runner.isolated_filesystem()
        self.home = self.ifs.__enter__()
        cli.setup_db(self.home)

    def tearDown(self):
        self.ifs.__exit__(None, None, None)

    def invoke(self, args):
        args = ['--home', self.home] + args.split()
        result = self.runner.invoke(cli, args)
        self.assertEqual(result.exit_code, 0)
        return result

    def test_version(self):
        result = self.invoke('--version')
        self.assertIn(yap.__version__, result.output)

    def test_list(self):
        task = Task()
        task.title = 'test'
        session = Session()
        session.add(task)
        session.commit()
        result = self.invoke('list')
        self.assertIn('test', result.output)

    def test_add(self):
        self.invoke('add deneme')
        session = Session()
        task = session.query(Task).one()
        self.assertEqual(task.title, 'deneme')

    def test_archive(self):
        """archived tasks should not appear in list output"""
        task = Task()
        task.title = 'test'
        session = Session()
        session.add(task)
        session.commit()
        self.invoke('archive 1')
        result = self.invoke('list')
        self.assertNotIn('test', result.output)

    def test_delete(self):
        task = Task()
        task.title = 'test'
        session = Session()
        session.add(task)
        session.commit()
        self.invoke('delete 1')
        task = session.query(Task).first()
        self.assertIsNone(task)
