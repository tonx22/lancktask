## lancktack
Task implementation for lanck

### Getting Started
    docker-compose --project-name="lancktest" up -d

Server and client are implemented as a single project. The client library is located in the 'client' folder. Running in server mode does not require commands (with default settings). Run in client mode example:

`./testapp -mode client -search 3556800001,50223290002,4673101002003 -type streaming`