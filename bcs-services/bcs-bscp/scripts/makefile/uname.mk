# only support Linux/Darwin
UNAME := $(shell uname)

ifeq ($(UNAME),Linux)
	GENERATOR := $(SCRIPTS)/install/generator_linux.sh
else ifeq ($(UNAME),Darwin)
	GENERATOR := $(SCRIPTS)/install/generator_darwin.sh
else
$(error not support ${UNAME} build)
endif