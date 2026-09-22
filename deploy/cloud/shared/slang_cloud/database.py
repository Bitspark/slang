import os

from playhouse.db_url import connect
from telstar.com.pw import StagedMessage


def bind_models(models, staged=True):
    db = connect(os.environ['DATABASE'])
    db.bind([*models, *([StagedMessage] if staged else [])])
    if os.environ.get('INIT_SCHEMA') == 'true':
        db.create_tables([*models, *([StagedMessage] if staged else [])], safe=True)
    return db


def request_connections(app, db):
    @app.before_request
    def open_connection():
        db.connect(reuse_if_open=True)

    @app.teardown_request
    def close_connection(error):
        if not db.is_closed():
            db.close()
