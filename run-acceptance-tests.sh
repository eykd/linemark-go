#!/usr/bin/env bash
set -euo pipefail
rm -rf generated-acceptance-tests/ acceptance-pipeline/ir/
go run ./acceptance/cmd/pipeline -action=run
