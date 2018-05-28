default: bin

.PHONY: std-bin
std-bin: clean
	go build

.PHONY: bin
bin: std-bin
	upx logspout

.PHONY: clean-bin
clean-bin:
	rm -f logspout

.PHONY: clean-logs
clean-logs:
	rm -f *.log

.PHONY: clean
clean: clean-bin clean-logs

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
	@sedi -i '' -e 's/const logspoutVersion = ".*"/const logspoutVersion = "${VERSION}"/' version.go
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
