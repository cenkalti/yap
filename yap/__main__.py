import yap.db
import yap.commands

if __name__ == '__main__':
    yap.db.setup()
    yap.db.do_maintenance()
    yap.commands.parse_args()
