.PHONY: FORCE
.SECONDEXPANSION:

app_name:=$(shell basename `git rev-parse --show-toplevel`)

# Define a function to read values from any config file
# Usage: $(call read_config,config_file,key_name)
define read_config
$(shell grep -m 1 "^$(2)\s*=\s*" $(1) | cut -d'=' -f2)
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

docker_build_args=\
				--platform $(platforms) \
				--build-arg RUNTIME_IMAGE=$(docker_runtime_image) \
				-f docker/Dockerfile \
				--label org.opencontainers.image.created=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
				--label org.opencontainers.image.source=$(shell git remote get-url origin) \
				--label org.opencontainers.image.version=$(shell git describe --tags --always) \
				--label org.opencontainers.image.revision=$(shell git rev-parse HEAD) \
				--label org.opencontainers.image.ref.name=$(shell git rev-parse --abbrev-ref HEAD) \
				--label org.opencontainers.image.title=$(app_name) \
				--label org.opencontainers.image.url=$(shell git remote get-url origin) \
				../

.git-HEAD: FORCE
	@git rev-parse HEAD > $@.tmp
	@if cmp -s $@ $@.tmp; then \
		rm $@.tmp; \
	else \
		mv $@.tmp $@; \
	fi

# Build local image
# Target must be in form: docker/.local-<binary>-image
# Example: make docker/.local-server-image
docker/.local-%-image:
	$(eval registry := $(call read_config,build.cfg,docker_local_registry_host))
	$(eval image := $(registry)/$(app_name)-$*:latest)
	@echo "Building $(image)"
	docker build -t $(image) \
			--build-arg TARGET_BINARY="$*" \
			$(docker_build_args)
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
		rm -f docker/.local-* docker/.remote-*

dist/clean:
		rm -rf dist

clean: docker/clean dist/clean
		rm -f .git-HEAD