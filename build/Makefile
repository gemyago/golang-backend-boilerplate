.PHONY: FORCE
.SECONDEXPANSION:

app_name:=$(shell basename `git rev-parse --show-toplevel`)

# Define a function to read values from any config file
# Usage: $(call read_config,config_file,key_name)
define read_config
$(shell grep -m 1 "^$(2)\s*=\s*" $(1) | cut -d'=' -f2 | sed 's/^[[:space:]]*//;s/[[:space:]]*$$//')
endef

# Reading config using the new function
docker_runtime_image:=$(call read_config,build.cfg,docker_runtime_image)
platforms:=$(call read_config,build.cfg,platforms)

comma:=,

# platforms in config are in form linux/amd64,linux/arm64...
# transforming to dist/linux/amd64 dist/linux/arm64...
dist_platforms:=$(patsubst %,dist/%,$(subst $(comma), ,$(platforms)))

# Example: dist/linux/amd64
dist/%: FORCE
	$(eval GOOS:=$(word 1,$(subst /, ,$*)))
	$(eval GOARCH:=$(word 2,$(subst /, ,$*)))
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -C .. \
			-tags=release \
			-o build/dist/$(GOOS)/$(GOARCH)/ ./cmd/...;

dist: $(dist_platforms)

.git-meta:
	@mkdir -p $@

.git-meta/repo-name: .git-meta
	@basename $(shell git rev-parse --show-toplevel) > $@

.git-meta/remote-url: .git-meta
	@git remote get-url origin | sed -E \
		-e 's|^git@([^:]+):|https://\1/|' \
		-e 's|^http:|https:|' \
		-e 's|\.git$$||' > $@.tmp
	@cmp -s $@.tmp $@ || mv $@.tmp $@
	@rm -f $@.tmp

.git-meta/HEAD: FORCE .git-meta
	@git rev-parse HEAD > $@.tmp
	@if cmp -s $@ $@.tmp; then \
		rm $@.tmp; \
	else \
		mv $@.tmp $@; \
	fi

docker/.labels: .git-meta/HEAD .git-meta/repo-name .git-meta/HEAD-time .git-meta/remote-url
	@echo "org.opencontainers.image.created=\"$(shell git show -s --format=%cI HEAD)\"" > $@
	@echo "org.opencontainers.image.source=\"$(shell git remote get-url origin)\"" >> $@
	@echo "org.opencontainers.image.version=\"$(shell git describe --tags --always)\"" >> $@
	@echo "org.opencontainers.image.revision=\"$(shell cat .git-meta/HEAD)\"" >> $@
	@echo "org.opencontainers.image.ref.name=\"$(shell git rev-parse --abbrev-ref HEAD)\"" >> $@
	@echo "org.opencontainers.image.title=\"$(shell cat .git-meta/repo-name)\"" >> $@
	@echo "org.opencontainers.image.url=\"$(shell cat .git-remote-url)\"" >> $@

docker/.local-args: docker/.labels
	@echo "--build-arg RUNTIME_IMAGE=$(call read_config,build.cfg,docker_runtime_image)" > $@
	@echo --platform $(platforms) >> $@
	@echo -f docker/Dockerfile >> $@
	@sed 's/^/--label /' docker/.labels >> $@
	@echo "../" >> $@

docker/.local-tags: .git-meta/HEAD
	./scripts/resolve-docker-tags.sh \
		--commit-sha $(shell cat .git-meta/HEAD) \
		--git-ref $(shell git rev-parse --abbrev-ref HEAD) \
		--stable-branches $(call read_config,build.cfg,stable_branches) \
		> $@ 

# Build local image
# Target must be in form: docker/.local-<binary>-image
# Example: make docker/.local-server-image
docker/.local-%-image: docker/.local-args
	$(eval registry := $(call read_config,build.cfg,docker_local_registry))
	$(eval image := $(registry)/$(shell cat .git-meta/repo-name)-$*:latest)
	@echo "Building $(image)"
	docker build -t $(image) \
			--build-arg TARGET_BINARY="$*" \
			$(shell cat $<)
	docker inspect $(image) > $@-inspect
	cat $@-inspect | jq -r '.[0].RepoTags[] | select(. | startswith("$(registry)"))' > $@

# Build all local images for each binary in dist/
# Fill use top folders from cmd folder to understand which binaries to build (wildcard)
.PHONY: docker/local-images
docker/local-images: $(patsubst %, docker/.local-%-image, $(notdir $(wildcard ../cmd/*)))
	@echo "Local images are ready: $?"

# Push image to remote registry
# Target must be in form: docker/.remote-<registry>-<binary>-image
# Example: make docker/.remote-ghcr-server-image
# The registry host is read from build.cfg under docker_<registry>_registry_host key
docker/.remote-%-image: docker/.local-$$(word 2, $$(subst -, ,$$*))-image docker/.remote-tags
	$(eval registry_name := $(word 1,$(subst -, ,$*)))
	$(eval target_binary := $(word 2,$(subst -, ,$*)))
	$(eval registry := $(call read_config,build.cfg,docker_$(registry_name)_registry_host))
	$(eval local_image_id := $(call read_config,$<,id))
	$(eval local_image_tag := $(call read_config,$<,repo_tag))
	$(eval local_image_digest := $(call read_config,$<,repo_digest))
	$(eval remote_image_base_name := $(registry)/$(app_name)-$(target_binary))
	$(eval remote_tags := $(shell cat docker/.remote-tags))
	@echo "Pushing $(local_image_tag) image to $(remote_image_base_name) with tags: $(remote_tags)"
	$(foreach tag,$(remote_tags),\
			$(shell docker tag $(local_image_digest) $(remote_image_base_name):$(tag)) \
			$(shell docker push $(remote_image_base_name):$(tag)) \
	)
	@echo "id=$(local_image_id)" > $@
	@echo "tags=$(remote_tags)" >> $@

# Push all images to remote registry
.PHONY: docker/remote-images
docker/remote-images: $(eval docker_push_registries := $(call read_config,build.cfg,docker_push_registries))
docker/remote-images: $(eval binaries := $(notdir $(wildcard ../cmd/*)))
docker/remote-images: $(patsubst %, docker/.remote-%-image, $(foreach registry,$(subst $(comma), ,$(docker_push_registries)),$(foreach binary,$(binaries),$(registry)-$(binary))))
		@echo "Remote images pushed: $?"

.PHONY: clean docker/clean dist/clean

docker/clean:
	rm -f docker/.local-* docker/.remote-* docker/.labels

dist/clean:
	rm -rf dist

clean: docker/clean dist/clean
	rm -rf .git-meta