##
## to run this locally you need to have yq installed
##
## known, acceptable divergences from the platform spec are ignored via the
## WHITELIST and PATH_IGNORES lists in filter_spec.sh
##
PLATFORM_SPEC_URL_PROD ?= https://platform.steadybit.com/api/spec/platform
PLATFORM_SPEC_URL_DEV ?= https://platform.dev.steadybit.com/api/spec/platform

diff-platform-preflight-spec:
	curl -sL $(PLATFORM_SPEC_URL_PROD) -o - | yq -p json -o yaml > platform-spec.yaml
	./filter_spec.sh platform-spec.yaml filtered_platform-spec.yaml
	./filter_spec.sh ../spec.yml filtered_preflight-kit-spec.yaml
	# Show unified diff and fail if different
	if ! diff -u filtered_preflight-kit-spec.yaml filtered_platform-spec.yaml >&2; then \
		echo "ERROR: filtered_preflight-kit-spec.yaml and filtered_platform-spec.yaml differ"; \
		rm -f platform-spec.json platform-spec.yaml filtered_platform-spec.yaml filtered_preflight-kit-spec.yaml; \
		exit 1; \
	fi
	rm -f platform-spec.json platform-spec.yaml filtered_platform-spec.yaml filtered_preflight-kit-spec.yaml
