#!/usr/bin/env python3
import argparse
import sys
import json
import yaml
from deepdiff import DeepDiff

def load_yaml(filename):
    with open(filename, 'r') as f:
        return yaml.safe_load(f)

def remove_whitelisted(data, whitelist):
    """
    Remove keys present in the whitelist from a dictionary.
    This function removes any top-level key found in the whitelist.
    """
    if isinstance(data, dict):
        for key in whitelist:
            if key in data:
                del data[key]
    return data

def get_nested(data, key_path):
    """
    Retrieve a nested value in a dict given a dotted key path.
    For example, key_path="components.schemas" returns data["components"]["schemas"].
    If any key is missing, returns an empty dict.
    """
    keys = key_path.split('.')
    for key in keys:
        if not isinstance(data, dict) or key not in data:
            return {}
        data = data[key]
    return data

def main():
    parser = argparse.ArgumentParser(
        description="Compare a portion of two OpenAPI specs using DeepDiff."
    )
    parser.add_argument("--old", required=True, help="Path to the old spec YAML file.")
    parser.add_argument("--new", required=True, help="Path to the new spec YAML file.")
    parser.add_argument("--key", default="components.schemas",
                        help="Dotted path in the YAML to compare (default: components.schemas)")
    parser.add_argument("--whitelist", help="Path to a whitelist text file (one key per line) to ignore in the diff.")
    parser.add_argument("--exclude", help="Path to a text file with DeepDiff exclude paths (one per line).")
    parser.add_argument("--fail-if-diff", action="store_true",
                        help="Exit with a non-zero status if differences are found.")
    args = parser.parse_args()

    # Load the old and new specs
    old_spec = load_yaml(args.old)
    new_spec = load_yaml(args.new)

    # Extract the target nested data using the provided key path.
    old_data = get_nested(old_spec, args.key)
    new_data = get_nested(new_spec, args.key)

    # Apply whitelist removal if provided.
    if args.whitelist:
        try:
            with open(args.whitelist, 'r') as f:
                whitelist = [line.strip() for line in f if line.strip()]
            old_data = remove_whitelisted(old_data, whitelist)
            new_data = remove_whitelisted(new_data, whitelist)
        except Exception as e:
            print(f"Error loading whitelist file: {e}")
            sys.exit(1)

    # Load exclude paths for DeepDiff.
    exclude_paths = []
    if args.exclude:
        try:
            with open(args.exclude, 'r') as f:
                exclude_paths = [line.strip() for line in f if line.strip()]
        except Exception as e:
            print(f"Error loading exclude file: {e}")
            sys.exit(1)

    # Compare using DeepDiff and pass the exclude_paths.
    diff = DeepDiff(old_data, new_data, ignore_order=True, exclude_paths=exclude_paths)

    if diff:
        print("Differences found:")
        print(json.dumps(diff, indent=2, default=str))
        if args.fail_if_diff:
            sys.exit(1)
    else:
        print("No differences found.")

if __name__ == "__main__":
    main()