# NDSLabs iRODS Federation Server

Prototype server that handles federation requests for an iRODS iCAT server, used in NDS Labs. This is included in the iRODS [iCAT](https://github.com/nds-org/ndslabs-irods/tree/master/dockerfiles/icat) image to allow authorized iCAT images to federate.

# Latest release

The latest release is available under [releases](https://github.com/nds-org/ndslabs-irods-federate/releases)

# Building

Prerequisites:
* Assumes running under OS X 
* Go 1.5

To simply compile the application:
```
go build
```

To build for multiple architectures
```
./build.sh
```

# Running

```
./ndslabs-irods-federate --host localhost --port 8080 --password admin --zone tempZone
```

# See also
* https://github.com/nds-org/ndslabs-irods/
