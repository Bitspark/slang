"""Record source revisions and actual local image content IDs after activation."""
import datetime
import json
from pathlib import Path
import subprocess
import sys

source, tag = sys.argv[1:]
manifest = {'recorded_at': datetime.datetime.now(datetime.timezone.utc).isoformat(),
    'tag': tag, 'sources': json.loads(Path(source).read_text()), 'images': {}}
for service in ('auth','repo','deployment','usage','vision','vision-renderer','search','meta','transit','customer','narrow'):
    name = 'slang-' + service + ':' + tag
    manifest['images'][name] = subprocess.check_output(['docker','image','inspect','--format','{{.Id}}',name],text=True).strip()
Path('/srv/slang/control-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('Recorded source revisions and immutable image content IDs')
