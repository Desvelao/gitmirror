#!/usr/bin/env bash

set -euo pipefail

echo "Formatting Go files with gofmt..."
while IFS= read -r -d '' file; do
    echo "Formatting $file..."
	gofmt -w "$file"
done < <(find . -name "*.go" -not -path "./vendor/*" -print0)

echo "Formatting README.md..."
if command -v prettier >/dev/null 2>&1; then
	if ! prettier --write README.md; then
		echo "Skipping README formatting: 'prettier' command failed to execute."
	fi
elif command -v npx >/dev/null 2>&1; then
	node_major=""
	if command -v node >/dev/null 2>&1; then
		node_major="$(node -p "process.versions.node.split('.')[0]" 2>/dev/null || true)"
	fi

	if [[ -n "$node_major" && "$node_major" -ge 14 ]]; then
		if ! npx --yes --package prettier@3.3.3 prettier --write README.md; then
			echo "Skipping README formatting: npx prettier@3.3.3 failed."
		fi
	else
		if ! npx --yes --package prettier@2.8.8 prettier --write README.md; then
			echo "Skipping README formatting: npx prettier@2.8.8 failed."
		fi
	fi
else
	echo "Skipping README formatting: install prettier or npx to format Markdown."
fi

echo "Formatting completed."
