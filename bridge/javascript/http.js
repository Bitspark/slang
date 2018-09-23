var http = require('http');
var op = require('./{{ .OpFile }}');

var server  = http.createServer(function (req, res) {
    var requestBody = '';
    req.on('data', function (data) {
        requestBody += data;
    });
    req.on('end', function () {
        res.writeHead(200, {'Content-Type': 'text/json'});
        var responseBody = JSON.stringify(op.{{ .Method }}(JSON.parse(requestBody)));
        res.write(responseBody);
        res.end();
    });
});
server.listen(0);
server.on('listening', function() {
    console.log('http://localhost:' + server.address().port)
});