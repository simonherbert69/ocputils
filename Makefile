
# Basic go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install

listocpgroups :
	$(GOBUILD) -o listocpgroups cmd/listocpgroups.go

listocpprojects :
	$(GOBUILD) -o listocpprojects cmd/listocpprojects.go

extractprojectsetups :
	$(GOBUILD) -o extractprojectsetups cmd/extractprojectsetups.go

all: listocpgroups listocpprojects extractprojectsetups

clean :
	rm -f listocpgroups listocpprojects extractprojectsetups

install-all:
	$(GOINSTALL) cmd/listocpgroups.go
	$(GOINSTALL) cmd/listocpprojects.go
	$(GOINSTALL) cmd/extractprojectsetups.go
