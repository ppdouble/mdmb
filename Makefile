VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

MDMB=\
	mdmb-darwin-amd64 \
	mdmb-linux-amd64 \
	mdmb-windows-amd64.exe

MDMBAPNS=\
	mdmbapns-darwin-amd64 \
	mdmbapns-linux-amd64 \
	mdmbapns-windows-amd64.exe


#myapns: mdmbapns-$(OSARCH)
#	$(info $$MDMBAPNS = $(MDMBAPNS))

myb: mdmb-$(OSARCH)
	$(info $$MDMB = $(MDMB))

$(MDMB): cmd/mdmb
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<
#	$(info $$GOOS = $(GOOS))

$(MDMBAPNS): cmd/mdmbapns
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

%-$(VERSION).zip: %.exe
	rm -f $@
	zip $@ $<

%-$(VERSION).zip: %
	rm -f $@
	zip $@ $<

clean:
	rm -f mdmb-*
	rm -f mdmbpans-*

cleanmdmb:
	rm -f mdmb-*

cleanmdmapns:
	rm -f mdmbapns-*

#release:
#	$(foreach bin,$(MDMB),$(subst .exe,,$(bin))-$(VERSION).zip)
#	$(foreach bin,$(MDMBAPNS),$(subst .exe,,$(bin))-$(VERSION).zip)

rmdmb:
	$(foreach bin,$(MDMB),$(subst .exe,,$(bin))-$(VERSION).zip)


rmdmbapns:
	$(foreach bin,$(MDMBAPNS),$(subst .exe,,$(bin))-$(VERSION).zip)
	$(info $$MDMBAPNS = $(MDMBAPNS))

mdmbbuild: myb $(MDMB) cleanmdmb rmdmb

mdmapnsbuild: myapns $(MDMBAPNS) cleanmdmapns rmdmbapns

mdmbbld: cleanmdmb myb $(MDMB)

mdmapnsbld: cleanmdmapns myapns $(MDMBAPNS)

build:  mdmapnsbuild mdmbbuild

.PHONY: build
	#my $(MDMB) clean release
#	myapns $(MDMBAPNS) clean releaseapns

all: build
