default: help

.PHONY: help
help:
	@echo "What you can do with this Makefile"
	@echo "    make mac - make the executable for macOS"
	@echo "    make linux - make the executable for Linux"
	@echo ""
	@echo "Check the Makefile for more options."

.PHONY: linux
linux: std-linux-bin
# Comment out upx as it seems causing problems.
#	upx logspout

.PHONY: std-linux-bin
std-linux-bin: clean generate
	env GOOS=linux GOARCH=amd64 go build

.PHONY: std-bin
std-bin: clean generate
	go build

.PHONY: bin
mac: std-bin
# Comment out upx as it seems causing problems.
#	upx logspout

.PHONY: clean-bin
clean-bin:
	rm -f logspout

.PHONY: clean-logs
clean-logs:
	rm -f *.log

.PHONY: clean
clean: clean-bin clean-logs

.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test

are-you-sure:
	@read -p "I suppose you know what you are doing. Are you sure? [Y/n]" -n 1 -r; \
	if [[ $$REPLY != "Y" ]]; then \
		echo -e -n "\nSee you. >> "; \
		exit 1; \
	fi
	@echo -e "\nGood, proceeding..."; \
		$(eval confirmed := "Yes")

# Change the value of initVersion and tag it with the version number.
# Usage: make VERSION=blabla tag-and-release
# TODO: It is deprecated and will be replaced by a safer way soon.
tag-and-release: are-you-sure
	# VERSION must be set
	@if [ -z "$$VERSION" ]; then \
		echo -e "The VERSION is unset or an empty string"; \
		exit 1; \
	fi
	# Let you know what you are doing.
	@echo -e "Changing the version number and tag it with ${VERSION}."
	# Stash your current work, in case of any
	@git stash
	# We tag and release on the master branch
	@git checkout master
	# Make sure it's up-to-date
	@git fetch origin
	@git reset --hard origin/master
	# Replace the version with the value provided by you
	@sed -i 's/const logspoutVersion = ".*"/const logspoutVersion = "${VERSION}"/' version.go
	# And commit it
	@git add version.go
	@git commit -m "Changed version to ${VERSION}"
	# Now we create tags as a release
	@git tag -a ${VERSION} -m "Version: ${VERSION}"
	# Push it to the remote master branch
	@git push origin master --follow-tags
	# Resume your workspace
	@git checkout ${prev_branch}
	@git stash pop
