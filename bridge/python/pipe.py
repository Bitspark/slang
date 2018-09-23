import json, sys
from {{ .OpFile }} import {{ .Method }}

if __name__ == '__main__':
    while True:
        request_json = sys.stdin.readline()
        if not request_json:
            break
        response_json = json.dumps({{ .Method }}(json.loads(request_json)))
        print response_json
        sys.stdout.flush()