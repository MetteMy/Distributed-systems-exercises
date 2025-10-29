**Build and run**
1. Open terminal
2. change directory to server
2. From server directory, run 'go run server.go'. Program builds and server gets going
3. Now, you have two options for adding clients to the program:

**Option 1: Run client(s) from the same machine as the server's**
1. Open a new terminal
2. cd to client
3. From client directory, run 'go run client.go [name]' (replace [name] with a string representing the username of
   the person joining ChitChat)
4. Begin chatting!
5. When you want to leave, simply press ctrl+c

**Option 2: Run client(s) from a different machine than the server's**
1. Open a terminal
2. cd to client
3. From client directory, run 'go run client.go [name] [ip]'.
    3a. The ip address can be found with the command 'ipconfig' in powershell
4. Begin chatting!
5. When you want to leave, simply press ctrl+c


**Check out the program log**
1. The chitchat.log-file resides in the server-folder. 
