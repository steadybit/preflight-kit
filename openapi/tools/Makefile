##
## to run this locally you need to have yq and python installed
## also you need to install the pyyaml and deepdiff python packages:
##  pip3 install pyyaml deepdiff
##
diff-platform-preflight-spec:
	curl -sL https://platform.steadybit.com/api/spec/platform -o - | yq -p json -o yaml > platform-spec.yaml
	./filter_spec.sh platform-spec.yaml filtered_platform-spec.yaml
	./filter_spec.sh ../spec.yml filtered_preflight-kit-spec.yaml
	./diff.py --old filtered_preflight-kit-spec.yaml --new filtered_platform-spec.yaml --key components.schemas --exclude ./api-spec-ignores.txt --fail-if-diff; \
	STATUS=$$?; \
	rm -f platform-spec.json platform-spec.yaml filtered_platform-spec.yaml filtered_preflight-kit-spec.yaml; \
	exit $$STATUS;