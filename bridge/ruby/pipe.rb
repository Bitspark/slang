require "readline"
require "json"
require_relative "{{ .OpFile }}"

if __FILE__ == $0
    loop do
        request_json = STDIN.gets
        response_json = JSON.dump({{ .Method }}(JSON.parse(request_json)))
        STDOUT.puts response_json
        STDOUT.flush
    end
end
