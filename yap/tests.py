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

    def invoke(self, *args):
        args = ('--home', self.home) + args
        return self.runner.invoke(cli, args)

    def test_version(self):
        result = self.invoke('--version')
        self.assertEqual(result.exit_code, 0)
        self.assertIn(yap.__version__, result.output)

    def test_list(self):
        task = Task()
        task.title = 'test'
        session = Session()
        session.add(task)
        session.commit()
        result = self.invoke('list')
        self.assertEqual(result.exit_code, 0)
        self.assertIn('test', result.output)
