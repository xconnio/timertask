# workerpool
Imagine a websocket server that sends ping message every X seconds of inactivity to each client, to be able to acheive that, the server would need to run on goroutine per client. Those goroutines are mostly idle and result in memory consumption. workpool attempts to solve that problems by running a single loop that runs every second to check which clients should be sent ping message.
