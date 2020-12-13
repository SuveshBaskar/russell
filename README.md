# Russell

A CLI Tool built on GoLang, that can help you to extract the compose file for the stacks that you have deployed in Docker.


#### Build

First choose the platform on which this will be going to be run, for example you need to compile the program of a specific OS and CPU Arch.
```bash
# Local Env
go build                                                    // russell bin

# Get the OS and CPU Arch where Docker is Installed
docker version --format '{{.Server.Os}}/{{.Server.Arch}}'   // linux/amd64

# Build with above parameters
env GOOS="linux" GOARCH="amd64" go build                    // russell bin
```

#### Usage
```bash
./russell -n portainer                                      // prints compose file in terminal(stdout)
./russell -n portainer -o portainer.yml                     // saves compose file to portainer.yml file in the same directory
```

#### Dependencies
```bash
Docker Go SDK
```

#### RoadMap
```bash
1) Adding more info on the output file
2) Adding Test Cases
```