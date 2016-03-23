BUILD_DATE := `date +%Y-%m-%d\ %H:%M`
VERSIONFILE := version.go

gensrc:
	rm -f $(VERSIONFILE)
	@echo "package main" > $(VERSIONFILE)
	@echo "const (" >> $(VERSIONFILE)
	@echo "  VERSION = \"0.1\"" >> $(VERSIONFILE)
	@echo "  BUILD_DATE = \"$(BUILD_DATE)\"" >> $(VERSIONFILE)
	@echo ")" >> $(VERSIONFILE)
	mkdir -p build/bin/amd64 build/bin/darwin build/pkg
	docker run --rm -it -v `pwd`:/go/src/github.com/ndslabs-irods-federate/ -v `pwd`/build/bin:/go/bin -v `pwd`/build/pkg:/go/pkg -v `pwd`/build.sh:/build.sh golang  /build.sh
