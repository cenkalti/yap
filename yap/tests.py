from click.testing import CliRunner

import yap
from yap.commands import cli


def test_version():
    runner = CliRunner()
    result = runner.invoke(cli, ['--version'])
    assert result.exit_code == 0
    assert result.output.split(' ')[-1].strip() == yap.__version__
