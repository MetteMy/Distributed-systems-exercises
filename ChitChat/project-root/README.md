Usage:
1. cd to server
2. From server directory, run 'go run server.go'
3. Then, open a new terminal.
4. cd to client
5. From client directory, run 'go run client.go [name]' (replace [name] with a string representing the username of 
the person joining ChitChat)
5b. If you are running the client on a different machine then the server, you can specify an ip address like this: 'go run client.go [name] [ip]'. the ip address can be found with the command 'ipconfig' in powershell

6. If you want, you can currently open as many new client-sessions as you want. 
7. To terminate a client, press ctrl+c.