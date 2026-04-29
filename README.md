# cuddly-fishstick

An asynchronous, non-blocking webhook reciever.
- When a POST request is sent to the endpoint, the handler confirms the request method and content size.
- Parses the JSON into a generic map and pushes the payload into the processing channel.
    + processing channel is buffered for 120 payloads
- Immediately sends a 200 OK response to end the request. 
- Background processing occurs in a separate Goroutine
- Shutdown provides 5sec buffer to HTTP handlers, waits for processing channel to empty before shutdown.

An asynchronous, non-blocking websocket reciever. 
- Events recieved by the websocket conn are sent directly into the processing channel

