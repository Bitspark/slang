"""Import engine-generated library bundles without overwriting user records."""
import json
import sys
import telstar
from index import db
from models import Bundle


def normalize(bundle):
    for blueprint in bundle['blueprints'].values():
        for kind in ('services', 'delegates'):
            for port in (blueprint.get(kind) or {}).values():
                if not port.get('geometry'):
                    port['geometry'] = {'in': {'position': 0}, 'out': {'position': 0}}
    return bundle


def main(directory):
    from pathlib import Path
    bundles = [(json.loads(p.read_text()), False) for p in Path(directory).glob('*.json')]
    for end, name, connections in [('1', 'Blank', {}), ('2', 'Echo', {'(': [')']})]:
        identifier = '00000000-0000-4000-8000-00000000000' + end
        bundles.append(({'main': identifier, 'blueprints': {identifier: {
            'id': identifier, 'meta': {'name': name, 'icon': 'project-diagram',
            'tags': ['template'], 'shortDescription': 'Start a new blueprint.' if end == '1' else 'Returns the input text.'},
            'services': {'main': {'in': {'type': 'string'}, 'out': {'type': 'string'}}},
            'operators': {}, 'connections': connections,
            'geometry': {'size': {'width': 440, 'height': 320}}}}}, True))
    with db.connection_context(), db.atomic():
        for definition, template in bundles:
            definition = normalize(definition)
            bundle, created = Bundle.get_or_create(operator_id=definition['main'], defaults=dict(
                owner='system', library=not template, template=template, definition=definition))
            if bundle.owner != 'system':
                raise RuntimeError('Refusing to replace a user-owned bundle')
            if not created and bundle.definition == definition:
                continue
            bundle.definition = definition
            bundle.save()
            telstar.stage('bundleCreated' if created else 'bundleUpdated', bundle.to_dict())
    print(f'Library ready: {len(bundles)} bundles including Blank and Echo templates')


if __name__ == '__main__':
    main(sys.argv[1])
