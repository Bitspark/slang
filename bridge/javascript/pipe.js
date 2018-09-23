var readline = require('readline');
var op = require('./{{ .OpFile }}');

var rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false
});

rl.on('line', function(line){
    var requestJson = line;
    var responseJson = JSON.stringify(op.{{ .Method }}(JSON.parse(requestJson)));
    console.log(responseJson);
});