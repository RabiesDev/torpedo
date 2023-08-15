const webSocket = new WebSocket("ws://127.0.0.1:1337");
webSocket.onmessage = async (event) => {
    if (event.data.size != 0) {
        return;
    }

    let elements = document.getElementsByClassName("val-2gQskI");
    webSocket.send(JSON.stringify({
        status: "established",
        pointX: elements[0].innerText,
        pointY: elements[1].innerText,
    }));
};
