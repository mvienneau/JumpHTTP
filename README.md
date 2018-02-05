# JumpHTTP

An HTTP server to handle hash, shutdown, and other requests

# Comments

Hash: In keeping the socket open, I figured 'sleeping' for 5 seconds would indeed leave the socket open.
This may not be the best way, or keep it open in the way wanted, but it is the best I could do with how I
interpreted that part


Shutdown: To provide a graceful shutdown, Golang as of 1.8 has included the Shutdown function on an http server. This however does not wait for sockets to close, so inorder to handle all remaining requests to /hash, I implemented WaitGroups and have a go routine waiting on all WaitGroups to decrement before calling Shutdown.
