.ONESHELL:

varGoos=linux
varGoosArch=arm64
varExeName=bootstrap
varEntryFileName=main.go
varDeploymentFolder=deployments

compile-get:
	@echo "Compiling get package"
	@cd ./get/
	GOOS=$(varGoos) GOARCH=$(varGoosArch) go build -tags lambda.norpc -o $(varExeName) $(varEntryFileName)

compile-new:
	@echo "Compiling new package"
	@cd ./new/
	GOOS=$(varGoos) GOARCH=$(varGoosArch) go build -tags lambda.norpc -o $(varExeName) $(varEntryFileName)

zip:
	zip -j ./$(varDeploymentFolder)/get_deployment.zip ./get/bootstrap
	zip -j ./$(varDeploymentFolder)/new_deployment.zip ./new/bootstrap

compile:
	mkdir -p deployments
	@echo "Compiling..."
	$(MAKE) compile-get
	$(MAKE) compile-new
	@echo "Zip..."
	$(MAKE) zip
	@echo "Done 🚀"
