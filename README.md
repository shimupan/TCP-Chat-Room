### Concurrent and Threadsafe TCP Chat Room Written in GO

### Usage
# Server

- Make sure you have ```go``` and ```Makefile``` installed and simple type
  ```
  make run_server
  ```
- This will start the tcp server on port 1337
- Commands and help lists are shown upon startup

# Client

- After the server is running, open up any amount of terminals and then run
  ```
  make run_client
  ```
- This will turn that terminal session into the "client" and connect to the server
- Commands and help lists are shown upon startup
