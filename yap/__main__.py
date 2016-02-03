import yap.db
import yap.commands


def main():
    yap.db.setup()
    yap.commands.parse_args()

if __name__ == '__main__':
    main()
