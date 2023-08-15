const Websocket = require('ws');
const server = new Websocket.Server({
    port: 1337
});

server.on('connection', function (ws) {
    ws.on('message', function (message) {
        console.log("[+] Message being spread ("+message.toString('utf8')+")");
        server.clients.forEach(function (client) {
            client.send(message);
        });
    });
});
