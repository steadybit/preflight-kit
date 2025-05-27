# You may need to install dependencies `pip install deepdiff pyyaml`

test-local:
	curl -L https://platform.steadybit.com/api/spec/platform -o platform-spec.json
	yq -p json -o yaml -P platform-spec.json > platform-spec.yaml
	./filter_spec.sh platform-spec.yaml filtered_platform-spec.yaml
	./filter_spec.sh ./../../openapi/spec.yml filtered_preflight-kit-spec.yaml
	python ./diff.py --old filtered_preflight-kit-spec.yaml --new filtered_platform-spec.yaml --key components.schemas --exclude ./api-spec-ignores/exclude.txt --fail-if-diff; \
	STATUS=$$?; \
	rm -f platform-spec.json platform-spec.yaml filtered_platform-spec.yaml filtered_preflight-kit-spec.yaml; \
	exit $$STATUS