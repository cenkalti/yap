import yap.db
import yap.parser


def main():
    yap.db.setup()
    yap.parser.parse_args()

if __name__ == '__main__':
    main()
